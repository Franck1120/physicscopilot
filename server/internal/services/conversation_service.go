package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
)

// frameSampleSize is the maximum number of bytes from the base64 frame
// used to compute the deduplication hash. Hashing only a prefix avoids
// processing multi-megabyte strings while still catching identical frames.
const frameSampleSize = 1024

// ProcessResult holds the combined Gemini analysis, overlay annotations,
// and current step position returned to the client.
type ProcessResult struct {
	Text    string   `json:"text"`
	Overlay OverlayData `json:"overlay"`
	Step    StepInfo    `json:"step"`
}

// StepInfo represents the user's position in a guided repair flow.
type StepInfo struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

// ConversationService orchestrates frame analysis and text conversations
// by coordinating SessionService (state), GeminiService (AI), and an
// optional RAGService (knowledge-base context enrichment).
type ConversationService struct {
	sessions    *SessionService
	gemini      *GeminiService
	rag         *RAGService
	frameHashes map[string]string // sessionID -> hash of last processed frame
	mu          sync.Mutex
}

// NewConversationService creates a ConversationService wired to the given
// session store, Gemini client, and optional knowledge-base service.
// rag may be nil, in which case KB enrichment is skipped.
func NewConversationService(sessions *SessionService, gemini *GeminiService, rag *RAGService) *ConversationService {
	return &ConversationService{
		sessions:    sessions,
		gemini:      gemini,
		rag:         rag,
		frameHashes: make(map[string]string),
	}
}

// ProcessFrame analyzes a camera frame through Gemini and returns structured
// guidance. Identical consecutive frames (by hash) are silently skipped to
// avoid redundant API calls. An optional userText is recorded in the
// conversation history before analysis.
func (c *ConversationService) ProcessFrame(ctx context.Context, sessionID, frameBase64, userText string) (*ProcessResult, error) {
	// Deduplicate identical consecutive frames
	hash := hashFrame(frameBase64)
	if c.isDuplicateFrame(sessionID, hash) {
		return nil, nil
	}

	// Store hash optimistically to block concurrent duplicate frames (TOCTOU fix).
	// If Gemini fails, the hash is cleared so the frame can be retried.
	c.storeFrameHash(sessionID, hash)

	// Record user message if provided
	if userText != "" {
		if err := c.sessions.AddMessage(sessionID, "user", userText, false); err != nil {
			c.clearFrameHash(sessionID)
			return nil, fmt.Errorf("add user message: %w", err)
		}
	}

	// Build conversation context for Gemini
	conversationCtx, err := c.sessions.BuildContextForGemini(sessionID)
	if err != nil {
		c.clearFrameHash(sessionID)
		return nil, fmt.Errorf("build context: %w", err)
	}

	// Enrich context with KB matches when RAG is available.
	// Query with userText; fall back to detected problem if text is empty.
	if c.rag != nil {
		query := userText
		if query == "" {
			if snap, snapErr := c.sessions.GetSessionSnapshot(sessionID); snapErr == nil {
				query = snap.ProblemDetected
			}
		}
		if kb := c.rag.QueryKB(query, 3); len(kb) > 0 {
			kbCtx := c.rag.FormatForPrompt(kb)
			if conversationCtx != "" {
				conversationCtx = kbCtx + "\n" + conversationCtx
			} else {
				conversationCtx = kbCtx
			}
		}
	}

	// Call Gemini Vision API
	response, err := c.gemini.AnalyzeFrame(ctx, frameBase64, conversationCtx)
	if err != nil {
		c.clearFrameHash(sessionID)
		return nil, fmt.Errorf("analyze frame: %w", err)
	}

	// Record detected problem
	if response.Problem != nil {
		if err := c.sessions.SetProblemDetected(sessionID, *response.Problem); err != nil {
			return nil, fmt.Errorf("set problem detected: %w", err)
		}
	}

	// Build assistant response text
	responseText := response.Analysis + "\n\n" + response.Instruction

	// Record assistant response in history
	if err := c.sessions.AddMessage(sessionID, "assistant", responseText, false); err != nil {
		return nil, fmt.Errorf("add assistant message: %w", err)
	}

	// Retrieve current step info (snapshot avoids data race with concurrent UpdateStep)
	session, err := c.sessions.GetSessionSnapshot(sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session for step info: %w", err)
	}

	return &ProcessResult{
		Text:    responseText,
		Overlay: response.Overlay,
		Step: StepInfo{
			Current: session.CurrentStep,
			Total:   session.TotalSteps,
		},
	}, nil
}

// ProcessTextMessage handles a text-only conversation turn (no camera frame).
// The user message is recorded, sent to Gemini with empty frame data, and
// the assistant response is stored in history.
func (c *ConversationService) ProcessTextMessage(ctx context.Context, sessionID, userText string) (*ProcessResult, error) {
	if err := c.sessions.AddMessage(sessionID, "user", userText, false); err != nil {
		return nil, fmt.Errorf("add user message: %w", err)
	}

	conversationCtx, err := c.sessions.BuildContextForGemini(sessionID)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	// Enrich context with KB matches when RAG is available.
	if c.rag != nil {
		if kb := c.rag.QueryKB(userText, 3); len(kb) > 0 {
			kbCtx := c.rag.FormatForPrompt(kb)
			if conversationCtx != "" {
				conversationCtx = kbCtx + "\n" + conversationCtx
			} else {
				conversationCtx = kbCtx
			}
		}
	}

	response, err := c.gemini.AnalyzeFrame(ctx, "", conversationCtx)
	if err != nil {
		return nil, fmt.Errorf("analyze text message: %w", err)
	}

	responseText := response.Analysis + "\n\n" + response.Instruction

	if err := c.sessions.AddMessage(sessionID, "assistant", responseText, false); err != nil {
		return nil, fmt.Errorf("add assistant message: %w", err)
	}

	// Snapshot avoids data race with concurrent UpdateStep
	session, err := c.sessions.GetSessionSnapshot(sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session for step info: %w", err)
	}

	return &ProcessResult{
		Text:    responseText,
		Overlay: response.Overlay,
		Step: StepInfo{
			Current: session.CurrentStep,
			Total:   session.TotalSteps,
		},
	}, nil
}

// GetSessionStep returns the current step position for the given session.
func (c *ConversationService) GetSessionStep(sessionID string) (*StepInfo, error) {
	session, err := c.sessions.GetSessionSnapshot(sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	return &StepInfo{
		Current: session.CurrentStep,
		Total:   session.TotalSteps,
	}, nil
}

// hashFrame computes a SHA-256 hex digest over the first frameSampleSize
// bytes of the base64-encoded frame data.
func hashFrame(frameBase64 string) string {
	sample := frameBase64
	if len(sample) > frameSampleSize {
		sample = sample[:frameSampleSize]
	}
	h := sha256.Sum256([]byte(sample))
	return fmt.Sprintf("%x", h)
}

// isDuplicateFrame checks whether the frame hash matches the last processed
// frame for this session. Thread-safe via mutex.
func (c *ConversationService) isDuplicateFrame(sessionID, hash string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.frameHashes[sessionID] == hash
}

// storeFrameHash records the latest frame hash for a session. Thread-safe.
func (c *ConversationService) storeFrameHash(sessionID, hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.frameHashes[sessionID] = hash
}

// clearFrameHash removes the stored frame hash for a session, allowing the
// same frame to be reprocessed. Called when Gemini analysis fails after an
// optimistic hash store.
func (c *ConversationService) clearFrameHash(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.frameHashes, sessionID)
}
