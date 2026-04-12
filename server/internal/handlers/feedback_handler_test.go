// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

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

// ── Validation edge cases ─────────────────────────────────────────────────────

func TestFeedbackSubmitEmptyRatingReturns400(t *testing.T) {
	// An empty rating string is not "positive" or "negative" → 400.
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-10","step_number":1,"rating":""}`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for empty rating, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitWhitespaceSessionIDReturns400(t *testing.T) {
	// session_id that is whitespace only must be rejected.
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"session_id":"   ","step_number":1,"rating":"positive"}`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for whitespace-only session_id, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitEmptyBodyReturns400(t *testing.T) {
	// Completely empty body should fail JSON parsing → 400.
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, ``)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for empty body, got %d", resp.StatusCode)
	}
}

func TestFeedbackSubmitErrorMessageContainsBadField(t *testing.T) {
	// When session_id is missing the 400 error body should mention it.
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"step_number":1,"rating":"positive"}`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), "session_id") {
		t.Errorf("expected error message to mention 'session_id', got: %s", body)
	}
}

func TestFeedbackSubmitNegativeRatingValidReturns202(t *testing.T) {
	// Confirm "negative" rating without comment is also accepted.
	db := &stubFeedbackDB{}
	app := newFeedbackTestApp(db)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-11","step_number":3,"rating":"negative"}`)

	if resp.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 202 for valid negative rating, got %d: %s", resp.StatusCode, b)
	}
	if len(db.saved) != 1 || db.saved[0].Rating != "negative" {
		t.Errorf("expected saved entry with rating 'negative'")
	}
}

func TestFeedbackSubmitStepNumberZeroIsValid(t *testing.T) {
	// step_number=0 is a valid value (first step); must return 202.
	db := &stubFeedbackDB{}
	app := newFeedbackTestApp(db)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-12","step_number":0,"rating":"positive"}`)

	if resp.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 202 for step_number=0, got %d: %s", resp.StatusCode, b)
	}
	if len(db.saved) != 1 || db.saved[0].StepNumber != 0 {
		t.Errorf("expected saved entry with step_number 0")
	}
}

func TestFeedbackSubmitCommentAtExactMaxLengthReturns202(t *testing.T) {
	// A comment of exactly 1000 chars must be accepted (boundary condition).
	db := &stubFeedbackDB{}
	app := newFeedbackTestApp(db)

	exactComment := strings.Repeat("x", 1000)
	body, _ := json.Marshal(map[string]interface{}{
		"session_id":  "sess-13",
		"step_number": 1,
		"rating":      "positive",
		"comment":     exactComment,
	})

	resp := doFeedbackPost(t, app, string(body))

	if resp.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 202 for comment at max length, got %d: %s", resp.StatusCode, b)
	}
}

func TestFeedbackSubmitHTMLInSessionIDErrorMessageIsInformative(t *testing.T) {
	// The 400 response for HTML in session_id must contain a descriptive message.
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"session_id":"<img>","step_number":1,"rating":"positive"}`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), "HTML") && !strings.Contains(string(body), "session_id") {
		t.Errorf("expected error message to mention HTML or session_id, got: %s", body)
	}
}
