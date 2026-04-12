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
	app.Get("/api/sessions/:id/steps", h.GetSessionSteps)
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

// ── Pagination tests ──────────────────────────────────────────────────────────

func TestListSessionsPaginationPage2(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	for i := 0; i < 5; i++ {
		sessions.CreateSession("Brand", "Model", "", "") //nolint:errcheck
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?page=2&page_size=2", nil)
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

	sessionsList, ok := body["sessions"].([]interface{})
	if !ok {
		t.Fatalf("sessions field missing or wrong type: %v", body["sessions"])
	}
	if len(sessionsList) != 2 {
		t.Errorf("want 2 sessions on page 2, got %d", len(sessionsList))
	}
	if int(body["page"].(float64)) != 2 {
		t.Errorf("want page=2, got %v", body["page"])
	}
	if int(body["page_size"].(float64)) != 2 {
		t.Errorf("want page_size=2, got %v", body["page_size"])
	}
	if int(body["total"].(float64)) != 5 {
		t.Errorf("want total=5, got %v", body["total"])
	}
}

func TestListSessionsPaginationDefaults(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("Brand", "Model", "", "") //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck

	if int(body["page"].(float64)) != 1 {
		t.Errorf("default page should be 1, got %v", body["page"])
	}
	if int(body["page_size"].(float64)) != 20 {
		t.Errorf("default page_size should be 20, got %v", body["page_size"])
	}
}

func TestListSessionsPaginationBeyondEndReturnsEmpty(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("Brand", "Model", "", "") //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?page=99&page_size=10", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck

	sessionsList, _ := body["sessions"].([]interface{})
	if len(sessionsList) != 0 {
		t.Errorf("want 0 sessions beyond last page, got %d", len(sessionsList))
	}
}

// ── Sorting tests ─────────────────────────────────────────────────────────────

func TestListSessionsSortByCreatedAtAsc(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("Alpha", "M1", "", "") //nolint:errcheck
	sessions.CreateSession("Beta", "M2", "", "")  //nolint:errcheck
	sessions.CreateSession("Gamma", "M3", "", "") //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?sort_by=created_at&sort_order=asc", nil)
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

	sessionsList, ok := body["sessions"].([]interface{})
	if !ok || len(sessionsList) != 3 {
		t.Fatalf("expected 3 sessions, got %v", body["sessions"])
	}

	// Verify ordering: each created_at must be <= the next one.
	for i := 0; i < len(sessionsList)-1; i++ {
		curr := sessionsList[i].(map[string]interface{})
		next := sessionsList[i+1].(map[string]interface{})
		currTime := curr["created_at"].(string)
		nextTime := next["created_at"].(string)
		if currTime > nextTime {
			t.Errorf("sessions not sorted asc by created_at: index %d (%s) > index %d (%s)",
				i, currTime, i+1, nextTime)
		}
	}
}

func TestListSessionsSortByLastActivityDesc(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("X", "M1", "", "") //nolint:errcheck
	sessions.CreateSession("Y", "M2", "", "") //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?sort_by=last_activity&sort_order=desc", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck

	sessionsList, ok := body["sessions"].([]interface{})
	if !ok || len(sessionsList) != 2 {
		t.Fatalf("expected 2 sessions, got %v", body["sessions"])
	}

	for i := 0; i < len(sessionsList)-1; i++ {
		curr := sessionsList[i].(map[string]interface{})
		next := sessionsList[i+1].(map[string]interface{})
		currTime := curr["last_activity"].(string)
		nextTime := next["last_activity"].(string)
		if currTime < nextTime {
			t.Errorf("sessions not sorted desc by last_activity at index %d", i)
		}
	}
}

// ── Filtering tests ───────────────────────────────────────────────────────────

func TestListSessionsFilterByDeviceBrand(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("Prusa", "MK4", "", "")  //nolint:errcheck
	sessions.CreateSession("Bambu", "X1C", "", "")  //nolint:errcheck
	sessions.CreateSession("Prusa", "Mini", "", "") //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?device_brand=Prusa", nil)
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

	sessionsList, ok := body["sessions"].([]interface{})
	if !ok {
		t.Fatalf("sessions field missing: %v", body)
	}
	if len(sessionsList) != 2 {
		t.Errorf("want 2 Prusa sessions, got %d", len(sessionsList))
	}
	for _, s := range sessionsList {
		sess := s.(map[string]interface{})
		device := sess["device"].(map[string]interface{})
		if strings.ToLower(device["brand"].(string)) != "prusa" {
			t.Errorf("expected brand 'Prusa', got %v", device["brand"])
		}
	}
}

func TestListSessionsFilterByDeviceBrandCaseInsensitive(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("Prusa", "MK4", "", "") //nolint:errcheck
	sessions.CreateSession("Bambu", "X1C", "", "") //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?device_brand=prusa", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck

	sessionsList, _ := body["sessions"].([]interface{})
	if len(sessionsList) != 1 {
		t.Errorf("case-insensitive filter: want 1 session, got %d", len(sessionsList))
	}
}

func TestListSessionsFilterByProblem(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	s1, _ := sessions.CreateSession("Prusa", "MK4", "", "")
	s2, _ := sessions.CreateSession("Bambu", "X1C", "", "")
	sessions.SetProblemDetected(s1.SessionID, "nozzle clog detected") //nolint:errcheck
	sessions.SetProblemDetected(s2.SessionID, "bed adhesion issue")   //nolint:errcheck
	sessions.CreateSession("Creality", "Ender3", "", "")              //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?problem=clog", nil)
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

	sessionsList, ok := body["sessions"].([]interface{})
	if !ok {
		t.Fatalf("sessions field missing: %v", body)
	}
	if len(sessionsList) != 1 {
		t.Errorf("want 1 session with clog, got %d", len(sessionsList))
	}
	sess := sessionsList[0].(map[string]interface{})
	if !strings.Contains(strings.ToLower(sess["problem_detected"].(string)), "clog") {
		t.Errorf("expected problem_detected to contain 'clog', got %v", sess["problem_detected"])
	}
}

func TestListSessionsFilterByStatus(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sessions.CreateSession("Prusa", "MK4", "", "") //nolint:errcheck
	sessions.CreateSession("Bambu", "X1C", "", "") //nolint:errcheck

	// All current sessions have status "active"
	req := httptest.NewRequest(http.MethodGet, "/api/sessions?status=active", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck

	sessionsList, _ := body["sessions"].([]interface{})
	if len(sessionsList) != 2 {
		t.Errorf("want 2 active sessions, got %d", len(sessionsList))
	}

	// Non-matching status returns empty
	req2 := httptest.NewRequest(http.MethodGet, "/api/sessions?status=expired", nil)
	resp2, _ := app.Test(req2)
	var body2 map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&body2) //nolint:errcheck
	sessionsList2, _ := body2["sessions"].([]interface{})
	if len(sessionsList2) != 0 {
		t.Errorf("want 0 expired sessions, got %d", len(sessionsList2))
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

// ── GetSessionSteps tests ─────────────────────────────────────────────────────

func TestGetSessionStepsReturns200WithCorrectFields(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Bambu", "X1C", "", "")
	sessions.UpdateStep(sess.SessionID, 2, 5) //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID+"/steps", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 200, got %d: %s", resp.StatusCode, b)
	}

	var dto sessionStepsResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if dto.SessionID != sess.SessionID {
		t.Errorf("session_id: want %q, got %q", sess.SessionID, dto.SessionID)
	}
	if dto.CurrentStep != 2 {
		t.Errorf("current_step: want 2, got %d", dto.CurrentStep)
	}
	if dto.TotalSteps != 5 {
		t.Errorf("total_steps: want 5, got %d", dto.TotalSteps)
	}
	if dto.ProgressPct != 40.0 {
		t.Errorf("progress_pct: want 40.0, got %f", dto.ProgressPct)
	}
}

func TestGetSessionStepsNotFoundReturns404(t *testing.T) {
	app, _ := newSessionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/nonexistent-id/steps", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestGetSessionStepsZeroTotalStepsProgressIsZero(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Prusa", "MK4", "", "")

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID+"/steps", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	var dto sessionStepsResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if dto.ProgressPct != 0.0 {
		t.Errorf("progress_pct with 0 total steps: want 0.0, got %f", dto.ProgressPct)
	}
}

func TestGetSessionStepsCacheControlHeader(t *testing.T) {
	app, sessions := newSessionTestApp(t)
	sess, _ := sessions.CreateSession("Bambu", "X1C", "", "")

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sess.SessionID+"/steps", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	cc := resp.Header.Get("Cache-Control")
	if cc != "private, no-cache" {
		t.Errorf("Cache-Control: want %q, got %q", "private, no-cache", cc)
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
