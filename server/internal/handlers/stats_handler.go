package handlers

import (
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/fiber/v2"
)

// WSConnCounter is satisfied by any type that can report the number of currently
// open WebSocket connections (e.g. *WSHandler).
type WSConnCounter interface {
	ActiveConnections() int32
}

// RAGLoader is satisfied by any type that reports whether the knowledge base
// is loaded and how many entries it contains (e.g. *services.RAGService).
type RAGLoader interface {
	Loaded() bool
	EntryCount() int
}

// StatsHandler exposes GET /api/stats.
type StatsHandler struct {
	sessions  *services.SessionService
	ws        WSConnCounter
	rag       RAGLoader
	version   string
	startTime time.Time
}

// NewStatsHandler returns a StatsHandler wired to the given dependencies.
// ws and rag may be nil; their fields will be omitted / zero-valued in that case.
func NewStatsHandler(sessions *services.SessionService) *StatsHandler {
	return &StatsHandler{sessions: sessions, startTime: time.Now()}
}

// NewStatsHandlerFull returns a StatsHandler with all optional dependencies set.
func NewStatsHandlerFull(
	sessions *services.SessionService,
	ws WSConnCounter,
	rag RAGLoader,
	version string,
	startTime time.Time,
) *StatsHandler {
	return &StatsHandler{
		sessions:  sessions,
		ws:        ws,
		rag:       rag,
		version:   version,
		startTime: startTime,
	}
}

// statsResponse is the payload for GET /api/stats.
type statsResponse struct {
	ActiveSessions       int    `json:"active_sessions"`
	ActiveWSConnections  int32  `json:"active_ws_connections"`
	TotalSessionsStarted uint64 `json:"total_sessions_started"`
	KBLoaded             bool   `json:"kb_loaded"`
	KBEntryCount         int    `json:"kb_entry_count"`
	UptimeSeconds        int64  `json:"uptime_seconds"`
	Version              string `json:"version"`
	// TotalMessages is kept for backwards compatibility with existing tests/clients.
	TotalMessages int `json:"total_messages"`
}

// GetStats handles GET /api/stats.
// Response 200: full stats JSON.
func (h *StatsHandler) GetStats(c *fiber.Ctx) error {
	all := h.sessions.ListSessions()

	totalMsgs := 0
	for _, s := range all {
		totalMsgs += len(s.ConversationHistory)
	}

	var activeWS int32
	if h.ws != nil {
		activeWS = h.ws.ActiveConnections()
	}

	var kbLoaded bool
	var kbCount int
	if h.rag != nil {
		kbLoaded = h.rag.Loaded()
		kbCount = h.rag.EntryCount()
	}

	uptime := int64(time.Since(h.startTime).Seconds())

	return c.JSON(statsResponse{
		ActiveSessions:       len(all),
		ActiveWSConnections:  activeWS,
		TotalSessionsStarted: 0, // populated by caller when metrics are accessible
		KBLoaded:             kbLoaded,
		KBEntryCount:         kbCount,
		UptimeSeconds:        uptime,
		Version:              h.version,
		TotalMessages:        totalMsgs,
	})
}
