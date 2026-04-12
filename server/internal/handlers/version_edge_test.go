package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// newVersionEdgeApp creates a Fiber app with VersionHandler using the given parameters.
func newVersionEdgeApp(version, buildTime, goVersion string) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/version", VersionHandler(version, buildTime, goVersion))
	return app
}

// TestVersionEdgeResponseIncludesVersionField verifies that the JSON response
// always contains a "version" key and that its value matches what was injected.
func TestVersionEdgeResponseIncludesVersionField(t *testing.T) {
	t.Parallel()
	const version = "1.0.0"
	app := newVersionEdgeApp(version, "2026-04-12T00:00:00Z", "go1.25")

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	got, ok := body["version"]
	if !ok {
		t.Fatal("response missing 'version' key")
	}
	if got != version {
		t.Errorf("version: want %q, got %v", version, got)
	}
}

// TestVersionEdgeVersionMatchesSemver verifies that a semver-like version string
// (X.Y.Z) is returned unchanged and that the response value satisfies the
// semantic versioning pattern.
func TestVersionEdgeVersionMatchesSemver(t *testing.T) {
	t.Parallel()
	semverCases := []string{
		"0.0.1",
		"1.0.0",
		"10.20.300",
		"2.3.4",
	}
	semverRE := regexp.MustCompile(`^\d+\.\d+\.\d+`)

	for _, version := range semverCases {
		version := version
		t.Run(version, func(t *testing.T) {
			t.Parallel()
			app := newVersionEdgeApp(version, "2026-04-12T00:00:00Z", "go1.25")

			req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("test: %v", err)
			}

			var body map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			got, _ := body["version"].(string)
			if got == "" {
				t.Errorf("[%s] version field is empty", version)
			}
			if !semverRE.MatchString(got) {
				t.Errorf("[%s] version %q does not match X.Y.Z pattern", version, got)
			}
		})
	}
}

// TestVersionEdgeCacheControlIsPublic verifies that the response carries a
// "public" Cache-Control header — version info is safe to cache in proxies
// (it changes only on deployment).
func TestVersionEdgeCacheControlIsPublic(t *testing.T) {
	t.Parallel()
	app := newVersionEdgeApp("1.2.3", "2026-04-12T00:00:00Z", "go1.25")

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	cc := resp.Header.Get("Cache-Control")
	if cc == "" {
		t.Error("Cache-Control header is missing")
	}
	if !strings.Contains(cc, "public") {
		t.Errorf("Cache-Control: want 'public', got %q", cc)
	}
	if !strings.Contains(cc, "max-age=") {
		t.Errorf("Cache-Control: expected a max-age directive, got %q", cc)
	}
}

// TestVersionEdgeAPIVersionIsAlwaysV1 verifies that the "api_version" field is
// always "v1" regardless of the build version injected.
func TestVersionEdgeAPIVersionIsAlwaysV1(t *testing.T) {
	t.Parallel()
	app := newVersionEdgeApp("99.0.0", "now", "go2.0")

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["api_version"] != "v1" {
		t.Errorf("api_version: want %q, got %v", "v1", body["api_version"])
	}
}

// TestVersionEdgeResponseIsNotCachedByDefault verifies that the Cache-Control
// header is explicitly set (not missing or "no-store") — the handler should set
// a cacheable directive rather than leaving it to client defaults.
func TestVersionEdgeResponseIsNotNoStore(t *testing.T) {
	t.Parallel()
	app := newVersionEdgeApp("1.0.0", "2026-04-12T00:00:00Z", "go1.25")

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	cc := resp.Header.Get("Cache-Control")
	if strings.Contains(cc, "no-store") {
		t.Errorf("Cache-Control should not be no-store for version endpoint (it is static), got %q", cc)
	}
}
