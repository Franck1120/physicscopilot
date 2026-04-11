package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Franck1120/physicscopilot/server/internal/db"
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
	ConversationHistory []ConversationMessage `json:"conversation_history"`
	CurrentStep         int                   `json:"current_step"`
	TotalSteps          int                   `json:"total_steps"`
	ProblemDetected     string                `json:"problem_detected,omitempty"`
	CreatedAt           time.Time             `json:"created_at"`
	LastActivity        time.Time             `json:"last_activity"`
}

// SessionService manages repair sessions with thread-safe in-memory cache
// and optional PostgreSQL persistence. When sessionRepo and messageRepo are
// nil, the service operates in-memory only (useful for tests).
type SessionService struct {
	mu          sync.RWMutex
	sessions    map[string]*SessionState
	sessionRepo *db.SessionRepo
	messageRepo *db.MessageRepo
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewSessionService creates a SessionService.
// If sessionRepo and messageRepo are nil, the service operates in-memory only (e.g., for tests).
func NewSessionService(sessionRepo *db.SessionRepo, messageRepo *db.MessageRepo) *SessionService {
	ctx, cancel := context.WithCancel(context.Background())
	return &SessionService{
		sessions:    make(map[string]*SessionState),
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Shutdown cancels in-flight background DB writes and waits for them to complete.
// Call this before closing the database pool.
func (s *SessionService) Shutdown() {
	s.cancel()
	s.wg.Wait()
}

// CreateSession initializes a new repair session for the given device.
// When a DB repo is available, the session is persisted first and gets a
// Postgres-generated UUID. Otherwise, a local UUID is generated.
func (s *SessionService) CreateSession(deviceBrand, deviceModel string) (*SessionState, error) {
	now := time.Now()

	var sessionID string

	// Persist to DB when repo is available; the DB generates the UUID.
	if s.sessionRepo != nil {
		rec, err := s.sessionRepo.CreateSession(s.ctx, "anonymous", deviceBrand, deviceModel)
		if err != nil {
			slog.Warn("db: failed to persist new session, falling back to in-memory", "err", err)
		} else {
			sessionID = rec.ID
		}
	}

	// Fallback to a locally-generated UUID when DB is unavailable.
	if sessionID == "" {
		sessionID = generateUUID()
	}

	session := &SessionState{
		SessionID: sessionID,
		DeviceInfo: DeviceInfo{
			Brand: deviceBrand,
			Model: deviceModel,
		},
		ConversationHistory: make([]ConversationMessage, 0),
		CreatedAt:           now,
		LastActivity:        now,
	}

	s.mu.Lock()
	s.sessions[session.SessionID] = session
	s.mu.Unlock()

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
// When a DB repo is available, the message is also persisted in a background
// goroutine (fire-and-forget with warning on failure).
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
	if len(session.ConversationHistory) > maxConversationHistory {
		session.ConversationHistory = session.ConversationHistory[len(session.ConversationHistory)-maxConversationHistory:]
	}

	session.LastActivity = now

	// Persist to DB in background without blocking the caller.
	if s.messageRepo != nil {
		msgType := "text"
		if hasImage {
			msgType = "image"
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			if _, err := s.messageRepo.SaveMessage(s.ctx, sessionID, role, content, msgType); err != nil {
				slog.Warn("db: failed to persist message", "session_id", sessionID, "err", err)
			}
		}()
	}

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
// When a DB repo is available, the update is also persisted in the background.
func (s *SessionService) SetProblemDetected(sessionID, problem string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %q not found", sessionID)
	}

	session.ProblemDetected = problem
	session.LastActivity = time.Now()

	// Persist to DB in background without blocking the caller.
	if s.sessionRepo != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			if err := s.sessionRepo.UpdateSessionStatus(s.ctx, sessionID, "active", &problem); err != nil {
				slog.Warn("db: failed to persist problem detected", "session_id", sessionID, "err", err)
			}
		}()
	}

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

// DeleteSession removes a session from the in-memory cache. In the DB the
// session is soft-deleted by setting its status to "abandoned" (not physically
// removed). Returns an error if the session is not found in cache.
func (s *SessionService) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		return fmt.Errorf("session %q not found", sessionID)
	}

	delete(s.sessions, sessionID)

	// Soft-delete in DB by marking as abandoned (nil preserves existing problem_type).
	if s.sessionRepo != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			if err := s.sessionRepo.UpdateSessionStatus(s.ctx, sessionID, "abandoned", nil); err != nil {
				slog.Warn("db: failed to mark session as abandoned", "session_id", sessionID, "err", err)
			}
		}()
	}

	return nil
}

// CleanupExpiredSessions removes all sessions whose LastActivity is older than maxAge.
// Designed to be called periodically in a background goroutine.
func (s *SessionService) CleanupExpiredSessions(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, session := range s.sessions {
		if session.LastActivity.Before(cutoff) {
			delete(s.sessions, id)
		}
	}
}

// generateUUID returns a new random UUID string for in-memory-only sessions.
func generateUUID() string {
	return uuid.New().String()
}
