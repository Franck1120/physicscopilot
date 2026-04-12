package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// stub DB
// ---------------------------------------------------------------------------

type stubFeedbackDB struct {
	saved []*services.FeedbackEntry
	err   error // returned by SaveFeedback when non-nil
}

func (s *stubFeedbackDB) SaveFeedback(_ context.Context, f *services.FeedbackEntry) error {
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, f)
	return nil
}

// ---------------------------------------------------------------------------
// test helpers
// ---------------------------------------------------------------------------

func newFeedbackTestApp(db feedbackDB) *fiber.App {
	h := NewFeedbackHandler(db)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/api/feedback", h.Submit)
	return app
}

func doFeedbackPost(t *testing.T, app *fiber.App, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	return resp
}

// ---------------------------------------------------------------------------
// tests
// ---------------------------------------------------------------------------

func TestFeedbackSubmitPositiveReturns202(t *testing.T) {
	db := &stubFeedbackDB{}
	app := newFeedbackTestApp(db)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-1","step_number":1,"rating":"positive"}`)

	if resp.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 202, got %d: %s", resp.StatusCode, b)
	}
	if len(db.saved) != 1 {
		t.Fatalf("expected 1 saved entry, got %d", len(db.saved))
	}
	if db.saved[0].Rating != "positive" {
		t.Errorf("want rating 'positive', got %q", db.saved[0].Rating)
	}
}

func TestFeedbackSubmitNegativeWithCommentReturns202(t *testing.T) {
	db := &stubFeedbackDB{}
	app := newFeedbackTestApp(db)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-2","step_number":0,"rating":"negative","comment":"too fast"}`)

	if resp.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 202, got %d: %s", resp.StatusCode, b)
	}
	if db.saved[0].Comment == nil || *db.saved[0].Comment != "too fast" {
		t.Errorf("comment not saved correctly")
	}
}

func TestFeedbackSubmitNilDBReturns202(t *testing.T) {
	// No DB configured — feedback should be logged and still return 202.
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-3","step_number":2,"rating":"positive"}`)

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("want 202, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitDBErrorReturns202(t *testing.T) {
	// DB error is non-fatal; handler still returns 202.
	db := &stubFeedbackDB{err: errors.New("connection reset")}
	app := newFeedbackTestApp(db)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-4","step_number":1,"rating":"negative"}`)

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("want 202 even on DB error, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitMissingSessionIDReturns400(t *testing.T) {
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"step_number":1,"rating":"positive"}`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for missing session_id, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitInvalidRatingReturns400(t *testing.T) {
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-5","step_number":1,"rating":"meh"}`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for invalid rating, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitHTMLInSessionIDReturns400(t *testing.T) {
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"session_id":"<script>","step_number":1,"rating":"positive"}`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for HTML in session_id, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitCommentTooLongReturns400(t *testing.T) {
	app := newFeedbackTestApp(nil)

	longComment := strings.Repeat("a", 1001)
	body, _ := json.Marshal(map[string]interface{}{
		"session_id":  "sess-6",
		"step_number": 1,
		"rating":      "positive",
		"comment":     longComment,
	})

	resp := doFeedbackPost(t, app, string(body))

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for comment too long, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitBadJSONReturns400(t *testing.T) {
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `not-json{{`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for malformed JSON, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitResponseBodyIsJSON(t *testing.T) {
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-7","step_number":1,"rating":"positive"}`)

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("want status 'ok', got %v", body["status"])
	}
}
