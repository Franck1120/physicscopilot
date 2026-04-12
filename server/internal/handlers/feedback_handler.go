package handlers

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/metrics"
	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/fiber/v2"
)

// feedbackDB is the subset of services.DBBackend used by FeedbackHandler.
// Using a narrow interface keeps the handler testable without a real DB.
type feedbackDB interface {
	SaveFeedback(ctx context.Context, f *services.FeedbackEntry) error
}

// FeedbackHandler exposes POST /api/feedback.
type FeedbackHandler struct {
	db feedbackDB // nil when DATABASE_URL is not configured
}

// NewFeedbackHandler returns a FeedbackHandler. db may be nil; in that case
// feedback is only logged (no persistence).
func NewFeedbackHandler(db feedbackDB) *FeedbackHandler {
	return &FeedbackHandler{db: db}
}

// feedbackRequest is the expected JSON body for POST /api/feedback.
type feedbackRequest struct {
	SessionID  string  `json:"session_id"`
	StepNumber int     `json:"step_number"`
	Rating     string  `json:"rating"`
	Comment    *string `json:"comment,omitempty"`
}

// Submit handles POST /api/feedback.
//
// Body (JSON):
//
//	{"session_id":"<uuid>", "step_number":1, "rating":"positive", "comment":"optional"}
//
// Responses:
//
//	202 Accepted  — feedback recorded (persisted or logged)
//	400 Bad Request — missing/invalid fields
func (h *FeedbackHandler) Submit(c *fiber.Ctx) error {
	var req feedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}

	if err := validateFeedbackRequest(req); err != nil {
		return err
	}

	entry := &services.FeedbackEntry{
		SessionID:  req.SessionID,
		StepNumber: req.StepNumber,
		Rating:     req.Rating,
		Comment:    req.Comment,
	}

	if h.db != nil {
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()
		if err := h.db.SaveFeedback(ctx, entry); err != nil {
			// Non-fatal: track and still return 202 so the client is not
			// blocked by a transient DB failure.
			metrics.TrackError(metrics.CategoryDB, err, "session_id", req.SessionID)
		}
	} else {
		slog.Info("feedback received (no DB configured)",
			"session_id", req.SessionID,
			"step", req.StepNumber,
			"rating", req.Rating,
		)
	}

	metrics.FeedbackTotal.WithLabelValues(req.Rating).Inc()
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "ok"})
}

// validateFeedbackRequest enforces input constraints:
//   - session_id: non-empty, no HTML chars
//   - rating: must be "positive" or "negative"
//   - comment: if present, max 1000 chars
func validateFeedbackRequest(req feedbackRequest) error {
	if strings.TrimSpace(req.SessionID) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "session_id is required")
	}
	if strings.ContainsAny(req.SessionID, "<>") {
		return fiber.NewError(fiber.StatusBadRequest, "session_id must not contain HTML characters")
	}
	if req.Rating != "positive" && req.Rating != "negative" {
		return fiber.NewError(fiber.StatusBadRequest, `rating must be "positive" or "negative"`)
	}
	if req.Comment != nil && len(*req.Comment) > 1000 {
		return fiber.NewError(fiber.StatusBadRequest, "comment exceeds maximum length of 1000 characters")
	}
	return nil
}
