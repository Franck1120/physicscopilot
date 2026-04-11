package handlers

import (
	"context"
	"log"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/websocket/v2"
)

// maxFramesPerSecond is the maximum number of camera frames a single
// WebSocket connection may send within a one-second sliding window.
const maxFramesPerSecond = 5

// IncomingMessage represents a JSON message received from the client.
type IncomingMessage struct {
	Type      string `json:"type"`      // "frame" | "text" | "ping"
	Data      string `json:"data"`      // base64 image (only for type="frame")
	Content   string `json:"content"`   // user text (only for type="text")
	Timestamp int64  `json:"timestamp"` // ms epoch
}

// OutgoingMessage represents a JSON message sent to the client.
type OutgoingMessage struct {
	Type    string              `json:"type"`              // "response" | "error" | "pong"
	Text    string              `json:"text,omitempty"`
	Overlay services.OverlayData `json:"overlay,omitempty"`
	Step    services.StepInfo    `json:"step,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// WSHandler manages WebSocket connections by coordinating session
// lifecycle and conversation processing through the service layer.
type WSHandler struct {
	conversations *services.ConversationService
	sessions      *services.SessionService
}

// NewWSHandler creates a WSHandler wired to the given conversation
// and session services.
func NewWSHandler(conversations *services.ConversationService, sessions *services.SessionService) *WSHandler {
	return &WSHandler{
		conversations: conversations,
		sessions:      sessions,
	}
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
	log.Println("WebSocket client connected:", c.RemoteAddr())

	session, err := h.sessions.CreateSession("unknown", "unknown")
	if err != nil {
		log.Println("failed to create session:", err)
		return
	}
	sessionID := session.SessionID

	defer func() {
		if delErr := h.sessions.DeleteSession(sessionID); delErr != nil {
			log.Println("failed to delete session:", delErr)
		}
		log.Println("WebSocket client disconnected:", c.RemoteAddr())
	}()

	rateLimiter := &frameRateLimiter{windowStart: time.Now()}

	for {
		var msg IncomingMessage
		if err := c.ReadJSON(&msg); err != nil {
			// Connection closed or malformed message — exit the loop
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Println("unexpected WebSocket close:", err)
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
			log.Printf("unknown message type %q from %s", msg.Type, c.RemoteAddr())
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
		log.Printf("ProcessFrame error (session %s): %v", sessionID, err)
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
		log.Printf("ProcessTextMessage error (session %s): %v", sessionID, err)
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
		log.Println("failed to write pong:", err)
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
		log.Println("failed to write response:", err)
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
		log.Println("failed to write error:", err)
	}
}
