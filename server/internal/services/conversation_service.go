// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

// frameSampleSize is the maximum number of bytes from the base64 frame
// used to compute the deduplication hash. Hashing only a prefix avoids
// processing multi-megabyte strings while still catching identical frames.
const frameSampleSize = 1024

// frameHashTTL is how long a frame hash is retained. After this duration the
// same frame will be re-analysed rather than skipped as a duplicate — prevents
// the map from leaking memory when sessions are long-lived or abandoned.
const frameHashTTL = 30 * time.Minute

// ProcessResult holds the combined Gemini analysis, overlay annotations,
// current step position, and a TTS-optimised voice hint returned to the client.
type ProcessResult struct {
	Text      string      `json:"text"`
	VoiceText string      `json:"voice_text,omitempty"`
	Overlay   OverlayData `json:"overlay"`
	Step      StepInfo    `json:"step"`
}

// StepInfo represents the user's position in a guided repair flow.
type StepInfo struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

// frameHashEntry stores a frame fingerprint alongside the time it was recorded.
// When hasPHash is true, pHash holds the perceptual hash and SHA-256 is a fallback
// for comparison with entries that lack a pHash (non-JPEG frames).
type frameHashEntry struct {
	sha256     string
	pHash      uint64
	hasPHash   bool
	recordedAt time.Time
}

// ConversationService orchestrates frame analysis and text conversations
// by coordinating SessionService (state), an AIBackend (AI inference), and an
// optional RAGService (knowledge-base context enrichment).
type ConversationService struct {
	sessions    *SessionService
	ai          AIBackend
	rag         *RAGService
	frameHashes map[string]frameHashEntry // sessionID -> (hash, timestamp)
	mu          sync.Mutex
}

// NewConversationService creates a ConversationService wired to the given
// session store, AI backend, and optional knowledge-base service.
// rag may be nil, in which case KB enrichment is skipped.
func NewConversationService(sessions *SessionService, ai AIBackend, rag *RAGService) *ConversationService {
	return &ConversationService{
		sessions:    sessions,
		ai:          ai,
		rag:         rag,
		frameHashes: make(map[string]frameHashEntry),
	}
}

// ProcessFrame analyzes a camera frame through Gemini and returns structured
// guidance. Identical consecutive frames (by hash) are silently skipped to
// avoid redundant API calls. An optional userText is recorded in the
// conversation history before analysis.
//
// When domain is non-empty and the RAG service has entries for that domain,
// knowledge-base context is restricted to that domain (e.g. "hvac", "printer").
// An empty string falls back to the global search across all domains.
func (c *ConversationService) ProcessFrame(ctx context.Context, sessionID, frameBase64, userText, domain string) (*ProcessResult, error) {
	// Compute perceptual fingerprint (pHash when JPEG, SHA-256 otherwise).
	sha256Hash, ph, hasPHash := computeFrameFingerprint(frameBase64)

	// Atomic check+store eliminates the TOCTOU window between isDuplicateFrame
	// and storeFrameHash. If Gemini fails later, clearFrameHash undoes the store.
	if c.checkAndStoreFrameHash(sessionID, sha256Hash, ph, hasPHash) {
		return nil, nil
	}

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
		if kb := c.rag.QueryKBByDomain(domain, query, 3); len(kb) > 0 {
			kbCtx := c.rag.FormatForPrompt(kb)
			if conversationCtx != "" {
				conversationCtx = kbCtx + "\n" + conversationCtx
			} else {
				conversationCtx = kbCtx
			}
		}
	}

	// Read session language for the Gemini prompt.
	language := "it"
	if snap, snapErr := c.sessions.GetSessionSnapshot(sessionID); snapErr == nil && snap.Language != "" {
		language = snap.Language
	}

	// Call AI backend.
	response, err := c.ai.AnalyzeFrame(ctx, frameBase64, conversationCtx, language)
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
		Text:      responseText,
		VoiceText: toVoiceText(response.Instruction),
		Overlay:   response.Overlay,
		Step: StepInfo{
			Current: session.CurrentStep,
			Total:   session.TotalSteps,
		},
	}, nil
}

// ProcessTextMessage handles a text-only conversation turn (no camera frame).
// The user message is recorded, sent to Gemini with empty frame data, and
// the assistant response is stored in history.
//
// domain restricts KB lookups to a specific domain when non-empty; an empty
// string falls back to the global search across all loaded KB domains.
func (c *ConversationService) ProcessTextMessage(ctx context.Context, sessionID, userText, domain string) (*ProcessResult, error) {
	if err := c.sessions.AddMessage(sessionID, "user", userText, false); err != nil {
		return nil, fmt.Errorf("add user message: %w", err)
	}

	conversationCtx, err := c.sessions.BuildContextForGemini(sessionID)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	// Enrich context with KB matches when RAG is available.
	if c.rag != nil {
		if kb := c.rag.QueryKBByDomain(domain, userText, 3); len(kb) > 0 {
			kbCtx := c.rag.FormatForPrompt(kb)
			if conversationCtx != "" {
				conversationCtx = kbCtx + "\n" + conversationCtx
			} else {
				conversationCtx = kbCtx
			}
		}
	}

	// Read session language for the Gemini prompt.
	language := "it"
	if snap, snapErr := c.sessions.GetSessionSnapshot(sessionID); snapErr == nil && snap.Language != "" {
		language = snap.Language
	}

	response, err := c.ai.AnalyzeFrame(ctx, "", conversationCtx, language)
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
		Text:      responseText,
		VoiceText: toVoiceText(response.Instruction),
		Overlay:   response.Overlay,
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

// computeFrameFingerprint computes both a SHA-256 hash (fallback) and a
// perceptual hash (preferred) for the given base64-encoded frame.
// hasPHash is false when the frame is not a decodable JPEG.
func computeFrameFingerprint(frameBase64 string) (sha256Hash string, ph uint64, hasPHash bool) {
	sha256Hash = hashFrame(frameBase64)
	ph, err := pHashFrame(frameBase64)
	if err != nil {
		return sha256Hash, 0, false
	}
	return sha256Hash, ph, true
}

// isDuplicateFrame returns true when the incoming frame is perceptually identical
// to the last processed frame for this session.
//
// When both frames carry a perceptual hash, the comparison uses Hamming distance
// (threshold: PHashDuplicateThreshold bits). Otherwise, it falls back to exact
// SHA-256 comparison.
func (c *ConversationService) isDuplicateFrame(sessionID, sha256Hash string, ph uint64, hasPHash bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.frameHashes[sessionID]
	if !ok {
		return false
	}
	if time.Since(entry.recordedAt) > frameHashTTL {
		delete(c.frameHashes, sessionID)
		return false
	}
	if hasPHash && entry.hasPHash {
		return hammingDistance(ph, entry.pHash) < PHashDuplicateThreshold
	}
	return sha256Hash == entry.sha256
}

// storeFrameHash records the latest frame fingerprint for a session.
func (c *ConversationService) storeFrameHash(sessionID, sha256Hash string, ph uint64, hasPHash bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.frameHashes[sessionID] = frameHashEntry{
		sha256:     sha256Hash,
		pHash:      ph,
		hasPHash:   hasPHash,
		recordedAt: time.Now(),
	}
}

// checkAndStoreFrameHash checks if the frame is a duplicate of the last
// processed frame for this session and, if not, stores it atomically.
// Returns true when the frame is a duplicate (caller should skip processing).
func (c *ConversationService) checkAndStoreFrameHash(sessionID, sha256Hash string, ph uint64, hasPHash bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.frameHashes[sessionID]
	if ok {
		if time.Since(entry.recordedAt) <= frameHashTTL {
			var isDup bool
			if hasPHash && entry.hasPHash {
				isDup = hammingDistance(ph, entry.pHash) < PHashDuplicateThreshold
			} else {
				isDup = sha256Hash == entry.sha256
			}
			if isDup {
				return true
			}
		}
	}
	c.frameHashes[sessionID] = frameHashEntry{
		sha256:     sha256Hash,
		pHash:      ph,
		hasPHash:   hasPHash,
		recordedAt: time.Now(),
	}
	return false
}

// CleanupSession removes the stored frame hash for the given session.
// Call this when a session is permanently deleted to free the associated
// frame-dedup entry immediately rather than waiting for TTL expiry.
func (c *ConversationService) CleanupSession(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.frameHashes, sessionID)
}

// clearFrameHash removes the stored frame hash for a session, allowing the
// same frame to be reprocessed. Called when AI analysis fails after an
// optimistic hash store.
func (c *ConversationService) clearFrameHash(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.frameHashes, sessionID)
}

// toVoiceText converts an instruction string into a TTS-friendly version by
// stripping Markdown syntax and collapsing excess whitespace.
func toVoiceText(text string) string {
	r := strings.NewReplacer(
		"**", "", "*", "", "__", "", "_", "",
		"```", "", "`", "",
		"###", "", "##", "", "#", "",
	)
	v := r.Replace(text)
	v = strings.Join(strings.Fields(v), " ")
	return strings.TrimSpace(v)
}
