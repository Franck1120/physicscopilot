package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Franck1120/physicscopilot/server/internal/handlers"
	applogger "github.com/Franck1120/physicscopilot/server/internal/logger"
	"github.com/Franck1120/physicscopilot/server/internal/metrics"
	"github.com/Franck1120/physicscopilot/server/internal/middleware"
	"github.com/Franck1120/physicscopilot/server/internal/services"
)

const version = "0.1.0"

var startTime = time.Now()

func main() {
	applogger.Init()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ── Services ────────────────────────────────────────────────────────────
	sessionSvc := services.NewSessionService()

	aiBackend, err := services.NewAIBackend()
	if err != nil {
		slog.Error("AI backend init failed", "err", err)
		os.Exit(1)
	}
	ragSvc, err := services.NewRAGService()
	if err != nil {
		slog.Error("KB init failed", "err", err)
		os.Exit(1)
	}
	if !ragSvc.Loaded() {
		slog.Warn("knowledge base not loaded — KB_PATH absent or file missing; running without KB context")
	}

	// ── Optional Postgres backend ────────────────────────────────────────────
	var dbSvc *services.DBService
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		svc, err := services.NewDBService(ctx, dbURL)
		if err != nil {
			slog.Warn("DB init failed — running without persistence", "err", err)
		} else {
			dbSvc = svc
			sessionSvc.SetDB(dbSvc)
			if err := sessionSvc.HydrateFromDB(ctx); err != nil {
				slog.Warn("failed to hydrate sessions from DB", "err", err)
			}
		}
	}

	convSvc := services.NewConversationService(sessionSvc, aiBackend, ragSvc)
	wsHandler := handlers.NewWSHandler(convSvc, sessionSvc)
	sessionHandler := handlers.NewSessionHandler(sessionSvc)

	// Background cleanup of expired sessions every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			sessionSvc.CleanupExpiredSessions(30 * time.Minute)
		}
	}()

	// ── HTTP app ─────────────────────────────────────────────────────────────
	app := newFiberApp(version, sessionHandler, wsHandler, dbSvc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		slog.Info("server starting", "port", port, "version", version)
		if err := app.Listen(":" + port); err != nil {
			slog.Error("server stopped", "err", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received — closing connections")

	wsHandler.CloseAll()
	// Brief grace period for clients to acknowledge the close frame
	time.Sleep(500 * time.Millisecond)

	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		slog.Error("shutdown error", "err", err)
		os.Exit(1)
	}
	if dbSvc != nil {
		dbSvc.Close()
	}
	slog.Info("server stopped cleanly")
}

// newFiberApp builds and returns the configured Fiber application.
// Extracted from main() so tests can construct the app without starting a
// listener or requiring env vars beyond the test's control.
func newFiberApp(
	ver string,
	sessionHandler *handlers.SessionHandler,
	wsHandler *handlers.WSHandler,
	db handlers.DBPinger, // nil when DATABASE_URL not set
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:     "PhysicsCopilot Server v" + ver,
		// Reject request bodies larger than 1 MB to prevent memory exhaustion.
		BodyLimit:   1 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			slog.Error("request error", "path", c.Path(), "method", c.Method(), "status", code, "err", err)
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			slog.Error("panic recovered",
				"error", fmt.Sprintf("%v", e),
				"path", c.Path(),
				"method", c.Method(),
			)
		},
	}))

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		if os.Getenv("APP_ENV") == "production" {
			slog.Warn("ALLOWED_ORIGINS is not set — cross-origin requests will be blocked")
		} else {
			allowedOrigins = "*"
		}
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: "GET,POST,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	app.Use(middleware.StructuredLogger())

	// Request timeout — abort handlers that take longer than 30 s.
	app.Use(func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
		defer cancel()
		c.SetUserContext(ctx)
		return c.Next()
	})

	// HSTS — only in production to avoid breaking local HTTP dev.
	if os.Getenv("APP_ENV") == "production" {
		app.Use(func(c *fiber.Ctx) error {
			c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			return c.Next()
		})
	}

	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/metrics" {
			return c.Next()
		}
		start := time.Now()
		err := c.Next()
		status := strconv.Itoa(c.Response().StatusCode())
		dur := time.Since(start).Seconds()
		metrics.HttpRequestsTotal.WithLabelValues(c.Method(), c.Path(), status).Inc()
		metrics.HttpRequestDuration.WithLabelValues(c.Method(), c.Path()).Observe(dur)
		return err
	})

	apiLimiter := middleware.NewIPRateLimiter()
	app.Use("/health", apiLimiter.Middleware())

	app.Get("/health", handlers.NewHealthHandler(ver, startTime, wsHandler, db))

	api := app.Group("/api", apiLimiter.Middleware(), handlers.WSAuthMiddleware())
	api.Get("/docs", handlers.OpenAPIHandler())
	api.Post("/sessions", sessionHandler.CreateSession)
	api.Get("/sessions", sessionHandler.ListSessions)
	api.Get("/sessions/:id", sessionHandler.GetSession)
	api.Delete("/sessions/:id", sessionHandler.DeleteSession)

	app.Get("/metrics", middleware.MetricsBasicAuth(), adaptor.HTTPHandler(promhttp.Handler()))

	app.Use("/ws", handlers.WSAuthMiddleware(), func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", websocket.New(wsHandler.Handle))

	return app
}
