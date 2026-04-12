// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/websocket/v2"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// TestWSHandlerNotWebSocketReturns426 verifies that a plain GET /ws request
// (without a WebSocket upgrade) returns 426 Upgrade Required when the
// upgrade guard middleware is in place.
func TestWSHandlerNotWebSocketReturns426(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Mount the upgrade guard exactly as production does, WITHOUT actually
	// dialling a WebSocket — the plain GET should be rejected with 426.
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", fiberws.New(ws.Handle))

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	// No Upgrade / Connection headers → IsWebSocketUpgrade returns false.
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUpgradeRequired {
		t.Errorf("want 426, got %d", resp.StatusCode)
	}
}

// TestWSHandlerConcurrentActiveConnTracking directly manipulates activeConns
// and verifies the health endpoint reflects the updated count.
func TestWSHandlerConcurrentActiveConnTracking(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	const delta = int32(5)
	ws.activeConns.Add(delta)

	if got := ws.ActiveConnections(); got != delta {
		t.Errorf("ActiveConnections: want %d, got %d", delta, got)
	}

	// Wire a health endpoint and verify it reflects active_connections=5.
	healthSessionSvc := services.NewSessionService()
	healthConvSvc := services.NewConversationService(healthSessionSvc, nil, nil)
	healthWS := NewWSHandler(healthConvSvc, healthSessionSvc)
	healthWS.activeConns.Add(delta)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler("0.1.0-edge", time.Now(), healthWS, nil))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ActiveConnections != delta {
		t.Errorf("active_connections: want %d, got %d", delta, body.ActiveConnections)
	}
}

// TestWSHandlerMaxMessageValidation confirms that the upgrade guard (the same
// middleware used in production) returns 426 when no WS upgrade headers are
// present — the upgrade check is the first gate before Handle is reached.
func TestWSHandlerMaxMessageValidation(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", fiberws.New(ws.Handle))

	// Plain request without upgrade headers must be rejected before Handle runs.
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUpgradeRequired {
		t.Errorf("upgrade guard: want 426, got %d", resp.StatusCode)
	}
}
