package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

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
	ConversationHistory []ConversationMessage `json:"conversation_history"`
	CurrentStep         int                   `json:"current_step"`
	TotalSteps          int                   `json:"total_steps"`
	ProblemDetected     string                `json:"problem_detected,omitempty"`
	CreatedAt           time.Time             `json:"created_at"`
	LastActivity        time.Time             `json:"last_activity"`
}

// SessionService manages in-memory repair sessions with thread-safe access.
type SessionService struct {
	mu       sync.RWMutex
	sessions map[string]*SessionState
}

// NewSessionService creates a SessionService with an initialized in-memory store.
func NewSessionService() *SessionService {
	return &SessionService{
		sessions: make(map[string]*SessionState),
	}
}

// CreateSession initializes a new repair session for the given device.
func (s *SessionService) CreateSession(deviceBrand, deviceModel string) (*SessionState, error) {
	now := time.Now()
	session := &SessionState{
		SessionID: uuid.New().String(),
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

// DeleteSession removes a session from the store. Returns an error if not found.
func (s *SessionService) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		return fmt.Errorf("session %q not found", sessionID)
	}

	delete(s.sessions, sessionID)
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
