// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func newVersionApp(version, buildTime, goVersion, commitHash string) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/version", VersionHandler(version, buildTime, goVersion, commitHash))
	return app
}

func TestVersionHandlerReturnsJSON(t *testing.T) {
	app := newVersionApp("1.2.3", "2026-04-12T00:00:00Z", "go1.25.0", "abc1234")

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	var body VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.Version != "1.2.3" {
		t.Errorf("version: want %q, got %q", "1.2.3", body.Version)
	}
	if body.BuildTime != "2026-04-12T00:00:00Z" {
		t.Errorf("build_time: want %q, got %q", "2026-04-12T00:00:00Z", body.BuildTime)
	}
	if body.GoVersion != "go1.25.0" {
		t.Errorf("go_version: want %q, got %q", "go1.25.0", body.GoVersion)
	}
}

func TestVersionHandlerIncludesCommitHash(t *testing.T) {
	const wantCommit = "deadbeef1234"
	app := newVersionApp("0.1.0", "2026-04-12T00:00:00Z", "go1.25.0", wantCommit)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	var body VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.CommitHash != wantCommit {
		t.Errorf("commit_hash: want %q, got %q", wantCommit, body.CommitHash)
	}
}

func TestVersionHandlerDevFallback(t *testing.T) {
	// When GIT_COMMIT_HASH is unset the caller passes "dev" as the fallback.
	app := newVersionApp("0.1.0", "unknown", "go1.25.0", "dev")

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.CommitHash != "dev" {
		t.Errorf("commit_hash fallback: want %q, got %q", "dev", body.CommitHash)
	}
}
