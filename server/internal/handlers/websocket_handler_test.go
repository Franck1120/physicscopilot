package handlers

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	gorilla "github.com/gorilla/websocket"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

func TestNewWSHandler(t *testing.T) {
	sessionSvc := services.NewSessionService()
	// NewGeminiService requires GEMINI_API_KEY, so we pass nil for gemini
	// to test the handler constructor in isolation.
	convSvc := services.NewConversationService(sessionSvc, nil, nil)

	handler := NewWSHandler(convSvc, sessionSvc)

	if handler == nil {
		t.Fatal("expected non-nil WSHandler")
	}
	if handler.conversations == nil {
		t.Error("expected non-nil conversations service")
	}
	if handler.sessions == nil {
		t.Error("expected non-nil sessions service")
	}
}

func TestFrameRateLimiterAllowsUpToLimit(t *testing.T) {
	rl := &frameRateLimiter{max: defaultMaxFPS, windowStart: time.Now()}

	for i := 0; i < defaultMaxFPS; i++ {
		if !rl.allow() {
			t.Errorf("expected frame %d to be allowed", i+1)
		}
	}

	if rl.count != defaultMaxFPS {
		t.Errorf("expected count %d, got %d", defaultMaxFPS, rl.count)
	}
}

func TestFrameRateLimiterBlocksExcess(t *testing.T) {
	rl := &frameRateLimiter{max: defaultMaxFPS, windowStart: time.Now()}

	// Exhaust the limit
	for i := 0; i < defaultMaxFPS; i++ {
		rl.allow()
	}

	// Next frame should be blocked
	if rl.allow() {
		t.Error("expected frame beyond limit to be blocked")
	}
}

func TestFrameRateLimiterResetsAfterOneSecond(t *testing.T) {
	// Start the window in the past so the next call triggers a reset
	pastStart := time.Now().Add(-2 * time.Second)
	rl := &frameRateLimiter{
		max:         defaultMaxFPS,
		count:       defaultMaxFPS,
		windowStart: pastStart,
	}

	// Should reset and allow
	if !rl.allow() {
		t.Error("expected frame to be allowed after window reset")
	}
	if rl.count != 1 {
		t.Errorf("expected count 1 after reset, got %d", rl.count)
	}
}

func TestFrameRateLimiterWindowBoundary(t *testing.T) {
	// Simulate being right at the boundary: window started exactly 1 second ago
	rl := &frameRateLimiter{
		max:         defaultMaxFPS,
		count:       defaultMaxFPS,
		windowStart: time.Now().Add(-time.Second),
	}

	// time.Now().Sub(windowStart) >= time.Second should be true, so reset happens
	if !rl.allow() {
		t.Error("expected frame to be allowed when window has expired")
	}
}

func TestIncomingMessageJSONDeserialization(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    IncomingMessage
	}{
		{
			name:  "frame message",
			input: `{"type":"frame","data":"base64data==","timestamp":1700000000000}`,
			want: IncomingMessage{
				Type:      "frame",
				Data:      "base64data==",
				Timestamp: 1700000000000,
			},
		},
		{
			name:  "text message",
			input: `{"type":"text","content":"help me fix this","timestamp":1700000000001}`,
			want: IncomingMessage{
				Type:      "text",
				Content:   "help me fix this",
				Timestamp: 1700000000001,
			},
		},
		{
			name:  "ping message",
			input: `{"type":"ping","timestamp":1700000000002}`,
			want: IncomingMessage{
				Type:      "ping",
				Timestamp: 1700000000002,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got IncomingMessage
			if err := json.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type: got %q, want %q", got.Type, tt.want.Type)
			}
			if got.Data != tt.want.Data {
				t.Errorf("Data: got %q, want %q", got.Data, tt.want.Data)
			}
			if got.Content != tt.want.Content {
				t.Errorf("Content: got %q, want %q", got.Content, tt.want.Content)
			}
			if got.Timestamp != tt.want.Timestamp {
				t.Errorf("Timestamp: got %d, want %d", got.Timestamp, tt.want.Timestamp)
			}
		})
	}
}

func TestOutgoingMessageJSONSerialization(t *testing.T) {
	tests := []struct {
		name string
		msg  OutgoingMessage
		check func(t *testing.T, data map[string]interface{})
	}{
		{
			name: "response message",
			msg: OutgoingMessage{
				Type: "response",
				Text: "Detected stringing on the left side",
				Overlay: services.OverlayData{
					Boxes: []services.BoundingBox{
						{X: 0.1, Y: 0.2, W: 0.3, H: 0.4, Label: "stringing"},
					},
					Arrows: []services.Arrow{},
				},
				Step: services.StepInfo{Current: 2, Total: 5},
			},
			check: func(t *testing.T, data map[string]interface{}) {
				if data["type"] != "response" {
					t.Errorf("expected type 'response', got %v", data["type"])
				}
				if data["text"] != "Detected stringing on the left side" {
					t.Errorf("unexpected text: %v", data["text"])
				}
				if _, ok := data["error"]; ok {
					t.Error("error field should be omitted for response messages")
				}
			},
		},
		{
			name: "error message",
			msg: OutgoingMessage{
				Type:  "error",
				Error: "session not found",
			},
			check: func(t *testing.T, data map[string]interface{}) {
				if data["type"] != "error" {
					t.Errorf("expected type 'error', got %v", data["type"])
				}
				if data["error"] != "session not found" {
					t.Errorf("unexpected error: %v", data["error"])
				}
				if _, ok := data["text"]; ok {
					t.Error("text field should be omitted for error messages")
				}
			},
		},
		{
			name: "pong message",
			msg: OutgoingMessage{
				Type: "pong",
			},
			check: func(t *testing.T, data map[string]interface{}) {
				if data["type"] != "pong" {
					t.Errorf("expected type 'pong', got %v", data["type"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			var data map[string]interface{}
			if err := json.Unmarshal(b, &data); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			tt.check(t, data)
		})
	}
}

func TestOutgoingMessageOmitsEmptyFields(t *testing.T) {
	msg := OutgoingMessage{Type: "pong"}
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, ok := data["text"]; ok {
		t.Error("text should be omitted when empty")
	}
	if _, ok := data["error"]; ok {
		t.Error("error should be omitted when empty")
	}
}

func TestDefaultMaxFPS(t *testing.T) {
	if defaultMaxFPS != 5 {
		t.Errorf("expected defaultMaxFPS to be 5, got %d", defaultMaxFPS)
	}
	// wsMaxFPS reads WS_MAX_FPS from env; without it should return the default.
	if got := wsMaxFPS(); got != defaultMaxFPS {
		t.Errorf("wsMaxFPS() = %d, want %d (no WS_MAX_FPS set)", got, defaultMaxFPS)
	}
}

// TestWSHandlerAcceptsConnection verifies that the production WSHandler
// accepts a WebSocket upgrade and sends no immediate error frame.
// The test connects, reads one message (if any arrive within 300 ms),
// and expects the connection to stay alive (no close frame from the server).
func TestWSHandlerAcceptsConnection(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)

	// Use the same short heartbeat as heartbeat_test so the test completes fast.
	wsHandler.PingInterval = time.Hour // don't fire during test
	wsHandler.PongWait = time.Hour

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Set a short read deadline so the test doesn't block waiting for a message.
	conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))

	// Read any initial messages; a timeout means no messages arrived — that is
	// also acceptable (no error close frame was sent).
	_, _, readErr := conn.ReadMessage()
	if readErr != nil {
		// A timeout means the connection is open but idle — expected.
		if gorilla.IsCloseError(readErr, gorilla.CloseGoingAway, gorilla.ClosePolicyViolation) {
			t.Errorf("server closed the connection unexpectedly: %v", readErr)
		}
		// Any other error (timeout, etc.) is acceptable — connection was live.
	}

	// Verify the session was created in the service.
	sessions := sessionSvc.ListSessions()
	if len(sessions) == 0 {
		t.Error("expected WSHandler to create an in-memory session on connect")
	}
}

// ── sanitizeText ─────────────────────────────────────────────────────────────

func TestSanitizeTextStripsHTMLTags(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantOK  bool
	}{
		{"plain text unchanged", "hello world", "hello world", true},
		{"strips bold tag", "<b>bold</b> text", "bold text", true},
		{"strips script tag", "<script>alert(1)</script>", "alert(1)", true},
		{"empty string", "", "", false},
		{"only whitespace", "   ", "", false},
		{"only HTML tags", "<br/>", "", false},
		{"trims surrounding whitespace", "  hello  ", "hello", true},
		{"nested tags", "<div><p>text</p></div>", "text", true},
		{
			"truncates to maxTextContentLen",
			strings.Repeat("a", maxTextContentLen+100),
			strings.Repeat("a", maxTextContentLen),
			true,
		},
		{
			"exactly maxTextContentLen passes unchanged",
			strings.Repeat("b", maxTextContentLen),
			strings.Repeat("b", maxTextContentLen),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := sanitizeText(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ok: got %v, want %v (input len=%d)", ok, tt.wantOK, len(tt.input))
			}
			if got != tt.want {
				short := tt.want
				if len(short) > 40 {
					short = short[:40] + "..."
				}
				t.Errorf("text: want %q, got len=%d", short, len(got))
			}
		})
	}
}

// ── isValidJPEG ──────────────────────────────────────────────────────────────

func TestIsValidJPEG(t *testing.T) {
	// JPEG magic: FF D8 FF E0 00 10
	validJPEG := base64.StdEncoding.EncodeToString([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10})
	notJPEG := base64.StdEncoding.EncodeToString([]byte("hello world"))

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", false},
		{"valid JPEG magic", validJPEG, true},
		{"non-JPEG data", notJPEG, false},
		{"corrupt base64", "not!valid!base64!!!!!", false},
		{"too short base64", base64.StdEncoding.EncodeToString([]byte{0xFF, 0xD8}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidJPEG(tt.input)
			if got != tt.want {
				t.Errorf("isValidJPEG = %v, want %v", got, tt.want)
			}
		})
	}
}

// ── wsMaxFPS env overrides ────────────────────────────────────────────────────

func TestWsMaxFPSFromEnvVar(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{"below min clamped to 1", "0", 1},
		{"negative clamped to 1", "-5", 1},
		{"above max clamped to 30", "100", 30},
		{"valid value 15", "15", 15},
		{"min boundary 1", "1", 1},
		{"max boundary 30", "30", 30},
		{"non-numeric falls back to default", "notanumber", defaultMaxFPS},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WS_MAX_FPS", tt.value)
			got := wsMaxFPS()
			if got != tt.want {
				t.Errorf("wsMaxFPS(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

// ── effectivePingInterval / effectivePongWait ─────────────────────────────────

func TestEffectivePingIntervalUsesOverrideWhenSet(t *testing.T) {
	h := &WSHandler{}
	if got := h.effectivePingInterval(); got != pingInterval {
		t.Errorf("zero PingInterval: want %v, got %v", pingInterval, got)
	}
	h.PingInterval = 5 * time.Second
	if got := h.effectivePingInterval(); got != 5*time.Second {
		t.Errorf("override PingInterval: want 5s, got %v", got)
	}
}

func TestEffectivePongWaitUsesOverrideWhenSet(t *testing.T) {
	h := &WSHandler{}
	if got := h.effectivePongWait(); got != pongWait {
		t.Errorf("zero PongWait: want %v, got %v", pongWait, got)
	}
	h.PongWait = 10 * time.Second
	if got := h.effectivePongWait(); got != 10*time.Second {
		t.Errorf("override PongWait: want 10s, got %v", got)
	}
}

// ── CloseAll ──────────────────────────────────────────────────────────────────

func TestCloseAllWithNoConnectionsDoesNotPanic(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	h := NewWSHandler(convSvc, sessionSvc)
	// Must not panic even when the connection map is empty.
	h.CloseAll()
}

// TestCloseAllSendsCloseFrameToActiveConnections verifies that CloseAll sends a
// GoingAway close frame, causing the connected client to see a close error.
func TestCloseAllSendsCloseFrameToActiveConnections(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)
	wsHandler.PingInterval = time.Hour
	wsHandler.PongWait = time.Hour

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Allow the Handle goroutine to register the connection.
	time.Sleep(50 * time.Millisecond)

	// CloseAll sends a GoingAway close frame to all registered connections.
	wsHandler.CloseAll()

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, _, readErr := conn.ReadMessage()
	if readErr == nil {
		t.Error("expected connection to be closed after CloseAll")
	}
}

// ── WebSocket message handling ────────────────────────────────────────────────

// TestWSHandlerHandlesPingMessage verifies that a "ping" JSON message produces
// a {"type":"pong"} response from the server.
func TestWSHandlerHandlesPingMessage(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)
	wsHandler.PingInterval = time.Hour
	wsHandler.PongWait = time.Hour

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	if err := conn.WriteJSON(IncomingMessage{Type: "ping"}); err != nil {
		t.Fatalf("write ping: %v", err)
	}

	var resp OutgoingMessage
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read pong: %v", err)
	}
	if resp.Type != "pong" {
		t.Errorf("expected pong, got %q", resp.Type)
	}
}

// TestWSHandlerRejectsNonJPEGFrame verifies that a "frame" message whose data
// is not a valid JPEG (wrong magic bytes) returns an error response.
func TestWSHandlerRejectsNonJPEGFrame(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)
	wsHandler.PingInterval = time.Hour
	wsHandler.PongWait = time.Hour

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	notJPEG := base64.StdEncoding.EncodeToString([]byte("not a jpeg image"))
	if err := conn.WriteJSON(IncomingMessage{Type: "frame", Data: notJPEG}); err != nil {
		t.Fatalf("write frame: %v", err)
	}

	var resp OutgoingMessage
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read response: %v", err)
	}
	if resp.Type != "error" {
		t.Errorf("expected error response, got type=%q", resp.Type)
	}
	if !strings.Contains(resp.Error, "JPEG") {
		t.Errorf("expected JPEG error message, got %q", resp.Error)
	}
}

// TestWSHandlerRejectsHTMLOnlyTextMessage verifies that a "text" message whose
// content is entirely HTML tags (empty after sanitization) returns an error.
func TestWSHandlerRejectsHTMLOnlyTextMessage(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)
	wsHandler.PingInterval = time.Hour
	wsHandler.PongWait = time.Hour

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// HTML-only content → empty after sanitization → error before AI call.
	if err := conn.WriteJSON(IncomingMessage{Type: "text", Content: "<br/><script/>"}); err != nil {
		t.Fatalf("write text: %v", err)
	}

	var resp OutgoingMessage
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read response: %v", err)
	}
	if resp.Type != "error" {
		t.Errorf("expected error response, got type=%q", resp.Type)
	}
	if !strings.Contains(resp.Error, "empty after sanitization") {
		t.Errorf("expected sanitization error, got %q", resp.Error)
	}
}

// TestWSHandlerIgnoresUnknownMessageType verifies that an unknown message type
// does not close the connection — the server logs a warning and continues.
func TestWSHandlerIgnoresUnknownMessageType(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)
	wsHandler.PingInterval = time.Hour
	wsHandler.PongWait = time.Hour

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Send unknown type — server should ignore it and keep connection alive.
	if err := conn.WriteJSON(IncomingMessage{Type: "unknown_message_type"}); err != nil {
		t.Fatalf("write unknown type: %v", err)
	}

	// Follow up with a ping to confirm the connection is still alive.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err := conn.WriteJSON(IncomingMessage{Type: "ping"}); err != nil {
		t.Fatalf("write ping after unknown: %v", err)
	}

	var resp OutgoingMessage
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read after unknown type: %v", err)
	}
	if resp.Type != "pong" {
		t.Errorf("expected pong after unknown type, got %q", resp.Type)
	}
}

// ── Ping/pong heartbeat cycle ─────────────────────────────────────────────────

// TestWSHandlerHeartbeatPingCycleKeepsConnectionAlive verifies that the server
// sends WebSocket Ping control frames and the connection stays alive when the
// client responds with Pong frames as expected.
func TestWSHandlerHeartbeatPingCycleKeepsConnectionAlive(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)

	// Use a short ping interval so the test completes quickly.
	wsHandler.PingInterval = 80 * time.Millisecond
	wsHandler.PongWait = 500 * time.Millisecond

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// The gorilla client sends Pong automatically when it receives a Ping
	// (default ping handler). Count how many pings are received.
	pingCount := 0
	conn.SetPingHandler(func(data string) error {
		pingCount++
		// Reply with a Pong control frame to keep the server happy.
		return conn.WriteControl(gorilla.PongMessage, []byte(data), time.Now().Add(time.Second))
	})

	// Read any application messages for 300 ms while the ping handler runs.
	conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	for {
		_, _, readErr := conn.ReadMessage()
		if readErr != nil {
			break
		}
	}

	// With PingInterval=80ms, we expect at least 2 server pings in 300 ms.
	if pingCount < 2 {
		t.Errorf("expected at least 2 ping frames, got %d", pingCount)
	}
}

// ── Graceful close ────────────────────────────────────────────────────────────

// TestWSHandlerGracefulCloseRemovesSession verifies that when the client sends
// a normal close frame the session is removed from the SessionService.
func TestWSHandlerGracefulCloseRemovesSession(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)
	wsHandler.PingInterval = time.Hour
	wsHandler.PongWait = time.Hour

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	// Give the handler goroutine time to register the session.
	time.Sleep(50 * time.Millisecond)

	sessionsBefore := sessionSvc.ListSessions()
	if len(sessionsBefore) == 0 {
		t.Fatal("expected session to be created on connect")
	}

	// Send a normal close frame and close the client side.
	conn.WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, "bye")) //nolint:errcheck
	conn.Close()

	// Allow the server's deferred cleanup to run.
	time.Sleep(100 * time.Millisecond)

	sessionsAfter := sessionSvc.ListSessions()
	if len(sessionsAfter) != 0 {
		t.Errorf("expected session to be removed after disconnect, got %d sessions", len(sessionsAfter))
	}
}

// TestWSHandlerActiveConnectionsDecreasesOnDisconnect checks the active
// connection counter is incremented on connect and decremented on disconnect.
func TestWSHandlerActiveConnectionsDecreasesOnDisconnect(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)
	wsHandler.PingInterval = time.Hour
	wsHandler.PongWait = time.Hour

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	if wsHandler.ActiveConnections() != 0 {
		t.Fatalf("expected 0 active connections before any client, got %d", wsHandler.ActiveConnections())
	}

	conn, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	// Allow Handle goroutine to increment the counter.
	time.Sleep(50 * time.Millisecond)
	if wsHandler.ActiveConnections() != 1 {
		t.Errorf("expected 1 active connection after connect, got %d", wsHandler.ActiveConnections())
	}

	conn.Close()

	// Allow defer in Handle to decrement the counter.
	time.Sleep(100 * time.Millisecond)
	if wsHandler.ActiveConnections() != 0 {
		t.Errorf("expected 0 active connections after disconnect, got %d", wsHandler.ActiveConnections())
	}
}

// ── Handler state after multiple disconnects ──────────────────────────────────

// TestWSHandlerConnsMapEmptyAfterAllClientsDisconnect verifies that the internal
// connection map is empty after all clients disconnect, preventing resource leaks.
func TestWSHandlerConnsMapEmptyAfterAllClientsDisconnect(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	wsHandler := NewWSHandler(convSvc, sessionSvc)
	wsHandler.PingInterval = time.Hour
	wsHandler.PongWait = time.Hour

	addr, shutdown := startTestWSServer(t, wsHandler.Handle)
	defer shutdown()

	const numClients = 3
	conns := make([]*gorilla.Conn, numClients)
	for i := range conns {
		c, _, err := gorilla.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
		if err != nil {
			t.Fatalf("dial client %d: %v", i, err)
		}
		conns[i] = c
	}

	// Allow all Handle goroutines to register.
	time.Sleep(80 * time.Millisecond)

	wsHandler.connsMu.Lock()
	registeredCount := len(wsHandler.conns)
	wsHandler.connsMu.Unlock()
	if registeredCount != numClients {
		t.Errorf("expected %d registered conns, got %d", numClients, registeredCount)
	}

	// Disconnect all clients.
	for _, c := range conns {
		c.Close()
	}

	// Allow deferred deregister to run.
	time.Sleep(150 * time.Millisecond)

	wsHandler.connsMu.Lock()
	remaining := len(wsHandler.conns)
	wsHandler.connsMu.Unlock()
	if remaining != 0 {
		t.Errorf("expected 0 registered conns after all clients disconnected, got %d", remaining)
	}
}
