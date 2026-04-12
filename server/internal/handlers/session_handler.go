package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/services"
	"github.com/gofiber/fiber/v2"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
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

// sessionETag computes a weak ETag over a single sessionResponse by hashing
// the first 8 bytes of its SHA-256 JSON digest.
func sessionETag(s sessionResponse) string {
	b, _ := json.Marshal(s)
	h := sha256.Sum256(b)
	return fmt.Sprintf(`W/"%x"`, h[:8])
}

// sessionListETag computes a weak ETag over a slice of sessionResponse values.
func sessionListETag(dtos []sessionResponse) string {
	b, _ := json.Marshal(dtos)
	h := sha256.Sum256(b)
	return fmt.Sprintf(`W/"%x"`, h[:8])
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

	session, err := h.sessions.CreateSession(req.DeviceBrand, req.DeviceModel, "", "")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Cache-Control", "no-store")
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

// parsePagination parses the page and page_size query parameters from the
// request context and applies defaults and clamps. page defaults to 1 (min 1),
// page_size defaults to 20 (min 1, max 100).
func parsePagination(c *fiber.Ctx) (page, pageSize int) {
	page = c.QueryInt("page", defaultPage)
	if page < 1 {
		page = 1
	}
	pageSize = c.QueryInt("page_size", defaultPageSize)
	if pageSize < 1 {
		pageSize = 1
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// paginateSessions returns the sub-slice for the requested page together with
// the total count of entries before slicing.
func paginateSessions(dtos []sessionResponse, page, pageSize int) (paged []sessionResponse, total int) {
	total = len(dtos)
	start := (page - 1) * pageSize
	if start >= total {
		return []sessionResponse{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return dtos[start:end], total
}

// totalPages computes the number of pages required to hold total entries at
// pageSize items per page. Returns at least 1 when total is 0.
func totalPages(total, pageSize int) int {
	if total == 0 {
		return 1
	}
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	return pages
}

// sortSessions sorts the provided slice in-place according to sortBy and
// sortOrder. Supported sortBy values: "created_at", "last_activity", "status".
// sortOrder must be "asc" or "desc" (default "desc").
func sortSessions(dtos []sessionResponse, sortBy, sortOrder string) {
	if sortBy == "" {
		sortBy = "last_activity"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}
	asc := sortOrder == "asc"

	sort.Slice(dtos, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "created_at":
			less = dtos[i].CreatedAt.Before(dtos[j].CreatedAt)
		case "status":
			less = dtos[i].Status < dtos[j].Status
		default: // "last_activity"
			less = dtos[i].LastActivity.Before(dtos[j].LastActivity)
		}
		if asc {
			return less
		}
		return !less
	})
}

// ListSessions handles GET /api/sessions.
//
// Query parameters:
//   - page (int, default 1): 1-based page number
//   - page_size (int, default 20, max 100): items per page
//   - sort_by (string): "created_at" | "last_activity" | "status" (default "last_activity")
//   - sort_order (string): "asc" | "desc" (default "desc")
//
// Response 200: {"sessions":[...],"count":N,"page":P,"page_size":S,"total":T,"total_pages":TP} with ETag header.
// Response 304: when If-None-Match matches the current ETag.
func (h *SessionHandler) ListSessions(c *fiber.Ctx) error {
	all := h.sessions.ListSessions()
	dtos := make([]sessionResponse, len(all))
	for i, s := range all {
		dtos[i] = toResponse(s)
	}

	// 1. Sort
	sortSessions(dtos, c.Query("sort_by"), c.Query("sort_order"))

	etag := sessionListETag(dtos)
	if c.Get("If-None-Match") == etag {
		return c.SendStatus(fiber.StatusNotModified)
	}

	// 2. Paginate
	page, pageSize := parsePagination(c)
	paged, total := paginateSessions(dtos, page, pageSize)

	c.Set("ETag", etag)
	c.Set("Cache-Control", "private, no-cache")
	return c.JSON(fiber.Map{
		"sessions":    paged,
		"count":       len(paged),
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": totalPages(total, pageSize),
	})
}

// GetSession handles GET /api/sessions/:id.
//
// Response 200: session JSON with ETag and Cache-Control headers.
// Response 304: when If-None-Match matches the current ETag.
// Response 404: session not found.
func (h *SessionHandler) GetSession(c *fiber.Ctx) error {
	id := c.Params("id")
	session, err := h.sessions.GetSessionSnapshot(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	dto := toResponse(*session)
	etag := sessionETag(dto)
	if c.Get("If-None-Match") == etag {
		return c.SendStatus(fiber.StatusNotModified)
	}
	c.Set("ETag", etag)
	c.Set("Cache-Control", "private, max-age=0, must-revalidate")
	return c.JSON(dto)
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
	c.Set("Cache-Control", "no-store")
	return c.SendStatus(fiber.StatusNoContent)
}
