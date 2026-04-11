package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// defaultGeminiURL is the production endpoint for Gemini 2.5 Flash.
const defaultGeminiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

// maxRetries is the number of attempts before giving up on transient errors.
const maxRetries = 3

// geminiTemperature controls the randomness of Gemini responses (lower = more deterministic).
const geminiTemperature = 0.2

// geminiMaxOutputTokens caps the length of Gemini's generated response.
const geminiMaxOutputTokens = 1024

// httpTimeout is the deadline for each individual HTTP request to Gemini.
const httpTimeout = 30 * time.Second

// systemPrompt instructs Gemini to behave as a 3D printing technician
// and return structured JSON analysis of camera frames.
const systemPrompt = `You are PhysicsCopilot, an expert 3D printing technician. You see what the user's camera shows in real-time. Analyze the image, identify any 3D printing issues (stringing, warping, layer shifting, under-extrusion, etc.), and provide step-by-step guidance. Respond ONLY with valid JSON with these fields: {"analysis": "what you see", "problem": "identified issue or null", "instruction": "next step for user", "overlay": {"boxes": [{"x": 0.1, "y": 0.2, "w": 0.3, "h": 0.4, "label": "stringing"}], "arrows": [{"x1": 0.1, "y1": 0.2, "x2": 0.3, "y2": 0.4}]}}`

// GeminiResponse is the structured analysis returned by AnalyzeFrame.
type GeminiResponse struct {
	Analysis    string      `json:"analysis"`
	Problem     *string     `json:"problem"`
	Instruction string      `json:"instruction"`
	Overlay     OverlayData `json:"overlay"`
}

// OverlayData contains visual annotations to render on the camera feed.
type OverlayData struct {
	Boxes  []BoundingBox `json:"boxes"`
	Arrows []Arrow       `json:"arrows"`
}

// BoundingBox marks a rectangular region of interest in normalized coordinates (0-1).
type BoundingBox struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	H     float64 `json:"h"`
	Label string  `json:"label"`
}

// Arrow represents a directional indicator in normalized coordinates (0-1).
type Arrow struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

// GeminiService communicates with the Google Gemini Vision API to analyze
// 3D printer camera frames and return structured repair instructions.
type GeminiService struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewGeminiService creates a GeminiService by reading configuration from
// environment variables. Returns an error if GEMINI_API_KEY is not set.
// GEMINI_BASE_URL overrides the default endpoint (useful for local proxies).
func NewGeminiService() (*GeminiService, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	baseURL := os.Getenv("GEMINI_BASE_URL")
	if baseURL == "" {
		baseURL = defaultGeminiURL
	}

	return &GeminiService{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}, nil
}

// AnalyzeFrame sends a camera frame (base64-encoded JPEG) and conversation
// context to Gemini, then parses the structured JSON response.
// If frameBase64 is empty, only the text prompt is sent (no image).
// Retries up to 3 times with exponential backoff on HTTP 429 and 5xx errors.
func (g *GeminiService) AnalyzeFrame(ctx context.Context, frameBase64, conversationContext string) (*GeminiResponse, error) {
	reqBody := g.buildRequestBody(frameBase64, conversationContext)

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	respBody, err := g.doWithRetry(ctx, payload)
	if err != nil {
		return nil, err
	}

	return parseGeminiResponse(respBody)
}

// geminiRequest mirrors the JSON structure expected by the Gemini REST API.
type geminiRequest struct {
	Contents         []geminiContent  `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inline_data,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type generationConfig struct {
	ResponseMimeType string  `json:"responseMimeType"`
	Temperature      float64 `json:"temperature"`
	MaxOutputTokens  int     `json:"maxOutputTokens"`
}

// geminiRawResponse mirrors the top-level structure of a Gemini API response.
type geminiRawResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content geminiCandidateContent `json:"content"`
}

type geminiCandidateContent struct {
	Parts []geminiCandidatePart `json:"parts"`
}

type geminiCandidatePart struct {
	Text string `json:"text"`
}

// buildRequestBody assembles the Gemini request payload.
// When frameBase64 is non-empty, the image is included as inline_data.
func (g *GeminiService) buildRequestBody(frameBase64, conversationContext string) geminiRequest {
	promptText := systemPrompt
	if conversationContext != "" {
		promptText += "\n\nConversation context:\n" + conversationContext
	}

	parts := []geminiPart{
		{Text: promptText},
	}

	if frameBase64 != "" {
		parts = append(parts, geminiPart{
			InlineData: &inlineData{
				MimeType: "image/jpeg",
				Data:     frameBase64,
			},
		})
	}

	return geminiRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: parts,
			},
		},
		GenerationConfig: generationConfig{
			ResponseMimeType: "application/json",
			Temperature:      geminiTemperature,
			MaxOutputTokens:  geminiMaxOutputTokens,
		},
	}
}

// doWithRetry executes the HTTP POST to Gemini with exponential backoff.
// Retries on HTTP 429 (rate limit) and 5xx (server error).
// Does NOT retry on other 4xx client errors.
func (g *GeminiService) doWithRetry(ctx context.Context, payload []byte) ([]byte, error) {
	backoff := 1 * time.Second

	var lastErr error
	for attempt := range maxRetries {
		body, err := g.doRequest(ctx, g.baseURL, payload)
		if err == nil {
			return body, nil
		}

		// Only retry on retryable errors
		retryErr, ok := err.(*retryableError)
		if !ok {
			return nil, err
		}

		lastErr = retryErr

		// Don't sleep after the final attempt
		if attempt < maxRetries-1 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
			backoff *= 2
		}
	}

	return nil, fmt.Errorf("gemini API failed after %d attempts: %w", maxRetries, lastErr)
}

// retryableError signals that the HTTP request can be retried.
type retryableError struct {
	statusCode int
	message    string
}

func (e *retryableError) Error() string {
	return e.message
}

// doRequest sends a single HTTP POST and returns the response body.
// Returns a retryableError for 429 and 5xx status codes.
// Returns a plain error for other non-2xx status codes.
func (g *GeminiService) doRequest(ctx context.Context, url string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, &retryableError{
			statusCode: resp.StatusCode,
			message:    fmt.Sprintf("gemini API returned HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}

	return nil, fmt.Errorf("gemini API returned HTTP %d: %s", resp.StatusCode, string(body))
}

// parseGeminiResponse extracts the structured JSON from the Gemini API
// raw response envelope: candidates[0].content.parts[0].text -> GeminiResponse.
func parseGeminiResponse(body []byte) (*GeminiResponse, error) {
	var raw geminiRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse gemini response envelope: %w", err)
	}

	if len(raw.Candidates) == 0 {
		return nil, fmt.Errorf("gemini returned no candidates")
	}

	parts := raw.Candidates[0].Content.Parts
	if len(parts) == 0 {
		return nil, fmt.Errorf("gemini candidate has no content parts")
	}

	text := parts[0].Text
	if text == "" {
		return nil, fmt.Errorf("gemini candidate text is empty")
	}

	var result GeminiResponse
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse gemini structured response: %w", err)
	}

	return &result, nil
}
