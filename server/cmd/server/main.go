package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"

	"github.com/Franck1120/physicscopilot/server/internal/handlers"
	"github.com/Franck1120/physicscopilot/server/internal/services"
)

func main() {
	// Initialize services
	sessionSvc := services.NewSessionService()
	geminiSvc, err := services.NewGeminiService()
	if err != nil {
		log.Fatal("Gemini service init failed:", err)
	}
	convSvc := services.NewConversationService(sessionSvc, geminiSvc)
	wsHandler := handlers.NewWSHandler(convSvc, sessionSvc)

	// Background cleanup of expired sessions every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			sessionSvc.CleanupExpiredSessions(30 * time.Minute)
		}
	}()

	app := fiber.New(fiber.Config{
		AppName: "PhysicsCopilot Server v0.1.0",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "physicscopilot",
		})
	})

	// WebSocket upgrade middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
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

	log.Fatal(app.Listen(":" + port))
}
