package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/handlers"
	"github.com/Franck1120/physicscopilot/server/internal/services"
)

func TestNewFiberAppHealthEndpoint(t *testing.T) {
	sessions := services.NewSessionService()
	convSvc := services.NewConversationService(sessions, nil, nil)
	ws := handlers.NewWSHandler(convSvc, sessions)
	sessionHandler := handlers.NewSessionHandler(sessions)
	feedbackHandler := handlers.NewFeedbackHandler(nil)

	app := newFiberApp("test", "unknown", "dev", sessionHandler, feedbackHandler, ws, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestNewFiberAppVersionRouteRegistered(t *testing.T) {
	sessions := services.NewSessionService()
	convSvc := services.NewConversationService(sessions, nil, nil)
	ws := handlers.NewWSHandler(convSvc, sessions)
	sessionHandler := handlers.NewSessionHandler(sessions)
	feedbackHandler := handlers.NewFeedbackHandler(nil)

	app := newFiberApp("test", "unknown", "dev", sessionHandler, feedbackHandler, ws, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/version: want 200, got %d", resp.StatusCode)
	}
}

func TestServerStartTime(t *testing.T) {
	if startTime.IsZero() {
		t.Error("startTime should be initialized at package load time")
	}
	if time.Since(startTime) < 0 {
		t.Error("startTime should be in the past")
	}
}
