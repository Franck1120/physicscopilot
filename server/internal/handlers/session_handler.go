package handlers

import (
	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/fiber/v2"
)

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

// CreateSession handles POST /api/sessions.
//
// Body (JSON): {"device_brand": "Prusa", "device_model": "MK4"}
// Response 201: full SessionState JSON.
// Response 400: malformed body.
func (h *SessionHandler) CreateSession(c *fiber.Ctx) error {
	var req createSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}

	session, err := h.sessions.CreateSession(req.DeviceBrand, req.DeviceModel)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

// ListSessions handles GET /api/sessions.
//
// Response 200: {"sessions": [...], "count": N}
func (h *SessionHandler) ListSessions(c *fiber.Ctx) error {
	all := h.sessions.ListSessions()
	return c.JSON(fiber.Map{
		"sessions": all,
		"count":    len(all),
	})
}

// GetSession handles GET /api/sessions/:id.
//
// Response 200: SessionState JSON including conversation history.
// Response 404: session not found.
func (h *SessionHandler) GetSession(c *fiber.Ctx) error {
	id := c.Params("id")
	session, err := h.sessions.GetSessionSnapshot(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(session)
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
