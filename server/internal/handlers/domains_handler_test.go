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

// TestDomainsHandlerETagPresent verifies that a 200 response carries an ETag header.
func TestDomainsHandlerETagPresent(t *testing.T) {
	svc := &mockDomainsService{domains: []string{"hvac", "printer"}}
	app := newDomainsApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Error("ETag header missing on 200 response")
	}
	if !strings.HasPrefix(etag, `W/"`) {
		t.Errorf("ETag should be a weak ETag, got %q", etag)
	}
}

// TestDomainsHandlerReturns304WhenETagMatches verifies that sending the ETag
// from a previous response in If-None-Match yields a 304 Not Modified.
func TestDomainsHandlerReturns304WhenETagMatches(t *testing.T) {
	svc := &mockDomainsService{domains: []string{"hvac", "printer"}}
	app := newDomainsApp(svc)

	// First request to get the ETag.
	req1 := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp1, err := app.Test(req1)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	etag := resp1.Header.Get("ETag")
	if etag == "" {
		t.Fatal("ETag missing from first response")
	}

	// Second request with the ETag — expect 304.
	req2 := httptest.NewRequest(http.MethodGet, "/domains", nil)
	req2.Header.Set("If-None-Match", etag)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	if resp2.StatusCode != http.StatusNotModified {
		t.Errorf("status: want 304, got %d", resp2.StatusCode)
	}
}

// TestDomainsHandlerCacheControlHeader verifies that a 200 response includes
// Cache-Control: public, max-age=300.
func TestDomainsHandlerCacheControlHeader(t *testing.T) {
	svc := &mockDomainsService{domains: []string{"printer"}}
	app := newDomainsApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	cc := resp.Header.Get("Cache-Control")
	if cc != "public, max-age=300" {
		t.Errorf("Cache-Control: want %q, got %q", "public, max-age=300", cc)
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
