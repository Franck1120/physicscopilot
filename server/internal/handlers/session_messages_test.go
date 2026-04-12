package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// testSessionMessagesApp builds a minimal Fiber app for testing GET /api/sessions/:id/messages.
func testSessionMessagesApp() (*fiber.App, *services.SessionService) {
	svc := services.NewSessionService()
	h := NewSessionHandler(svc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/api/sessions", h.CreateSession)
	app.Get("/api/sessions/:id/messages", h.GetSessionMessages)
	return app, svc
}

// TestGetSessionMessagesExistingSession verifies 200 and a valid JSON payload for a
// session that exists but has no messages yet.
func TestGetSessionMessagesExistingSession(t *testing.T) {
	app, svc := testSessionMessagesApp()

	session, err := svc.CreateSession("Prusa", "MK4", "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+session.SessionID+"/messages", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var payload struct {
		SessionID string        `json:"session_id"`
		Count     int           `json:"count"`
		Messages  []interface{} `json:"messages"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.SessionID != session.SessionID {
		t.Errorf("session_id: want %q, got %q", session.SessionID, payload.SessionID)
	}
	if payload.Count != 0 {
		t.Errorf("count: want 0, got %d", payload.Count)
	}
	if payload.Messages == nil {
		t.Error("messages: expected non-nil empty array, got nil")
	}
}

// TestGetSessionMessagesWithHistory verifies that messages added to a session
// are returned correctly by the endpoint.
func TestGetSessionMessagesWithHistory(t *testing.T) {
	app, svc := testSessionMessagesApp()

	session, err := svc.CreateSession("Creality", "Ender3", "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := svc.AddMessage(session.SessionID, "user", "My nozzle is clogged", false); err != nil {
		t.Fatalf("AddMessage: %v", err)
	}
	if err := svc.AddMessage(session.SessionID, "assistant", "Try cold pull", false); err != nil {
		t.Fatalf("AddMessage: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+session.SessionID+"/messages", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var payload struct {
		Count    int `json:"count"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Count != 2 {
		t.Errorf("count: want 2, got %d", payload.Count)
	}
	if len(payload.Messages) != 2 {
		t.Fatalf("messages length: want 2, got %d", len(payload.Messages))
	}
	if payload.Messages[0].Role != "user" {
		t.Errorf("messages[0].role: want %q, got %q", "user", payload.Messages[0].Role)
	}
	if !strings.Contains(payload.Messages[0].Content, "clogged") {
		t.Errorf("messages[0].content: expected to contain %q", "clogged")
	}
}

// TestGetSessionMessagesNotFound verifies that a 404 is returned for unknown session IDs.
func TestGetSessionMessagesNotFound(t *testing.T) {
	app, _ := testSessionMessagesApp()

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/nonexistent-id/messages", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: want 404, got %d", resp.StatusCode)
	}
}
