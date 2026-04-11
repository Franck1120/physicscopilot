package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
)

func main() {
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
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// TODO: handle incoming camera frames, call Gemini, stream instructions back
		log.Println("WebSocket client connected:", c.RemoteAddr())
		defer func() {
			log.Println("WebSocket client disconnected:", c.RemoteAddr())
			c.Close()
		}()

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			// Echo for now — replace with AI pipeline
			if err := c.WriteMessage(mt, msg); err != nil {
				break
			}
		}
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
