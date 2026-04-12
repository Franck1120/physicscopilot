package services

import (
	"context"
	"fmt"
	"os"
)

// OpenAIBackend is an AIBackend stub using the OpenAI-compatible API.
//
// This is a structural placeholder that satisfies the AIBackend interface.
// Full implementation will use OPENAI_API_KEY and the gpt-4o vision endpoint.
// Set AI_BACKEND=openai and configure OPENAI_API_KEY to activate.
type OpenAIBackend struct {
	apiKey string
}

// NewOpenAIBackend creates an OpenAIBackend.
// Returns an error when OPENAI_API_KEY is not set.
func NewOpenAIBackend() (*OpenAIBackend, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("AI_BACKEND=openai requires OPENAI_API_KEY to be set")
	}
	return &OpenAIBackend{apiKey: apiKey}, nil
}

// AnalyzeFrame is not yet implemented for the OpenAI backend.
func (o *OpenAIBackend) AnalyzeFrame(_ context.Context, _, _, _ string) (*GeminiResponse, error) {
	return nil, fmt.Errorf("OpenAI backend: AnalyzeFrame not yet implemented")
}

// ClaudeBackend is an AIBackend stub using the Anthropic Claude API.
//
// This is a structural placeholder that satisfies the AIBackend interface.
// Full implementation will use ANTHROPIC_API_KEY and the Claude vision API.
// Set AI_BACKEND=claude and configure ANTHROPIC_API_KEY to activate.
type ClaudeBackend struct {
	apiKey string
}

// NewClaudeBackend creates a ClaudeBackend.
// Returns an error when ANTHROPIC_API_KEY is not set.
func NewClaudeBackend() (*ClaudeBackend, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("AI_BACKEND=claude requires ANTHROPIC_API_KEY to be set")
	}
	return &ClaudeBackend{apiKey: apiKey}, nil
}

// AnalyzeFrame is not yet implemented for the Claude backend.
func (b *ClaudeBackend) AnalyzeFrame(_ context.Context, _, _, _ string) (*GeminiResponse, error) {
	return nil, fmt.Errorf("Claude backend: AnalyzeFrame not yet implemented")
}
