package handlers

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/metrics"
	"github.com/Franck1120/physicscopilot/server/internal/middleware"
	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/websocket/v2"
)

// defaultMaxFPS is the default per-connection frame rate limit.
// Override with the WS_MAX_FPS environment variable (min 1, max 30).
const defaultMaxFPS = 5

// maxConnsPerIP is the maximum number of concurrent WebSocket connections
// allowed from a single IP address.
const maxConnsPerIP = 10

// maxMessageSize is the maximum size of a single WebSocket message (10 MB).
// Frame messages carry base64-encoded JPEG camera frames; 10 MB is generous
// headroom even for high-res images.
const maxMessageSize = 10 << 20 // 10 MB

// Heartbeat timing constants.
const (
	// pingInterval is how often the server sends a WebSocket Ping frame.
	pingInterval = 30 * time.Second
	// pongWait is the read deadline set after each Ping. If no Pong arrives
	// within this window the read returns an error and the connection closes.
	pongWait = 40 * time.Second // pingInterval + 10 s grace
	// writeWait is the deadline for a single control-frame write.
	writeWait = 10 * time.Second
)

// IncomingMessage represents a JSON message received from the client.
type IncomingMessage struct {
	Type      string `json:"type"`      // "frame" | "text" | "ping"
	Data      string `json:"data"`      // base64 image (only for type="frame")
	Content   string `json:"content"`   // user text (only for type="text")
	Timestamp int64  `json:"timestamp"` // ms epoch
}

// OutgoingMessage represents a JSON message sent to the client.
type OutgoingMessage struct {
	Type      string               `json:"type"`                 // "response" | "error" | "pong"
	Text      string               `json:"text,omitempty"`
	VoiceText string               `json:"voice_text,omitempty"` // TTS-optimised instruction (no markdown)
	Overlay   services.OverlayData `json:"overlay,omitempty"`
	Step      services.StepInfo    `json:"step,omitempty"`
	Error     string               `json:"error,omitempty"`
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

// ---------------------------------------------------------------------------
// safeConn — gorilla/websocket is not concurrent-safe for writes.
// safeConn serialises all writes through a mutex so the ping goroutine and
// the main read-loop goroutine can both write without data races.
// ---------------------------------------------------------------------------

type safeConn struct {
	c  *websocket.Conn
	mu sync.Mutex
}

func (s *safeConn) writeJSON(v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.c.WriteJSON(v)
}

func (s *safeConn) writePing() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.c.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait))
}

func (s *safeConn) writeClose(code int, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Errors are intentionally ignored: the connection may already be closing
	// from the peer side, making write failures expected and non-actionable.
	_ = s.c.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, text),
	)
	_ = s.c.Close()
}

// ---------------------------------------------------------------------------
// WSHandler
// ---------------------------------------------------------------------------

// WSHandler manages WebSocket connections by coordinating session
// lifecycle and conversation processing through the service layer.
type WSHandler struct {
	conversations *services.ConversationService
	sessions      *services.SessionService
	activeConns   atomic.Int32
	ipConns       *ipConnTracker
	userRL        *middleware.UserRateLimiter // per-authenticated-user message rate limiter
	maxFPS        int // per-connection frame rate cap (from WS_MAX_FPS or defaultMaxFPS)

	connsMu sync.Mutex
	conns   map[*safeConn]struct{}

	// Heartbeat overrides — zero values fall back to the package-level constants.
	// Set these fields in tests to use short timeouts without mutating globals.
	PingInterval time.Duration
	PongWait     time.Duration
}

// NewWSHandler creates a WSHandler wired to the given conversation
// and session services. The per-connection frame rate limit is read from
// WS_MAX_FPS (default: 5; clamped to [1, 30]).
func NewWSHandler(conversations *services.ConversationService, sessions *services.SessionService) *WSHandler {
	return &WSHandler{
		conversations: conversations,
		sessions:      sessions,
		ipConns:       newIPConnTracker(),
		userRL:        middleware.NewUserRateLimiter(),
		conns:         make(map[*safeConn]struct{}),
		maxFPS:        wsMaxFPS(),
	}
}

// wsMaxFPS reads WS_MAX_FPS from env with fallback to defaultMaxFPS.
// Values outside [1, 30] are clamped to keep behaviour predictable.
func wsMaxFPS() int {
	if raw := os.Getenv("WS_MAX_FPS"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			if v < 1 {
				return 1
			}
			if v > 30 {
				return 30
			}
			return v
		}
		slog.Warn("WS_MAX_FPS is not a valid integer — using default", "value", raw, "default", defaultMaxFPS)
	}
	return defaultMaxFPS
}

func (h *WSHandler) effectivePingInterval() time.Duration {
	if h.PingInterval > 0 {
		return h.PingInterval
	}
	return pingInterval
}

func (h *WSHandler) effectivePongWait() time.Duration {
	if h.PongWait > 0 {
		return h.PongWait
	}
	return pongWait
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
	for sc := range h.conns {
		sc.writeClose(websocket.CloseGoingAway, "server shutting down")
	}
	slog.Info("closed all WebSocket connections", "count", len(h.conns))
}

func (h *WSHandler) register(sc *safeConn) {
	h.connsMu.Lock()
	h.conns[sc] = struct{}{}
	h.connsMu.Unlock()
}

func (h *WSHandler) deregister(sc *safeConn) {
	h.connsMu.Lock()
	delete(h.conns, sc)
	h.connsMu.Unlock()
}

// ---------------------------------------------------------------------------
// Frame rate limiter (per-connection, not concurrent)
// ---------------------------------------------------------------------------

// frameRateLimiter enforces a per-connection limit on camera frames
// using a fixed one-second window counter.
type frameRateLimiter struct {
	max         int
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
	if r.count >= r.max {
		return false
	}
	r.count++
	return true
}

// ---------------------------------------------------------------------------
// Handle
// ---------------------------------------------------------------------------

// Handle is the WebSocket connection handler. It creates a session,
// configures heartbeat and message-size limits, reads incoming JSON messages
// in a loop, dispatches them to the appropriate service method, and writes
// responses back to the client.
func (h *WSHandler) Handle(c *websocket.Conn) {
	// ── Limits & heartbeat setup ─────────────────────────────────────────────

	// Reject messages larger than maxMessageSize (returns CloseMessageTooBig).
	c.SetReadLimit(maxMessageSize)

	// Set initial read deadline; the pong handler resets it on each pong.
	effectivePW := h.effectivePongWait()
	if err := c.SetReadDeadline(time.Now().Add(effectivePW)); err != nil {
		slog.Error("failed to set read deadline", "err", err)
		return
	}
	c.SetPongHandler(func(string) error {
		return c.SetReadDeadline(time.Now().Add(effectivePW))
	})

	// ── IP / connection-limit check ──────────────────────────────────────────
	ip, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		ip = c.RemoteAddr().String()
	}

	if !h.ipConns.add(ip) {
		slog.Warn("WebSocket connection limit reached for IP", "ip", ip, "limit", maxConnsPerIP)
		// Best-effort close frame; error is expected if the connection is already disrupted.
		if err := c.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "connection limit reached"),
		); err != nil {
			slog.Debug("close frame write failed on limit reject", "err", err)
		}
		return
	}

	// Extract the authenticated user ID stored by WSAuthMiddleware.
	// Empty string in dev mode (no JWT secret configured).
	userID, _ := c.Locals("user_id").(string)

	sc := &safeConn{c: c}
	h.activeConns.Add(1)
	metrics.WsActiveConnections.Inc()
	h.register(sc)

	slog.Info("WebSocket client connected",
		"remote_addr", c.RemoteAddr(),
		"active_conns", h.activeConns.Load(),
	)

	// ── Session ───────────────────────────────────────────────────────────────
	lang := c.Query("lang", "it")
	session, err := h.sessions.CreateSession("unknown", "unknown", lang)
	if err != nil {
		slog.Error("failed to create session", "err", err, "remote_addr", c.RemoteAddr())
		h.activeConns.Add(-1)
		h.ipConns.remove(ip)
		h.deregister(sc)
		return
	}
	sessionID := session.SessionID

	// Track active sessions per language.
	metrics.WsActiveSessionsByLanguage.WithLabelValues(lang).Inc()

	defer func() {
		metrics.WsActiveSessionsByLanguage.WithLabelValues(lang).Dec()
		if delErr := h.sessions.DeleteSession(sessionID); delErr != nil {
			slog.Warn("failed to delete session", "session_id", sessionID, "err", delErr)
		}
		h.activeConns.Add(-1)
		metrics.WsActiveConnections.Dec()
		h.ipConns.remove(ip)
		h.deregister(sc)
		slog.Info("WebSocket client disconnected",
			"remote_addr", c.RemoteAddr(),
			"session_id", sessionID,
			"active_conns", h.activeConns.Load(),
		)
	}()

	// ── Ping goroutine ────────────────────────────────────────────────────────
	// Sends a Ping every pingInterval. If the client doesn't reply with Pong
	// within pongWait the read deadline expires and the read loop exits,
	// closing the connection and freeing all resources.
	done := make(chan struct{})
	defer close(done)
	go h.pingLoop(sc, done, sessionID)

	// ── Message loop ──────────────────────────────────────────────────────────
	rateLimiter := &frameRateLimiter{max: h.maxFPS, windowStart: time.Now()}

	for {
		var msg IncomingMessage
		if err := c.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("unexpected WebSocket close", "err", err, "session_id", sessionID)
			}
			break
		}

		msgType := msg.Type
		if msgType == "" {
			msgType = "unknown"
		}
		metrics.WsMessagesTotal.WithLabelValues(msgType).Inc()

		switch msg.Type {
		case "frame":
			h.handleFrame(sc, sessionID, userID, msg, rateLimiter)
		case "text":
			h.handleText(sc, sessionID, userID, msg)
		case "ping":
			h.handlePing(sc)
		default:
			slog.Warn("unknown message type",
				"type", msg.Type,
				"remote_addr", c.RemoteAddr(),
				"session_id", sessionID,
			)
		}
	}
}

// pingLoop sends Ping frames every effectivePingInterval until done is closed or a
// write fails (which means the connection is gone).
func (h *WSHandler) pingLoop(sc *safeConn, done <-chan struct{}, sessionID string) {
	ticker := time.NewTicker(h.effectivePingInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := sc.writePing(); err != nil {
				slog.Warn("ping failed — connection likely dead",
					"session_id", sessionID, "err", err)
				return
			}
		case <-done:
			return
		}
	}
}

// handleFrame processes a camera frame through the conversation service.
// Frames that exceed the per-connection FPS cap or the per-user API budget are dropped.
func (h *WSHandler) handleFrame(sc *safeConn, sessionID, userID string, msg IncomingMessage, rl *frameRateLimiter) {
	if !rl.allow() {
		return
	}
	if !h.userRL.Allow(userID) {
		writeError(sc, "rate limit exceeded — slow down")
		return
	}

	ctx := context.Background()
	t0 := time.Now()
	result, err := h.conversations.ProcessFrame(ctx, sessionID, msg.Data, "")
	metrics.AiInferenceDuration.Observe(time.Since(t0).Seconds())
	if err != nil {
		slog.Error("ProcessFrame error", "session_id", sessionID, "err", err)
		metrics.GeminiErrorsTotal.WithLabelValues("frame").Inc()
		writeError(sc, err.Error())
		return
	}

	// nil result means duplicate frame was skipped
	if result == nil {
		return
	}

	metrics.WsFramesProcessedTotal.Inc()
	writeResponse(sc, result)
}

// handleText processes a text-only conversation turn.
func (h *WSHandler) handleText(sc *safeConn, sessionID, userID string, msg IncomingMessage) {
	if !h.userRL.Allow(userID) {
		writeError(sc, "rate limit exceeded — slow down")
		return
	}

	ctx := context.Background()
	t0 := time.Now()
	result, err := h.conversations.ProcessTextMessage(ctx, sessionID, msg.Content)
	metrics.AiInferenceDuration.Observe(time.Since(t0).Seconds())
	if err != nil {
		slog.Error("ProcessTextMessage error", "session_id", sessionID, "err", err)
		metrics.GeminiErrorsTotal.WithLabelValues("text").Inc()
		writeError(sc, err.Error())
		return
	}

	if result == nil {
		return
	}

	writeResponse(sc, result)
}

// handlePing responds to a client ping with a pong message.
func (h *WSHandler) handlePing(sc *safeConn) {
	out := OutgoingMessage{Type: "pong"}
	if err := sc.writeJSON(out); err != nil {
		slog.Warn("failed to write pong", "err", err)
	}
}

// writeResponse sends a successful analysis result to the client.
func writeResponse(sc *safeConn, result *services.ProcessResult) {
	out := OutgoingMessage{
		Type:      "response",
		Text:      result.Text,
		VoiceText: result.VoiceText,
		Overlay:   result.Overlay,
		Step:      result.Step,
	}
	if err := sc.writeJSON(out); err != nil {
		slog.Warn("failed to write response", "err", err)
	}
}

// writeError sends an error message to the client without closing
// the connection.
func writeError(sc *safeConn, errMsg string) {
	out := OutgoingMessage{
		Type:  "error",
		Error: errMsg,
	}
	if err := sc.writeJSON(out); err != nil {
		slog.Warn("failed to write error", "err", err)
	}
}
