package handlers

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/websocket/v2"
)

// maxFramesPerSecond is the maximum number of camera frames a single
// WebSocket connection may send within a one-second sliding window.
const maxFramesPerSecond = 5

// maxConnsPerIP is the maximum number of concurrent WebSocket connections
// allowed from a single IP address.
const maxConnsPerIP = 10

// IncomingMessage represents a JSON message received from the client.
type IncomingMessage struct {
	Type      string `json:"type"`      // "frame" | "text" | "ping"
	Data      string `json:"data"`      // base64 image (only for type="frame")
	Content   string `json:"content"`   // user text (only for type="text")
	Timestamp int64  `json:"timestamp"` // ms epoch
}

// OutgoingMessage represents a JSON message sent to the client.
type OutgoingMessage struct {
	Type    string               `json:"type"`            // "response" | "error" | "pong"
	Text    string               `json:"text,omitempty"`
	Overlay services.OverlayData `json:"overlay,omitempty"`
	Step    services.StepInfo    `json:"step,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// ipConnTracker enforces a per-IP limit on concurrent WebSocket connections.
type ipConnTracker struct {
	mu    sync.Mutex
	conns map[string]int
}

func newIPConnTracker() *ipConnTracker {
	return &ipConnTracker{conns: make(map[string]int)}
}

// add increments the connection count for ip and returns true if within
// the limit. Returns false if the limit is already reached.
func (t *ipConnTracker) add(ip string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conns[ip] >= maxConnsPerIP {
		return false
	}
	t.conns[ip]++
	return true
}

// remove decrements the connection count for ip.
func (t *ipConnTracker) remove(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conns[ip] > 0 {
		t.conns[ip]--
	}
}

// WSHandler manages WebSocket connections by coordinating session
// lifecycle and conversation processing through the service layer.
type WSHandler struct {
	conversations *services.ConversationService
	sessions      *services.SessionService
	activeConns   atomic.Int32
	ipConns       *ipConnTracker

	connsMu sync.Mutex
	conns   map[*websocket.Conn]struct{}
}

// NewWSHandler creates a WSHandler wired to the given conversation
// and session services.
func NewWSHandler(conversations *services.ConversationService, sessions *services.SessionService) *WSHandler {
	return &WSHandler{
		conversations: conversations,
		sessions:      sessions,
		ipConns:       newIPConnTracker(),
		conns:         make(map[*websocket.Conn]struct{}),
	}
}

// ActiveConnections returns the current number of open WebSocket connections.
func (h *WSHandler) ActiveConnections() int32 {
	return h.activeConns.Load()
}

// CloseAll sends a close frame to all active connections.
// Used during graceful shutdown.
func (h *WSHandler) CloseAll() {
	h.connsMu.Lock()
	defer h.connsMu.Unlock()
	for c := range h.conns {
		_ = c.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
		)
		_ = c.Close()
	}
	slog.Info("closed all WebSocket connections", "count", len(h.conns))
}

func (h *WSHandler) register(c *websocket.Conn) {
	h.connsMu.Lock()
	h.conns[c] = struct{}{}
	h.connsMu.Unlock()
}

func (h *WSHandler) deregister(c *websocket.Conn) {
	h.connsMu.Lock()
	delete(h.conns, c)
	h.connsMu.Unlock()
}

// frameRateLimiter enforces a per-connection limit on camera frames
// using a fixed one-second window counter.
type frameRateLimiter struct {
	count       int
	windowStart time.Time
}

// allow returns true if the frame is within the rate limit, and
// increments the counter. Resets the window when a new second begins.
func (r *frameRateLimiter) allow() bool {
	now := time.Now()
	if now.Sub(r.windowStart) >= time.Second {
		r.count = 0
		r.windowStart = now
	}
	if r.count >= maxFramesPerSecond {
		return false
	}
	r.count++
	return true
}

// Handle is the WebSocket connection handler. It creates a session,
// reads incoming JSON messages in a loop, dispatches them to the
// appropriate service method, and writes responses back to the client.
func (h *WSHandler) Handle(c *websocket.Conn) {
	// Extract IP (RemoteAddr format is "IP:port")
	ip, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		ip = c.RemoteAddr().String()
	}

	if !h.ipConns.add(ip) {
		slog.Warn("WebSocket connection limit reached for IP", "ip", ip, "limit", maxConnsPerIP)
		_ = c.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "connection limit reached"),
		)
		return
	}

	h.activeConns.Add(1)
	h.register(c)

	slog.Info("WebSocket client connected",
		"remote_addr", c.RemoteAddr(),
		"active_conns", h.activeConns.Load(),
	)

	session, err := h.sessions.CreateSession("unknown", "unknown")
	if err != nil {
		slog.Error("failed to create session", "err", err, "remote_addr", c.RemoteAddr())
		h.activeConns.Add(-1)
		h.ipConns.remove(ip)
		h.deregister(c)
		return
	}
	sessionID := session.SessionID

	defer func() {
		if delErr := h.sessions.DeleteSession(sessionID); delErr != nil {
			slog.Warn("failed to delete session", "session_id", sessionID, "err", delErr)
		}
		h.activeConns.Add(-1)
		h.ipConns.remove(ip)
		h.deregister(c)
		slog.Info("WebSocket client disconnected",
			"remote_addr", c.RemoteAddr(),
			"session_id", sessionID,
			"active_conns", h.activeConns.Load(),
		)
	}()

	rateLimiter := &frameRateLimiter{windowStart: time.Now()}

	for {
		var msg IncomingMessage
		if err := c.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("unexpected WebSocket close", "err", err, "session_id", sessionID)
			}
			break
		}

		switch msg.Type {
		case "frame":
			h.handleFrame(c, sessionID, msg, rateLimiter)
		case "text":
			h.handleText(c, sessionID, msg)
		case "ping":
			h.handlePing(c)
		default:
			slog.Warn("unknown message type",
				"type", msg.Type,
				"remote_addr", c.RemoteAddr(),
				"session_id", sessionID,
			)
		}
	}
}

// handleFrame processes a camera frame through the conversation service.
// Frames that exceed the rate limit are silently dropped.
func (h *WSHandler) handleFrame(c *websocket.Conn, sessionID string, msg IncomingMessage, rl *frameRateLimiter) {
	if !rl.allow() {
		return
	}

	ctx := context.Background()
	result, err := h.conversations.ProcessFrame(ctx, sessionID, msg.Data, "")
	if err != nil {
		slog.Error("ProcessFrame error", "session_id", sessionID, "err", err)
		writeError(c, err.Error())
		return
	}

	// nil result means duplicate frame was skipped
	if result == nil {
		return
	}

	writeResponse(c, result)
}

// handleText processes a text-only conversation turn.
func (h *WSHandler) handleText(c *websocket.Conn, sessionID string, msg IncomingMessage) {
	ctx := context.Background()
	result, err := h.conversations.ProcessTextMessage(ctx, sessionID, msg.Content)
	if err != nil {
		slog.Error("ProcessTextMessage error", "session_id", sessionID, "err", err)
		writeError(c, err.Error())
		return
	}

	if result == nil {
		return
	}

	writeResponse(c, result)
}

// handlePing responds to a client ping with a pong message.
func (h *WSHandler) handlePing(c *websocket.Conn) {
	out := OutgoingMessage{Type: "pong"}
	if err := c.WriteJSON(out); err != nil {
		slog.Warn("failed to write pong", "err", err)
	}
}

// writeResponse sends a successful analysis result to the client.
func writeResponse(c *websocket.Conn, result *services.ProcessResult) {
	out := OutgoingMessage{
		Type:    "response",
		Text:    result.Text,
		Overlay: result.Overlay,
		Step:    result.Step,
	}
	if err := c.WriteJSON(out); err != nil {
		slog.Warn("failed to write response", "err", err)
	}
}

// writeError sends an error message to the client without closing
// the connection.
func writeError(c *websocket.Conn, errMsg string) {
	out := OutgoingMessage{
		Type:  "error",
		Error: errMsg,
	}
	if err := c.WriteJSON(out); err != nil {
		slog.Warn("failed to write error", "err", err)
	}
}
