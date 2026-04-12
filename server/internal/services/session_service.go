package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/metrics"
	"github.com/google/uuid"
)

// maxConversationHistory is the rolling window size for conversation messages.
const maxConversationHistory = 20

// maxContextMessages is how many recent messages BuildContextForGemini includes.
const maxContextMessages = 10

// DeviceInfo holds identifying information about the device being repaired.
type DeviceInfo struct {
	Brand string `json:"brand"`
	Model string `json:"model"`
}

// ConversationMessage represents a single message in the repair conversation.
type ConversationMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	HasImage  bool      `json:"has_image"`
}

// SessionState holds the in-memory state of an active repair session.
type SessionState struct {
	SessionID           string                `json:"session_id"`
	DeviceInfo          DeviceInfo            `json:"device_info"`
	Language            string                `json:"language,omitempty"` // BCP-47 code; defaults to "it"
	ConversationHistory []ConversationMessage `json:"conversation_history"`
	CurrentStep         int                   `json:"current_step"`
	TotalSteps          int                   `json:"total_steps"`
	ProblemDetected     string                `json:"problem_detected,omitempty"`
	CreatedAt           time.Time             `json:"created_at"`
	LastActivity        time.Time             `json:"last_activity"`
}

// SessionService manages in-memory repair sessions with thread-safe access.
// When a DBBackend is attached via SetDB the service performs best-effort
// write-through to Postgres; the in-memory store remains authoritative so
// the server keeps working even when the DB is temporarily unreachable.
type SessionService struct {
	mu       sync.RWMutex
	sessions map[string]*SessionState
	db       DBBackend // nil when DATABASE_URL is not configured
}

// NewSessionService creates a SessionService with an initialized in-memory store.
func NewSessionService() *SessionService {
	return &SessionService{
		sessions: make(map[string]*SessionState),
	}
}

// SetDB attaches an optional Postgres backend for write-through persistence.
// Call this once during startup if DATABASE_URL is configured.
func (s *SessionService) SetDB(db DBBackend) {
	s.db = db
}

// HydrateFromDB loads all active sessions from the DB backend into the
// in-memory store. Call once at startup (after SetDB) to recover sessions
// that were active before the last server restart.
// Is a no-op when no DB backend is attached.
func (s *SessionService) HydrateFromDB(ctx context.Context) error {
	if s.db == nil {
		return nil
	}
	sessions, err := s.db.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("hydrate sessions from DB: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range sessions {
		sess := sessions[i]
		if sess.ConversationHistory == nil {
			sess.ConversationHistory = make([]ConversationMessage, 0)
		}
		s.sessions[sess.SessionID] = &sess
	}
	slog.Info("hydrated sessions from DB", "count", len(sessions))
	return nil
}

// ListSessions returns a snapshot of every in-memory session, ordered
// arbitrarily. Each entry is a shallow copy safe for concurrent reads.
func (s *SessionService) ListSessions() []SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]SessionState, 0, len(s.sessions))
	for _, session := range s.sessions {
		result = append(result, *session)
	}
	return result
}

// CreateSession initializes a new repair session for the given device.
// language is a BCP-47 code (e.g. "it", "en") controlling the AI response language;
// defaults to "it" when empty.
// If a DB backend is attached, the session is also persisted to Postgres
// (best-effort: a DB error is logged but does not fail the in-memory create).
func (s *SessionService) CreateSession(deviceBrand, deviceModel, language string) (*SessionState, error) {
	if language == "" {
		language = "it"
	}
	now := time.Now()
	session := &SessionState{
		SessionID: uuid.New().String(),
		DeviceInfo: DeviceInfo{
			Brand: deviceBrand,
			Model: deviceModel,
		},
		Language:            language,
		ConversationHistory: make([]ConversationMessage, 0),
		CreatedAt:           now,
		LastActivity:        now,
	}

	s.mu.Lock()
	s.sessions[session.SessionID] = session
	s.mu.Unlock()

	if s.db != nil {
		if err := s.db.SaveSession(context.Background(), session); err != nil {
			slog.Warn("failed to persist session to DB", "session_id", session.SessionID, "err", err)
		}
	}

	return session, nil
}

// GetSession retrieves a session by ID. Returns an error if not found.
func (s *SessionService) GetSession(sessionID string) (*SessionState, error) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session %q not found", sessionID)
	}

	return session, nil
}

// GetSessionSnapshot returns a shallow copy of the session under lock, safe
// for concurrent reads. Callers receive an independent value that will not
// race with concurrent writes (e.g. UpdateStep). The returned pointer must
// not be written back into the store.
func (s *SessionService) GetSessionSnapshot(sessionID string) (*SessionState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %q not found", sessionID)
	}

	snapshot := *session
	return &snapshot, nil
}

// AddMessage appends a conversation message to the session, keeping only the
// most recent maxConversationHistory messages (rolling window).
func (s *SessionService) AddMessage(sessionID, role, content string, hasImage bool) error {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %q not found", sessionID)
	}

	msg := ConversationMessage{
		Role:      role,
		Content:   content,
		Timestamp: now,
		HasImage:  hasImage,
	}

	session.ConversationHistory = append(session.ConversationHistory, msg)

	// Trim to rolling window
	if maxConversationHistory > 0 && len(session.ConversationHistory) > maxConversationHistory {
		session.ConversationHistory = session.ConversationHistory[len(session.ConversationHistory)-maxConversationHistory:]
	}

	session.LastActivity = now

	return nil
}

// UpdateStep sets the current and total step count for a guided repair flow.
func (s *SessionService) UpdateStep(sessionID string, current, total int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %q not found", sessionID)
	}

	session.CurrentStep = current
	session.TotalSteps = total
	session.LastActivity = time.Now()

	return nil
}

// SetProblemDetected records the identified problem for a session.
func (s *SessionService) SetProblemDetected(sessionID, problem string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %q not found", sessionID)
	}

	session.ProblemDetected = problem
	session.LastActivity = time.Now()

	return nil
}

// BuildContextForGemini formats the last maxContextMessages messages as
// "role: content" lines separated by newlines, suitable for an LLM prompt.
// All slice processing happens under the read lock to prevent data races
// when AddMessage appends concurrently without reallocating the backing array.
func (s *SessionService) BuildContextForGemini(sessionID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return "", fmt.Errorf("session %q not found", sessionID)
	}

	history := session.ConversationHistory
	if len(history) == 0 {
		return "", nil
	}

	// Take only the last maxContextMessages
	start := 0
	if len(history) > maxContextMessages {
		start = len(history) - maxContextMessages
	}

	lines := make([]string, 0, len(history)-start)
	for _, msg := range history[start:] {
		lines = append(lines, msg.Role+": "+msg.Content)
	}

	return strings.Join(lines, "\n"), nil
}

// DeleteSession removes a session from the in-memory store. If a DB backend is
// attached it also soft-deletes the row in Postgres (best-effort).
// Returns an error only when the session does not exist in memory.
func (s *SessionService) DeleteSession(sessionID string) error {
	s.mu.Lock()
	if _, exists := s.sessions[sessionID]; !exists {
		s.mu.Unlock()
		return fmt.Errorf("session %q not found", sessionID)
	}
	delete(s.sessions, sessionID)
	s.mu.Unlock()

	if s.db != nil {
		if err := s.db.DeleteSession(context.Background(), sessionID); err != nil {
			slog.Warn("failed to delete session from DB", "session_id", sessionID, "err", err)
		}
	}
	return nil
}

// CleanupExpiredSessions removes all sessions whose LastActivity is older
// than maxAge from the in-memory store and marks them as 'expired' in Postgres
// (best-effort; a DB error is logged but does not abort the cleanup).
// Returns the number of sessions that were cleaned up.
// Designed to be called periodically in a background goroutine.
func (s *SessionService) CleanupExpiredSessions(maxAge time.Duration) int {
	cutoff := time.Now().Add(-maxAge)

	s.mu.Lock()
	var expired []string
	for id, session := range s.sessions {
		if session.LastActivity.Before(cutoff) {
			delete(s.sessions, id)
			expired = append(expired, id)
		}
	}
	s.mu.Unlock()

	if len(expired) == 0 {
		return 0
	}

	metrics.SessionsExpiredTotal.Add(float64(len(expired)))

	if s.db != nil {
		for _, id := range expired {
			if err := s.db.ExpireSession(context.Background(), id); err != nil {
				slog.Warn("failed to expire session in DB", "session_id", id, "err", err)
			}
		}
	}

	return len(expired)
}
