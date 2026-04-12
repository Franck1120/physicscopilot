package services

import (
	"context"
	"fmt"
	"os"
)

// AIBackend is the interface for AI inference backends.
// Any backend that can analyze camera frames and text and return structured
// PhysicsCopilot responses must implement this interface.
//
// The active backend is chosen at startup via the AI_BACKEND environment
// variable (default: "gemini"). Swap backends without changing callers.
type AIBackend interface {
	// AnalyzeFrame sends a camera frame (base64-encoded JPEG) and conversation
	// context for analysis, returning structured guidance.
	// frameBase64 may be empty for text-only turns.
	// language is a BCP-47 code (e.g. "it", "en") controlling the response language.
	AnalyzeFrame(ctx context.Context, frameBase64, conversationContext, language string) (*AIResponse, error)
}

// NewAIBackend creates the AI backend selected by the AI_BACKEND env var.
//
//   - "gemini"  (default) → GeminiService using GEMINI_API_KEY or CLIProxyAPI
//   - "openai"            → OpenAIBackend using OPENAI_API_KEY (stub)
//   - "claude"            → ClaudeBackend using ANTHROPIC_API_KEY (stub)
//
// Returns an error for unknown backend values or missing required API keys.
func NewAIBackend() (AIBackend, error) {
	backend := os.Getenv("AI_BACKEND")
	if backend == "" {
		backend = "gemini"
	}

	switch backend {
	case "gemini":
		return NewGeminiService()
	case "openai":
		return NewOpenAIBackend()
	case "claude":
		return NewClaudeBackend()
	default:
		return nil, fmt.Errorf("unknown AI_BACKEND %q: supported values are gemini, openai, claude", backend)
	}
}
