package handlers

import (
	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/fiber/v2"
)

// StatsHandler exposes GET /api/stats.
type StatsHandler struct {
	sessions *services.SessionService
}

// NewStatsHandler returns a StatsHandler wired to the given session store.
func NewStatsHandler(sessions *services.SessionService) *StatsHandler {
	return &StatsHandler{sessions: sessions}
}

// statsResponse is the payload for GET /api/stats.
type statsResponse struct {
	ActiveSessions int `json:"active_sessions"`
	TotalMessages  int `json:"total_messages"`
}

// GetStats handles GET /api/stats.
// Response 200: {"active_sessions": N, "total_messages": M}
func (h *StatsHandler) GetStats(c *fiber.Ctx) error {
	all := h.sessions.ListSessions()
	totalMsgs := 0
	for _, s := range all {
		totalMsgs += len(s.ConversationHistory)
	}
	return c.JSON(statsResponse{
		ActiveSessions: len(all),
		TotalMessages:  totalMsgs,
	})
}
