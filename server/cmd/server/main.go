package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
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

	// Warn loudly in production when SUPABASE_JWT_SECRET is absent — the
	// WebSocket auth middleware falls back to dev mode (no authentication).
	if os.Getenv("APP_ENV") == "production" && os.Getenv("SUPABASE_JWT_SECRET") == "" {
		slog.Warn("SUPABASE_JWT_SECRET is not set — WebSocket connections are not authenticated; all clients can connect without a valid JWT")
	}

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
	feedbackHandler := handlers.NewFeedbackHandler(dbSvc)

	// Background memory metrics collection every 30 seconds.
	// Warns at slog.Warn level when heap usage exceeds 80 % of GOMEMLIMIT.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				metrics.MemHeapAllocBytes.Set(float64(m.HeapAlloc))
				metrics.MemSysBytes.Set(float64(m.Sys))
				metrics.MemNumGCTotal.Set(float64(m.NumGC))

				// GOMEMLIMIT=-1 reads current limit without changing it.
				// Default is math.MaxInt64 (no limit set).
				limit := debug.SetMemoryLimit(-1)
				if limit > 0 && limit != math.MaxInt64 {
					usagePct := float64(m.HeapAlloc) / float64(limit) * 100
					if usagePct > 80 {
						slog.Warn("high memory usage",
							"heap_alloc_mb", m.HeapAlloc/1024/1024,
							"limit_mb", uint64(limit)/1024/1024,
							"usage_pct", int(usagePct),
						)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Background cleanup of expired sessions every 5 minutes.
	// CleanupExpiredSessions removes them from RAM and marks them 'expired' in DB.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if n := sessionSvc.CleanupExpiredSessions(30 * time.Minute); n > 0 {
				slog.Info("expired sessions cleaned up", "count", n)
			}
		}
	}()

	// ── HTTP app ─────────────────────────────────────────────────────────────
	app := newFiberApp(version, sessionHandler, feedbackHandler, wsHandler, dbSvc)

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
	feedbackHandler *handlers.FeedbackHandler,
	wsHandler *handlers.WSHandler,
	db handlers.DBPinger, // nil when DATABASE_URL not set
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "PhysicsCopilot Server v" + ver,
		// Reject request bodies larger than 1 MB to prevent memory exhaustion.
		BodyLimit: 1 * 1024 * 1024,
		// IdleTimeout is the maximum amount of time to wait for the next request
		// on a keep-alive connection. 60 s is a reasonable default that balances
		// resource usage against client reconnect overhead.
		IdleTimeout:  60 * time.Second,
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

	// Security headers — applied to every response to harden the HTTP surface.
	// CSRF protection is provided by the JWT Authorization-header auth scheme
	// (stateless, not cookie-based), so dedicated CSRF tokens are not required.
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'none'")
		return c.Next()
	})

	// Compress REST responses with gzip (level: best speed to favour latency).
	// WebSocket upgrade requests are excluded automatically — the compress
	// middleware does not touch hijacked connections.
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
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
	api.Post("/feedback", feedbackHandler.Submit)

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
