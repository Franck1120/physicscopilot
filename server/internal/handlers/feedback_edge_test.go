package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestFeedbackSubmitZeroRatingFieldReturns400 verifies that a payload where
// "rating" is the zero value for a string (empty string) is rejected with 400.
// This exercises the boundary: an omitted rating vs an explicitly empty one.
func TestFeedbackSubmitZeroRatingFieldReturns400(t *testing.T) {
	t.Parallel()
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-edge-1","step_number":0,"rating":""}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty-string rating: want 400, got %d", resp.StatusCode)
	}
}

// TestFeedbackSubmitEmptyCommentFieldIsAccepted verifies that an explicitly
// empty comment string (not omitted but "comment":"") is accepted — zero-length
// is within the 1000-char limit.
func TestFeedbackSubmitEmptyCommentFieldIsAccepted(t *testing.T) {
	t.Parallel()
	db := &stubFeedbackDB{}
	app := newFeedbackTestApp(db)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-edge-2","step_number":1,"rating":"positive","comment":""}`)
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("empty comment string: want 202, got %d", resp.StatusCode)
	}
}

// TestFeedbackSubmitMaxLengthCommentAccepted verifies that a comment of exactly
// 1000 characters (the boundary value) is accepted with 202.
func TestFeedbackSubmitMaxLengthCommentAccepted(t *testing.T) {
	t.Parallel()
	db := &stubFeedbackDB{}
	app := newFeedbackTestApp(db)

	exactComment := strings.Repeat("z", 1000)
	body, _ := json.Marshal(map[string]interface{}{
		"session_id":  "sess-edge-3",
		"step_number": 2,
		"rating":      "negative",
		"comment":     exactComment,
	})

	resp := doFeedbackPost(t, app, string(body))
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("1000-char comment: want 202, got %d", resp.StatusCode)
	}
	if len(db.saved) != 1 {
		t.Fatalf("expected 1 saved entry, got %d", len(db.saved))
	}
	if db.saved[0].Comment == nil || len(*db.saved[0].Comment) != 1000 {
		t.Errorf("comment not saved at full length")
	}
}

// TestFeedbackSubmitUnicodeCommentAccepted verifies that a comment containing
// multi-byte Unicode characters (including emoji and CJK) is accepted as long
// as its character count is within limits.
func TestFeedbackSubmitUnicodeCommentAccepted(t *testing.T) {
	t.Parallel()
	db := &stubFeedbackDB{}
	app := newFeedbackTestApp(db)

	// Mix of ASCII, accented Latin, CJK, and emoji — all valid text.
	unicodeComment := "Bon travail 👍 — 素晴らしい分析 — très utile!"
	body, _ := json.Marshal(map[string]interface{}{
		"session_id":  "sess-edge-4",
		"step_number": 1,
		"rating":      "positive",
		"comment":     unicodeComment,
	})

	resp := doFeedbackPost(t, app, string(body))
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("unicode comment: want 202, got %d", resp.StatusCode)
	}
	if len(db.saved) != 1 {
		t.Fatalf("expected 1 saved entry, got %d", len(db.saved))
	}
	if db.saved[0].Comment == nil || *db.saved[0].Comment != unicodeComment {
		t.Errorf("unicode comment not saved correctly")
	}
}

// TestFeedbackSubmitTwiceSameSessionIDBothAccepted verifies that submitting
// two feedback entries for the same session ID is allowed — the handler does
// not enforce uniqueness at the HTTP layer.
func TestFeedbackSubmitTwiceSameSessionIDBothAccepted(t *testing.T) {
	t.Parallel()
	db := &stubFeedbackDB{}
	app := newFeedbackTestApp(db)

	const sessID = "sess-edge-dup"
	body1 := `{"session_id":"` + sessID + `","step_number":1,"rating":"positive"}`
	body2 := `{"session_id":"` + sessID + `","step_number":2,"rating":"negative","comment":"changed mind"}`

	resp1 := doFeedbackPost(t, app, body1)
	if resp1.StatusCode != http.StatusAccepted {
		t.Errorf("first submit same session: want 202, got %d", resp1.StatusCode)
	}

	resp2 := doFeedbackPost(t, app, body2)
	if resp2.StatusCode != http.StatusAccepted {
		t.Errorf("second submit same session: want 202, got %d", resp2.StatusCode)
	}

	if len(db.saved) != 2 {
		t.Errorf("expected 2 saved entries for duplicate session ID, got %d", len(db.saved))
	}
}

// TestFeedbackSubmitResponseContentTypeIsJSON verifies that the 202 response
// carries "application/json" Content-Type regardless of the input payload.
func TestFeedbackSubmitResponseContentTypeIsJSON(t *testing.T) {
	t.Parallel()
	app := newFeedbackTestApp(nil)

	resp := doFeedbackPost(t, app, `{"session_id":"sess-edge-5","step_number":0,"rating":"positive"}`)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("want 202, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: want application/json, got %q", ct)
	}
}
