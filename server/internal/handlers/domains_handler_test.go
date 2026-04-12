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
	counts  map[string]int
}

func (m *mockDomainsService) KBDomains() []string { return m.domains }

func (m *mockDomainsService) DomainEntryCount(domain string) int {
	if m.counts == nil {
		return 0
	}
	return m.counts[domain]
}

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

func TestDomainsHandlerDetailedReturnsCountPerDomain(t *testing.T) {
	svc := &mockDomainsService{
		domains: []string{"hvac", "printer"},
		counts:  map[string]int{"hvac": 20, "printer": 8},
	}
	app := newDomainsApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/domains?detailed=true", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	var body struct {
		Domains []DomainDetail `json:"domains"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(body.Domains) != 2 {
		t.Fatalf("domains: want 2, got %d", len(body.Domains))
	}

	byName := make(map[string]DomainDetail, len(body.Domains))
	for _, d := range body.Domains {
		byName[d.Name] = d
	}

	if d, ok := byName["hvac"]; !ok {
		t.Error("detailed: expected 'hvac' in response")
	} else if d.Count != 20 {
		t.Errorf("hvac count: want 20, got %d", d.Count)
	}

	if d, ok := byName["printer"]; !ok {
		t.Error("detailed: expected 'printer' in response")
	} else if d.Count != 8 {
		t.Errorf("printer count: want 8, got %d", d.Count)
	}
}

func TestDomainsHandlerDetailedFalseReturnsFlatArray(t *testing.T) {
	svc := &mockDomainsService{
		domains: []string{"hvac", "printer"},
		counts:  map[string]int{"hvac": 20, "printer": 8},
	}
	app := newDomainsApp(svc)

	// detailed=false must behave exactly like the default (flat string array).
	req := httptest.NewRequest(http.MethodGet, "/domains?detailed=false", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	var flatBody map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&flatBody); err != nil {
		t.Fatalf("decode flat domains: %v", err)
	}

	got, ok := flatBody["domains"]
	if !ok {
		t.Fatal("response body missing 'domains' key")
	}

	want := map[string]bool{"hvac": true, "printer": true}
	if len(got) != len(want) {
		t.Errorf("domains length: want %d, got %d", len(want), len(got))
	}
	for _, d := range got {
		if !want[d] {
			t.Errorf("unexpected domain in response: %q", d)
		}
	}
}

func TestDomainsHandlerDetailedEmptyWhenNilService(t *testing.T) {
	app := newDomainsApp(nil)

	req := httptest.NewRequest(http.MethodGet, "/domains?detailed=true", nil)
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
