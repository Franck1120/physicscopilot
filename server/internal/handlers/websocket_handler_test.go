package handlers

import (
	"encoding/json"
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
