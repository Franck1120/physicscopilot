package handlers

import (
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/fiber/v2"
)

// WSConnCounter is satisfied by any type that can report the number of currently
// open WebSocket connections. *WSHandler implements this interface.
type WSConnCounter interface {
	ActiveConnections() int32
}

// RAGLoader is satisfied by any type that reports whether the knowledge base
// is loaded and how many entries it contains. *services.RAGService implements
// this interface.
type RAGLoader interface {
	Loaded() bool
	EntryCount() int
}

// StatsHandler exposes GET /api/stats.
// It aggregates in-memory session counts, WebSocket connection counts,
// knowledge-base status, server uptime, and build version into a single JSON
// payload suitable for dashboards and monitoring tools.
type StatsHandler struct {
	sessions  *services.SessionService
	ws        WSConnCounter
	rag       RAGLoader
	version   string
	startTime time.Time
}

// NewStatsHandler returns a StatsHandler with only the session service wired in.
// ws and rag default to nil; their corresponding response fields will be zero-valued.
// Use NewStatsHandlerFull to wire all optional dependencies.
func NewStatsHandler(sessions *services.SessionService) *StatsHandler {
	return &StatsHandler{sessions: sessions, startTime: time.Now()}
}

// NewStatsHandlerFull returns a StatsHandler with all dependencies set.
// ws may be nil if no WebSocket handler is running.
// rag may be nil if the knowledge base is not configured.
// version is the build version string baked in at startup.
// startTime is used to compute uptime_seconds; callers typically pass time.Now()
// at server start.
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
//
// Aggregates the following data points into a single JSON response:
//   - active_sessions:       current in-memory session count
//   - active_ws_connections: live WebSocket connections (0 when ws is nil)
//   - total_sessions_started: cumulative counter (populated by metrics layer)
//   - kb_loaded / kb_entry_count: knowledge-base state (0/false when rag is nil)
//   - uptime_seconds:        seconds since h.startTime
//   - version:               build version string
//   - total_messages:        sum of ConversationHistory lengths across all sessions
//
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
