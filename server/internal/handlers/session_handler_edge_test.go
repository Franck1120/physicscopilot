package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCreateSessionBodyLargerThanFiberDefaultReturns4xx verifies that a POST
// body that is unreasonably large (> 1 MB) is rejected by Fiber before
// reaching the handler.  Fiber's default body limit is 4 MB; we send 2 MB of
// JSON-like noise which is valid size-wise but should be rejected as bad JSON,
// returning 400 Bad Request.
func TestCreateSessionBodyLargerThanFiberDefaultReturns4xx(t *testing.T) {
	t.Parallel()
	app, _ := newSessionTestApp(t)

	// 2 MB of garbage — not valid JSON, so BodyParser will return 400.
	bigBody := strings.Repeat("x", 2*1024*1024)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(bigBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 10_000) // 10-second timeout
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode < 400 {
		t.Errorf("large garbage body: want 4xx, got %d", resp.StatusCode)
	}
}

// TestGetSessionIDWithSlashReturns404 verifies that a session ID containing a
// slash character is not found (Fiber will treat the extra path segment as a
// sub-route, so the param will only contain the part up to the first slash).
func TestGetSessionIDWithSlashReturns404(t *testing.T) {
	t.Parallel()
	app, _ := newSessionTestApp(t)

	// Fiber route "/api/sessions/:id" — a literal slash inside the ID is not
	// captured, so the session will never be found.
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/abc%2Fdef", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	// Either 404 (id not found) or 400 (invalid param) — both are acceptable.
	if resp.StatusCode == http.StatusOK {
		t.Errorf("slash-in-ID: want non-200, got 200")
	}
}

// TestGetSessionIDWithDotDotReturns404 verifies that a path-traversal-style
// session ID ("..") does not expose anything and returns a 404.
func TestGetSessionIDWithDotDotReturns404(t *testing.T) {
	t.Parallel()
	app, _ := newSessionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/..", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Errorf("dotdot ID: want non-200, got 200")
	}
}

// TestGetSessionIDWithSQLInjectionReturns404 verifies that a session ID that
// looks like a SQL injection fragment does not match any real session.
func TestGetSessionIDWithSQLInjectionReturns404(t *testing.T) {
	t.Parallel()
	app, _ := newSessionTestApp(t)

	sqlID := "1%27%20OR%20%271%27%3D%271" // URL-encoded: 1' OR '1'='1
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sqlID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Errorf("SQL injection ID: want non-200, got 200")
	}
}

// TestDeleteSessionAlreadyDeletedIsIdempotent verifies that deleting a session
// that was already deleted returns 404 (idempotency: the server is consistent
// — the resource is gone regardless of whether we deleted it or it never existed).
func TestDeleteSessionAlreadyDeletedIsIdempotent(t *testing.T) {
	t.Parallel()
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Prusa", "MK4", "", "")

	// First delete — must succeed.
	del1 := httptest.NewRequest(http.MethodDelete, "/api/sessions/"+sess.SessionID, nil)
	resp1, err := app.Test(del1)
	if err != nil {
		t.Fatalf("first delete: %v", err)
	}
	if resp1.StatusCode != http.StatusNoContent {
		t.Fatalf("first delete: want 204, got %d", resp1.StatusCode)
	}

	// Second delete of the same (now absent) session — must return 404.
	del2 := httptest.NewRequest(http.MethodDelete, "/api/sessions/"+sess.SessionID, nil)
	resp2, err := app.Test(del2)
	if err != nil {
		t.Fatalf("second delete: %v", err)
	}
	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("second delete (idempotency): want 404, got %d", resp2.StatusCode)
	}
}

// TestListSessionsEmptyReturnsArrayNotNull verifies that when there are no
// sessions the "sessions" field in the response JSON is an array literal []
// rather than the JSON null value.
func TestListSessionsEmptyReturnsArrayNotNull(t *testing.T) {
	t.Parallel()
	app, _ := newSessionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}

	sessionsRaw, ok := raw["sessions"]
	if !ok {
		t.Fatal("response missing 'sessions' key")
	}
	// Must be "[]" not "null".
	if string(sessionsRaw) == "null" {
		t.Errorf("sessions: want [], got null")
	}
	// Confirm it decodes as a (possibly empty) array.
	var arr []interface{}
	if err := json.Unmarshal(sessionsRaw, &arr); err != nil {
		t.Errorf("sessions field is not an array: %v", err)
	}
	if len(arr) != 0 {
		t.Errorf("sessions: want empty array, got %d elements", len(arr))
	}
}

// TestGetSessionIDWithQuestionMarkReturns404 verifies that a session ID
// containing a literal question mark (which could be confused with a query
// string delimiter) is handled safely and returns a non-200 response.
func TestGetSessionIDWithQuestionMarkReturns404(t *testing.T) {
	t.Parallel()
	app, _ := newSessionTestApp(t)

	// Use URL-encoded "?" — Fiber resolves the param before the query string.
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/abc%3Fid%3D1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Errorf("question-mark ID: want non-200, got 200")
	}
}
