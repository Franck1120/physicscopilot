// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// mockDomainsService implements DomainsService for tests.
type mockDomainsService struct {
	domains []string
}

func (m *mockDomainsService) KBDomains() []string { return m.domains }

func newDomainsApp(svc DomainsService) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/domains", DomainsHandler(svc))
	return app
}

func TestDomainsHandlerReturnsSortedList(t *testing.T) {
	svc := &mockDomainsService{domains: []string{"hvac", "appliances", "automotive"}}
	app := newDomainsApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: want application/json, got %q", ct)
	}

	var body map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	got, ok := body["domains"]
	if !ok {
		t.Fatal("response body missing 'domains' key")
	}

	want := map[string]bool{"hvac": true, "appliances": true, "automotive": true}
	if len(got) != len(want) {
		t.Errorf("domains length: want %d, got %d", len(want), len(got))
	}
	for _, d := range got {
		if !want[d] {
			t.Errorf("unexpected domain in response: %q", d)
		}
	}
}

func TestDomainsHandlerReturnsEmptyArrayWhenNil(t *testing.T) {
	svc := &mockDomainsService{domains: nil}
	app := newDomainsApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	var body map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	raw, ok := body["domains"]
	if !ok {
		t.Fatal("response body missing 'domains' key")
	}
	// Must be "[]" — not "null".
	if string(raw) != "[]" {
		t.Errorf("domains: want [], got %s", string(raw))
	}
}

func TestDomainsHandlerReturnsEmptyArrayWhenEmpty(t *testing.T) {
	svc := &mockDomainsService{domains: []string{}}
	app := newDomainsApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	var body map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	raw, ok := body["domains"]
	if !ok {
		t.Fatal("response body missing 'domains' key")
	}
	// Must be "[]" — not "null".
	if string(raw) != "[]" {
		t.Errorf("domains: want [], got %s", string(raw))
	}
}
