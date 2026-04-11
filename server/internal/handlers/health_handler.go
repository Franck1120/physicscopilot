package handlers

import (
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthResponse is the JSON payload returned by GET /health.
type HealthResponse struct {
	Status            string `json:"status"`
	Service           string `json:"service"`
	Version           string `json:"version"`
	Uptime            string `json:"uptime"`
	ActiveConnections int32  `json:"active_connections"`
	MemoryMB          uint64 `json:"memory_mb"`
}

// NewHealthHandler returns a Fiber handler for GET /health.
// It reports build version, uptime since startTime, current open WebSocket
// connections, and current heap allocation in MB.
func NewHealthHandler(version string, startTime time.Time, ws *WSHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return c.JSON(HealthResponse{
			Status:            "ok",
			Service:           "physicscopilot",
			Version:           version,
			Uptime:            time.Since(startTime).Round(time.Second).String(),
			ActiveConnections: ws.ActiveConnections(),
			MemoryMB:          m.Alloc / 1024 / 1024,
		})
	}
}
