package handlers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

func TestNewWSHandler(t *testing.T) {
	sessionSvc := services.NewSessionService()
	// NewGeminiService requires GEMINI_API_KEY, so we pass nil for gemini
	// to test the handler constructor in isolation.
	convSvc := services.NewConversationService(sessionSvc, nil)

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
	rl := &frameRateLimiter{windowStart: time.Now()}

	for i := 0; i < maxFramesPerSecond; i++ {
		if !rl.allow() {
			t.Errorf("expected frame %d to be allowed", i+1)
		}
	}

	if rl.count != maxFramesPerSecond {
		t.Errorf("expected count %d, got %d", maxFramesPerSecond, rl.count)
	}
}

func TestFrameRateLimiterBlocksExcess(t *testing.T) {
	rl := &frameRateLimiter{windowStart: time.Now()}

	// Exhaust the limit
	for i := 0; i < maxFramesPerSecond; i++ {
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
		count:       maxFramesPerSecond,
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
		count:       maxFramesPerSecond,
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

func TestMaxFramesPerSecondConstant(t *testing.T) {
	if maxFramesPerSecond != 5 {
		t.Errorf("expected maxFramesPerSecond to be 5, got %d", maxFramesPerSecond)
	}
}
