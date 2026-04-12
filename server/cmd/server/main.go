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

	// Initialize services
	sessionSvc := services.NewSessionService()
	geminiSvc, err := services.NewGeminiService()
	if err != nil {
		slog.Error("Gemini service init failed", "err", err)
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
	convSvc := services.NewConversationService(sessionSvc, geminiSvc, ragSvc)
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

	app := fiber.New(fiber.Config{
		AppName: "PhysicsCopilot Server v" + version,
		// JSON error responses for all errors (including panics after recover)
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			slog.Error("request error", "path", c.Path(), "method", c.Method(), "status", code, "err", err)
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Recovery — catch panics, log with slog, return JSON 500
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

	// CORS — origins controlled by ALLOWED_ORIGINS env var.
	// Development default: "*" (permissive).
	// Production: set ALLOWED_ORIGINS to a comma-separated list of exact origins,
	// e.g. "https://yourapp.com,https://www.yourapp.com".
	// If ALLOWED_ORIGINS is unset in production the empty string is forwarded to
	// the Fiber CORS middleware, which will block all cross-origin requests.
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

	// Structured request logger — replaces the default Fiber text logger.
	// Emits JSON in production (APP_ENV=production), text otherwise.
	// Each request gets a random request_id for correlation.
	app.Use(middleware.StructuredLogger())

	// Prometheus metrics middleware — records request count and latency for
	// every non-/metrics route (recording /metrics itself would be noisy).
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

	// REST API rate limiting — 60 req/min per IP
	apiLimiter := middleware.NewIPRateLimiter()
	app.Use("/health", apiLimiter.Middleware())

	// Health check — version, uptime, active connections, memory
	app.Get("/health", handlers.NewHealthHandler(version, startTime, wsHandler))

	// Session REST API — rate-limited + JWT auth (no-op when SUPABASE_JWT_SECRET is unset).
	// In production the client must send: Authorization: Bearer <supabase-jwt>
	api := app.Group("/api", apiLimiter.Middleware(), handlers.WSAuthMiddleware())
	api.Post("/sessions", sessionHandler.CreateSession)
	api.Get("/sessions", sessionHandler.ListSessions)
	api.Get("/sessions/:id", sessionHandler.GetSession)
	api.Delete("/sessions/:id", sessionHandler.DeleteSession)

	// Prometheus metrics endpoint — protected by HTTP Basic Auth.
	// Credentials: user=admin, password=$METRICS_PASSWORD.
	// Returns 503 if METRICS_PASSWORD is not set (endpoint disabled).
	app.Get("/metrics", middleware.MetricsBasicAuth(), adaptor.HTTPHandler(promhttp.Handler()))

	// WebSocket: JWT auth → upgrade guard → handler
	// WSAuthMiddleware is a no-op when SUPABASE_JWT_SECRET is unset (dev mode).
	app.Use("/ws", handlers.WSAuthMiddleware(), func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket endpoint — real-time repair guidance session
	app.Get("/ws", websocket.New(wsHandler.Handle))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown on SIGINT / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	slog.Info("server stopped cleanly")
}
