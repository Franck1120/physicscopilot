package handlers

import (
	"context"
	"runtime"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/fiber/v2"
)

// DBPinger is satisfied by any type that can report DB reachability.
// In production this is *services.DBService; in tests a mock or nil.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// PoolStatter extends DBPinger with connection pool statistics.
// *services.DBService satisfies this interface; tests can pass a plain DBPinger.
type PoolStatter interface {
	DBPinger
	PoolStats() services.DBPoolStats
}

// HealthResponse is the JSON payload returned by GET /health.
type HealthResponse struct {
	Status            string               `json:"status"`
	Service           string               `json:"service"`
	Version           string               `json:"version"`
	Uptime            string               `json:"uptime"`
	ActiveConnections int32                `json:"active_connections"`
	MemoryMB          uint64               `json:"memory_mb"`
	DBStatus          string               `json:"db_status"` // "ok" | "unavailable" | "not_configured"
	DBPool            *services.DBPoolStats `json:"db_pool,omitempty"`
}

// NewHealthHandler returns a Fiber handler for GET /health.
// It reports build version, uptime since startTime, current open WebSocket
// connections, heap allocation in MB, and optional Postgres reachability.
// Pass db=nil when DATABASE_URL is not configured.
func NewHealthHandler(version string, startTime time.Time, ws *WSHandler, db DBPinger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		dbStatus := "not_configured"
		if db != nil {
			ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
			defer cancel()
			if err := db.Ping(ctx); err != nil {
				dbStatus = "unavailable"
			} else {
				dbStatus = "ok"
			}
		}

		resp := HealthResponse{
			Status:            "ok",
			Service:           "physicscopilot",
			Version:           version,
			Uptime:            time.Since(startTime).Round(time.Second).String(),
			ActiveConnections: ws.ActiveConnections(),
			MemoryMB:          m.Alloc / 1024 / 1024,
			DBStatus:          dbStatus,
		}
		if ps, ok := db.(PoolStatter); ok && db != nil {
			stats := ps.PoolStats()
			resp.DBPool = &stats
		}
		return c.JSON(resp)
	}
}
