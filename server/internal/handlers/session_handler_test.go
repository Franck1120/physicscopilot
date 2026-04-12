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

// newSessionTestApp wires a SessionHandler to a fresh Fiber app and returns both.
func newSessionTestApp(t *testing.T) (*fiber.App, *services.SessionService) {
	t.Helper()
	sessions := services.NewSessionService()
	h := NewSessionHandler(sessions)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/api/sessions", h.CreateSession)
	app.Get("/api/sessions", h.ListSessions)
	app.Get("/api/sessions/:id", h.GetSession)
	app.Delete("/api/sessions/:id", h.DeleteSession)
	return app, sessions
}

func TestCreateSessionReturns201(t *testing.T) {
	app, _ := newSessionTestApp(t)

	body := `{"device_brand":"Prusa","device_model":"MK4"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 201, got %d: %s", resp.StatusCode, b)
	}

	var dto sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if dto.ID == "" {
		t.Error("expected non-empty id in response")
	}
	if dto.Device.Brand != "Prusa" {
		t.Errorf("expected brand 'Prusa', got %q", dto.Device.Brand)
	}
	if dto.Device.Model != "MK4" {
		t.Errorf("expected model 'MK4', got %q", dto.Device.Model)
	}
}

func TestCreateSessionEmptyBodyCreatesSession(t *testing.T) {
	// Empty brand/model should still succeed (server accepts unknown devices).
	app, _ := newSessionTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("want 201, got %d", resp.StatusCode)
	}
}

func TestCreateSessionBadBodyReturns400(t *testing.T) {
	app, _ := newSessionTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader("not-json{{"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

func TestListSessionsEmpty(t *testing.T) {
	app, _ := newSessionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck
	if count, ok := body["count"].(float64); !ok || count != 0 {
		t.Errorf("expected count 0, got %v", body["count"])
	}
}

func TestListSessionsAfterCreate(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("Bambu", "X1C", "", "") //nolint:errcheck
	sessions.CreateSession("Prusa", "MK4", "", "") //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck
	if count, ok := body["count"].(float64); !ok || int(count) != 2 {
		t.Errorf("expected count 2, got %v", body["count"])
	}
}

func TestGetSessionReturnsState(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Bambu", "X1C", "", "")

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 200, got %d: %s", resp.StatusCode, b)
	}

	var dto sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if dto.ID != sess.SessionID {
		t.Errorf("want id %q, got %q", sess.SessionID, dto.ID)
	}
	if dto.Device.Brand != "Bambu" {
		t.Errorf("want brand 'Bambu', got %q", dto.Device.Brand)
	}
}

func TestGetSessionNotFoundReturns404(t *testing.T) {
	app, _ := newSessionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/nonexistent-id", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestDeleteSessionReturns204(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Prusa", "MK4", "", "")

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/"+sess.SessionID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("want 204, got %d", resp.StatusCode)
	}
}

func TestDeleteSessionNotFoundReturns404(t *testing.T) {
	app, _ := newSessionTestApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/nonexistent-id", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestDeleteSessionThenGetReturns404(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Prusa", "MK4", "", "")

	// Delete
	delReq := httptest.NewRequest(http.MethodDelete, "/api/sessions/"+sess.SessionID, nil)
	app.Test(delReq) //nolint:errcheck

	// Get after delete must return 404
	getReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID, nil)
	resp, err := app.Test(getReq)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404 after delete, got %d", resp.StatusCode)
	}
}

// ── validateSessionRequest edge cases ────────────────────────────────────────

func TestCreateSessionHTMLInBrandReturns400(t *testing.T) {
	app, _ := newSessionTestApp(t)

	body := `{"device_brand":"<script>xss</script>","device_model":"MK4"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("HTML in brand: want 400, got %d", resp.StatusCode)
	}
}

func TestCreateSessionHTMLInModelReturns400(t *testing.T) {
	app, _ := newSessionTestApp(t)

	body := `{"device_brand":"Prusa","device_model":"<b>MK4</b>"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("HTML in model: want 400, got %d", resp.StatusCode)
	}
}

func TestCreateSessionBrandTooLongReturns400(t *testing.T) {
	app, _ := newSessionTestApp(t)

	body := `{"device_brand":"` + strings.Repeat("x", maxDeviceFieldLen+1) + `","device_model":"MK4"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("brand too long: want 400, got %d", resp.StatusCode)
	}
}

func TestCreateSessionModelTooLongReturns400(t *testing.T) {
	app, _ := newSessionTestApp(t)

	body := `{"device_brand":"Prusa","device_model":"` + strings.Repeat("y", maxDeviceFieldLen+1) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("model too long: want 400, got %d", resp.StatusCode)
	}
}

// ── ETag tests ────────────────────────────────────────────────────────────────

func TestGetSessionReturnsETagHeader(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Bambu", "X1C", "", "")

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Error("expected ETag header to be present in 200 response")
	}
	if !strings.HasPrefix(etag, `W/"`) {
		t.Errorf("expected weak ETag starting with W/\", got %q", etag)
	}
}

func TestGetSessionIfNoneMatchReturns304(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Bambu", "X1C", "", "")

	// First request — get the ETag.
	req1 := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID, nil)
	resp1, err := app.Test(req1)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	etag := resp1.Header.Get("ETag")
	if etag == "" {
		t.Fatal("first response missing ETag header")
	}

	// Second request with matching If-None-Match → 304.
	req2 := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID, nil)
	req2.Header.Set("If-None-Match", etag)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	if resp2.StatusCode != http.StatusNotModified {
		t.Errorf("want 304, got %d", resp2.StatusCode)
	}
}

func TestGetSessionIfNoneMatchMismatchReturns200(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Prusa", "MK4", "", "")

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID, nil)
	req.Header.Set("If-None-Match", `W/"stale-etag"`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200 on ETag mismatch, got %d", resp.StatusCode)
	}
}

func TestListSessionsReturnsETagHeader(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("Bambu", "X1C", "", "") //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Error("expected ETag header in ListSessions 200 response")
	}
}

func TestListSessionsIfNoneMatchReturns304(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("Bambu", "X1C", "", "") //nolint:errcheck

	// First request — get ETag.
	req1 := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	resp1, err := app.Test(req1)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	etag := resp1.Header.Get("ETag")
	if etag == "" {
		t.Fatal("first response missing ETag header")
	}

	// Second request with matching If-None-Match → 304.
	req2 := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req2.Header.Set("If-None-Match", etag)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	if resp2.StatusCode != http.StatusNotModified {
		t.Errorf("want 304, got %d", resp2.StatusCode)
	}
}

// ── Cache-Control tests ───────────────────────────────────────────────────────

func TestGetSessionCacheControlHeader(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Bambu", "X1C", "", "")

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	cc := resp.Header.Get("Cache-Control")
	if cc != "private, max-age=0, must-revalidate" {
		t.Errorf("GetSession Cache-Control: want %q, got %q", "private, max-age=0, must-revalidate", cc)
	}
}

func TestListSessionsCacheControlHeader(t *testing.T) {
	app, _ := newSessionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	cc := resp.Header.Get("Cache-Control")
	if cc != "private, no-cache" {
		t.Errorf("ListSessions Cache-Control: want %q, got %q", "private, no-cache", cc)
	}
}

func TestCreateSessionCacheControlHeader(t *testing.T) {
	app, _ := newSessionTestApp(t)

	body := `{"device_brand":"Prusa","device_model":"MK4"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("want 201, got %d", resp.StatusCode)
	}
	cc := resp.Header.Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("CreateSession Cache-Control: want %q, got %q", "no-store", cc)
	}
}

func TestDeleteSessionCacheControlHeader(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Prusa", "MK4", "", "")

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/"+sess.SessionID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("want 204, got %d", resp.StatusCode)
	}
	cc := resp.Header.Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("DeleteSession Cache-Control: want %q, got %q", "no-store", cc)
	}
}
