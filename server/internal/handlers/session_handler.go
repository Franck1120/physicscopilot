package handlers

import (
	"strings"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/fiber/v2"
)

const maxDeviceFieldLen = 100

// SessionHandler exposes REST endpoints for session lifecycle management.
// Sessions are stored in-memory via SessionService; they are not persisted
// to Supabase in this implementation.
type SessionHandler struct {
	sessions *services.SessionService
}

// NewSessionHandler returns a SessionHandler wired to the given session store.
func NewSessionHandler(sessions *services.SessionService) *SessionHandler {
	return &SessionHandler{sessions: sessions}
}

// createSessionRequest is the JSON body expected by POST /api/sessions.
type createSessionRequest struct {
	DeviceBrand string `json:"device_brand"`
	DeviceModel string `json:"device_model"`
}

// sessionDevice mirrors the device sub-object in the REST response.
type sessionDevice struct {
	Brand string `json:"brand"`
	Model string `json:"model"`
}

// sessionResponse is the API DTO for a session. Field names match the JSON keys
// expected by the Flutter client (Session.fromJson in lib/models/session.dart).
// The internal services.SessionState uses different field names, so we map here.
type sessionResponse struct {
	ID              string        `json:"id"`
	Status          string        `json:"status"`
	Device          sessionDevice `json:"device"`
	ProblemDetected string        `json:"problem_detected,omitempty"`
	CurrentStep     int           `json:"current_step"`
	TotalSteps      int           `json:"total_steps"`
	CreatedAt       time.Time     `json:"created_at"`
	LastActivity    time.Time     `json:"last_activity"`
}

// toResponse converts an internal SessionState to the public DTO.
func toResponse(s services.SessionState) sessionResponse {
	return sessionResponse{
		ID:     s.SessionID,
		Status: "active",
		Device: sessionDevice{
			Brand: s.DeviceInfo.Brand,
			Model: s.DeviceInfo.Model,
		},
		ProblemDetected: s.ProblemDetected,
		CurrentStep:     s.CurrentStep,
		TotalSteps:      s.TotalSteps,
		CreatedAt:       s.CreatedAt,
		LastActivity:    s.LastActivity,
	}
}

// CreateSession handles POST /api/sessions.
//
// Body (JSON): {"device_brand": "Prusa", "device_model": "MK4"}
// Response 201: full SessionState JSON.
// Response 400: malformed body or invalid field values.
func (h *SessionHandler) CreateSession(c *fiber.Ctx) error {
	var req createSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}

	if err := validateSessionRequest(req); err != nil {
		return err
	}

	session, err := h.sessions.CreateSession(req.DeviceBrand, req.DeviceModel, "")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(toResponse(*session))
}

// validateSessionRequest enforces input constraints on device fields:
//   - max length: 100 chars each
//   - no HTML/script injection: reject values containing '<' or '>'
func validateSessionRequest(req createSessionRequest) error {
	if len(req.DeviceBrand) > maxDeviceFieldLen {
		return fiber.NewError(fiber.StatusBadRequest,
			"device_brand exceeds maximum length of 100 characters")
	}
	if len(req.DeviceModel) > maxDeviceFieldLen {
		return fiber.NewError(fiber.StatusBadRequest,
			"device_model exceeds maximum length of 100 characters")
	}
	if strings.ContainsAny(req.DeviceBrand, "<>") || strings.ContainsAny(req.DeviceModel, "<>") {
		return fiber.NewError(fiber.StatusBadRequest,
			"device fields must not contain HTML characters")
	}
	return nil
}

// ListSessions handles GET /api/sessions.
//
// Response 200: {"sessions": [...], "count": N}
func (h *SessionHandler) ListSessions(c *fiber.Ctx) error {
	all := h.sessions.ListSessions()
	dtos := make([]sessionResponse, len(all))
	for i, s := range all {
		dtos[i] = toResponse(s)
	}
	return c.JSON(fiber.Map{
		"sessions": dtos,
		"count":    len(dtos),
	})
}

// GetSession handles GET /api/sessions/:id.
//
// Response 200: session JSON.
// Response 404: session not found.
func (h *SessionHandler) GetSession(c *fiber.Ctx) error {
	id := c.Params("id")
	session, err := h.sessions.GetSessionSnapshot(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(toResponse(*session))
}

// DeleteSession handles DELETE /api/sessions/:id.
//
// Response 204: session deleted.
// Response 404: session not found.
func (h *SessionHandler) DeleteSession(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.sessions.DeleteSession(id); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
