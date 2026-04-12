package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// sessionPayload is a minimal subset of the session response used in E2E tests.
type sessionPayload struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Device struct {
		Brand string `json:"brand"`
		Model string `json:"model"`
	} `json:"device"`
}

type sessionListPayload struct {
	Sessions []sessionPayload `json:"sessions"`
	Count    int              `json:"count"`
}

// TestE2ESessionLifecycle exercises the complete session lifecycle using the
// real Fiber app (no listener started):
//
//  1. Health check → 200
//  2. Create session → 201 + ID
//  3. List sessions → count == 1
//  4. Get session by ID → 200
//  5. Delete session → 204
//  6. Get deleted session → 404
func TestE2ESessionLifecycle(t *testing.T) {
	app := buildTestApp(t)

	// ── 1. Health check ───────────────────────────────────────────────────────
	resp := mustTest(t, app, http.MethodGet, "/health", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health: want 200, got %d", resp.StatusCode)
	}

	// ── 2. Create session ─────────────────────────────────────────────────────
	body := `{"device_brand":"Prusa","device_model":"MK4"}`
	resp = mustTest(t, app, http.MethodPost, "/api/sessions", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create session: want 201, got %d", resp.StatusCode)
	}

	var created sessionPayload
	mustDecode(t, resp.Body, &created)
	if created.ID == "" {
		t.Fatal("create session: empty ID in response")
	}
	if created.Status != "active" {
		t.Errorf("create session: status want 'active', got %q", created.Status)
	}
	if created.Device.Brand != "Prusa" || created.Device.Model != "MK4" {
		t.Errorf("create session: device mismatch got brand=%q model=%q",
			created.Device.Brand, created.Device.Model)
	}

	// ── 3. List sessions ──────────────────────────────────────────────────────
	resp = mustTest(t, app, http.MethodGet, "/api/sessions", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list sessions: want 200, got %d", resp.StatusCode)
	}

	var list sessionListPayload
	mustDecode(t, resp.Body, &list)
	if list.Count != 1 {
		t.Errorf("list sessions: want count=1, got %d", list.Count)
	}
	if len(list.Sessions) != 1 || list.Sessions[0].ID != created.ID {
		t.Errorf("list sessions: expected session %q in list", created.ID)
	}

	// ── 4. Get session ────────────────────────────────────────────────────────
	resp = mustTest(t, app, http.MethodGet, "/api/sessions/"+created.ID, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get session: want 200, got %d", resp.StatusCode)
	}

	var fetched sessionPayload
	mustDecode(t, resp.Body, &fetched)
	if fetched.ID != created.ID {
		t.Errorf("get session: ID mismatch want %q got %q", created.ID, fetched.ID)
	}

	// ── 5. Delete session ─────────────────────────────────────────────────────
	resp = mustTest(t, app, http.MethodDelete, "/api/sessions/"+created.ID, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete session: want 204, got %d", resp.StatusCode)
	}

	// ── 6. Verify 404 after delete ────────────────────────────────────────────
	resp = mustTest(t, app, http.MethodGet, "/api/sessions/"+created.ID, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("get deleted session: want 404, got %d", resp.StatusCode)
	}
}

// TestE2ECreateSessionValidation verifies that the sanitization rules are
// enforced end-to-end through the Fiber app.
func TestE2ECreateSessionValidation(t *testing.T) {
	app := buildTestApp(t)

	cases := []struct {
		name string
		body string
		want int
	}{
		{
			name: "HTML injection rejected",
			body: `{"device_brand":"<script>","device_model":"MK4"}`,
			want: http.StatusBadRequest,
		},
		{
			name: "field too long rejected",
			body: `{"device_brand":"` + strings.Repeat("x", 101) + `","device_model":"MK4"}`,
			want: http.StatusBadRequest,
		},
		{
			name: "empty body accepted",
			body: `{}`,
			want: http.StatusCreated,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := mustTest(t, app, http.MethodPost, "/api/sessions", tc.body)
			if resp.StatusCode != tc.want {
				t.Errorf("want %d, got %d", tc.want, resp.StatusCode)
			}
		})
	}
}

// TestE2EDocsEndpoint verifies that GET /api/docs returns YAML.
func TestE2EDocsEndpoint(t *testing.T) {
	app := buildTestApp(t)

	resp := mustTest(t, app, http.MethodGet, "/api/docs", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/docs: want 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "yaml") {
		t.Errorf("/api/docs: Content-Type want yaml, got %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "openapi:") {
		t.Error("/api/docs: body does not contain 'openapi:' — spec may be empty")
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func mustTest(t *testing.T, app interface{ Test(*http.Request, ...int) (*http.Response, error) }, method, path, body string) *http.Response {
	t.Helper()
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func mustDecode(t *testing.T, r io.Reader, v any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
