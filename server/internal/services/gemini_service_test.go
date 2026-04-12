package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// validGeminiJSON returns a well-formed Gemini API response envelope
// wrapping the given structured payload.
func validGeminiJSON(payload string) string {
	return `{"candidates":[{"content":{"parts":[{"text":` + payload + `}]}}]}`
}

func TestNewGeminiServiceFallsBackToProxy(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("CLIPROXY_URL", "")

	svc, err := NewGeminiService()
	if err != nil {
		t.Fatalf("unexpected error when using CLIProxyAPI fallback: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service when using CLIProxyAPI fallback")
	}
	if !svc.useProxy {
		t.Error("expected useProxy=true when GEMINI_API_KEY is missing")
	}
	expected := defaultCLIProxyBaseURL + "/v1/chat/completions"
	if svc.proxyURL != expected {
		t.Errorf("expected proxyURL %q, got %q", expected, svc.proxyURL)
	}
}

func TestNewGeminiServiceDefaultURL(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GEMINI_BASE_URL", "")

	svc, err := NewGeminiService()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.baseURL != defaultGeminiURL {
		t.Errorf("expected default URL %q, got %q", defaultGeminiURL, svc.baseURL)
	}
	if svc.apiKey != "test-key" {
		t.Errorf("expected apiKey 'test-key', got %q", svc.apiKey)
	}
}

func TestNewGeminiServiceCustomURL(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GEMINI_BASE_URL", "http://localhost:8085/v1/generate")

	svc, err := NewGeminiService()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.baseURL != "http://localhost:8085/v1/generate" {
		t.Errorf("expected custom URL, got %q", svc.baseURL)
	}
}

func TestAnalyzeFrameSuccess(t *testing.T) {
	problem := "stringing detected"
	structured := `"{\"analysis\":\"printer bed visible\",\"problem\":\"stringing detected\",\"instruction\":\"lower temp by 5C\",\"overlay\":{\"boxes\":[{\"x\":0.1,\"y\":0.2,\"w\":0.3,\"h\":0.4,\"label\":\"stringing\"}],\"arrows\":[{\"x1\":0.1,\"y1\":0.2,\"x2\":0.5,\"y2\":0.6}]}}"`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request structure
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		// Verify API key is sent via header, not query params
		key := r.Header.Get("x-goog-api-key")
		if key != "test-key-123" {
			t.Errorf("expected x-goog-api-key header 'test-key-123', got %q", key)
		}
		if qKey := r.URL.Query().Get("key"); qKey != "" {
			t.Errorf("API key should not be in query params, got %q", qKey)
		}

		// Verify request body has expected structure
		body, _ := io.ReadAll(r.Body)
		var req geminiRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if len(req.Contents) != 1 {
			t.Fatalf("expected 1 content entry, got %d", len(req.Contents))
		}
		if req.Contents[0].Role != "user" {
			t.Errorf("expected role 'user', got %q", req.Contents[0].Role)
		}
		// With frameBase64 provided, should have text + inline_data parts
		if len(req.Contents[0].Parts) != 2 {
			t.Fatalf("expected 2 parts (text + image), got %d", len(req.Contents[0].Parts))
		}
		if req.Contents[0].Parts[1].InlineData == nil {
			t.Fatal("expected inline_data in second part")
		}
		if req.Contents[0].Parts[1].InlineData.MimeType != "image/jpeg" {
			t.Errorf("expected mime_type image/jpeg, got %q", req.Contents[0].Parts[1].InlineData.MimeType)
		}
		if req.GenerationConfig.Temperature != geminiTemperature {
			t.Errorf("expected temperature %v, got %f", geminiTemperature, req.GenerationConfig.Temperature)
		}
		if req.GenerationConfig.MaxOutputTokens != geminiMaxOutputTokens {
			t.Errorf("expected maxOutputTokens %d, got %d", geminiMaxOutputTokens, req.GenerationConfig.MaxOutputTokens)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validGeminiJSON(structured)))
	}))
	defer server.Close()

	svc := &GeminiService{
		apiKey:     "test-key-123",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	resp, err := svc.AnalyzeFrame(context.Background(), "base64imagedata", "user: printer clicking", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Analysis != "printer bed visible" {
		t.Errorf("expected analysis 'printer bed visible', got %q", resp.Analysis)
	}
	if resp.Problem == nil || *resp.Problem != problem {
		t.Errorf("expected problem %q, got %v", problem, resp.Problem)
	}
	if resp.Instruction != "lower temp by 5C" {
		t.Errorf("expected instruction 'lower temp by 5C', got %q", resp.Instruction)
	}
	if len(resp.Overlay.Boxes) != 1 {
		t.Fatalf("expected 1 box, got %d", len(resp.Overlay.Boxes))
	}
	if resp.Overlay.Boxes[0].Label != "stringing" {
		t.Errorf("expected label 'stringing', got %q", resp.Overlay.Boxes[0].Label)
	}
	if len(resp.Overlay.Arrows) != 1 {
		t.Fatalf("expected 1 arrow, got %d", len(resp.Overlay.Arrows))
	}
}

func TestAnalyzeFrameWithoutImage(t *testing.T) {
	structured := `"{\"analysis\":\"no image provided\",\"problem\":null,\"instruction\":\"please show camera\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}"`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req geminiRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		// Without image, should only have 1 part (text only)
		if len(req.Contents[0].Parts) != 1 {
			t.Fatalf("expected 1 part (text only), got %d", len(req.Contents[0].Parts))
		}
		if req.Contents[0].Parts[0].InlineData != nil {
			t.Fatal("expected no inline_data when frameBase64 is empty")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validGeminiJSON(structured)))
	}))
	defer server.Close()

	svc := &GeminiService{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	resp, err := svc.AnalyzeFrame(context.Background(), "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Problem != nil {
		t.Errorf("expected nil problem, got %v", resp.Problem)
	}
	if resp.Analysis != "no image provided" {
		t.Errorf("expected analysis 'no image provided', got %q", resp.Analysis)
	}
}

func TestAnalyzeFrameNullProblem(t *testing.T) {
	structured := `"{\"analysis\":\"all good\",\"problem\":null,\"instruction\":\"keep going\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}"`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validGeminiJSON(structured)))
	}))
	defer server.Close()

	svc := &GeminiService{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	resp, err := svc.AnalyzeFrame(context.Background(), "img", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Problem != nil {
		t.Errorf("expected nil problem for null JSON value, got %v", resp.Problem)
	}
}

func TestAnalyzeFrameHTTP4xxNoRetry(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	svc := &GeminiService{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	_, err := svc.AnalyzeFrame(context.Background(), "img", "", "")
	if err == nil {
		t.Fatal("expected error on 400 response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error should contain status code 400, got: %v", err)
	}

	// 4xx (not 429) should NOT be retried — only 1 call
	if count := callCount.Load(); count != 1 {
		t.Errorf("expected exactly 1 call (no retry on 400), got %d", count)
	}
}

func TestAnalyzeFrameHTTP429Retries(t *testing.T) {
	var callCount atomic.Int32
	structured := `"{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"done\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}"`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limited"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validGeminiJSON(structured)))
	}))
	defer server.Close()

	svc := &GeminiService{
		apiKey:  "test-key",
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: server.Client().Transport,
		},
	}

	resp, err := svc.AnalyzeFrame(context.Background(), "img", "", "")
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if resp.Analysis != "ok" {
		t.Errorf("expected analysis 'ok', got %q", resp.Analysis)
	}

	// Should have been called 3 times: 2 retries + 1 success
	if count := callCount.Load(); count != 3 {
		t.Errorf("expected 3 calls (2 retries + success), got %d", count)
	}
}

func TestAnalyzeFrameHTTP500RetriesExhausted(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	svc := &GeminiService{
		apiKey:  "test-key",
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: server.Client().Transport,
		},
	}

	_, err := svc.AnalyzeFrame(context.Background(), "img", "", "")
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention status code 500, got: %v", err)
	}
	if !strings.Contains(err.Error(), "after 3 attempts") {
		t.Errorf("error should mention retry exhaustion, got: %v", err)
	}

	// Should have been called exactly 3 times
	if count := callCount.Load(); count != 3 {
		t.Errorf("expected 3 calls (all retries), got %d", count)
	}
}

func TestAnalyzeFrameNoCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"candidates":[]}`))
	}))
	defer server.Close()

	svc := &GeminiService{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	_, err := svc.AnalyzeFrame(context.Background(), "img", "", "")
	if err == nil {
		t.Fatal("expected error when no candidates returned")
	}
	if !strings.Contains(err.Error(), "no candidates") {
		t.Errorf("error should mention 'no candidates', got: %v", err)
	}
}

func TestAnalyzeFrameInvalidJSON(t *testing.T) {
	// The inner text is not valid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"not json at all"}]}}]}`))
	}))
	defer server.Close()

	svc := &GeminiService{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	_, err := svc.AnalyzeFrame(context.Background(), "img", "", "")
	if err == nil {
		t.Fatal("expected error when inner text is invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse gemini structured response") {
		t.Errorf("error should mention structured response parse, got: %v", err)
	}
}

func TestAnalyzeFrameContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response — the context should cancel before this returns
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := &GeminiService{
		apiKey:  "test-key",
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: server.Client().Transport,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := svc.AnalyzeFrame(ctx, "img", "", "")
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
}

func TestBuildRequestBodyWithConversationContext(t *testing.T) {
	svc := &GeminiService{
		apiKey:     "key",
		baseURL:    "http://test",
		httpClient: &http.Client{},
	}

	req := svc.buildRequestBody("imgdata", "user: help me", "it")

	if len(req.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(req.Contents))
	}

	textPart := req.Contents[0].Parts[0]
	if !strings.Contains(textPart.Text, systemPromptBase) {
		t.Error("expected system prompt in text part")
	}
	if !strings.Contains(textPart.Text, "Conversation context:") {
		t.Error("expected conversation context header in text part")
	}
	if !strings.Contains(textPart.Text, "user: help me") {
		t.Error("expected conversation context content in text part")
	}

	// Image part
	if len(req.Contents[0].Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(req.Contents[0].Parts))
	}
	imgPart := req.Contents[0].Parts[1]
	if imgPart.InlineData == nil {
		t.Fatal("expected inline_data for image")
	}
	if imgPart.InlineData.Data != "imgdata" {
		t.Errorf("expected image data 'imgdata', got %q", imgPart.InlineData.Data)
	}
}

func TestBuildRequestBodyWithoutContext(t *testing.T) {
	svc := &GeminiService{
		apiKey:     "key",
		baseURL:    "http://test",
		httpClient: &http.Client{},
	}

	req := svc.buildRequestBody("", "", "it")

	if len(req.Contents[0].Parts) != 1 {
		t.Fatalf("expected 1 part (text only), got %d", len(req.Contents[0].Parts))
	}

	textPart := req.Contents[0].Parts[0]
	if textPart.Text != systemPromptForLanguage("it") {
		t.Error("expected only system prompt when no context")
	}
	if textPart.InlineData != nil {
		t.Error("expected no inline_data when no image")
	}
	if !strings.Contains(textPart.Text, "PhysicsCopilot") {
		t.Error("expected PhysicsCopilot in system prompt")
	}
}

func TestParseGeminiResponseEmptyParts(t *testing.T) {
	body := []byte(`{"candidates":[{"content":{"parts":[]}}]}`)

	_, err := parseGeminiResponse(body)
	if err == nil {
		t.Fatal("expected error for empty parts")
	}
	if !strings.Contains(err.Error(), "no content parts") {
		t.Errorf("error should mention 'no content parts', got: %v", err)
	}
}

func TestParseGeminiResponseEmptyText(t *testing.T) {
	body := []byte(`{"candidates":[{"content":{"parts":[{"text":""}]}}]}`)

	_, err := parseGeminiResponse(body)
	if err == nil {
		t.Fatal("expected error for empty text")
	}
	if !strings.Contains(err.Error(), "text is empty") {
		t.Errorf("error should mention 'text is empty', got: %v", err)
	}
}

func TestParseGeminiResponseBrokenEnvelope(t *testing.T) {
	_, err := parseGeminiResponse([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for broken envelope JSON")
	}
	if !strings.Contains(err.Error(), "parse gemini response envelope") {
		t.Errorf("error should mention envelope parse, got: %v", err)
	}
}

// TestNewGeminiServiceIntegrationEnvOverride verifies that the factory
// function correctly reads both env vars together.
func TestNewGeminiServiceIntegrationEnvOverride(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "my-secret-key")
	t.Setenv("GEMINI_BASE_URL", "http://localhost:8085/proxy")

	svc, err := NewGeminiService()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.apiKey != "my-secret-key" {
		t.Errorf("expected apiKey 'my-secret-key', got %q", svc.apiKey)
	}
	if svc.baseURL != "http://localhost:8085/proxy" {
		t.Errorf("expected custom baseURL, got %q", svc.baseURL)
	}
	if svc.httpClient == nil {
		t.Fatal("expected non-nil httpClient")
	}
	if svc.httpClient.Timeout != httpTimeout {
		t.Errorf("expected timeout %v, got %v", httpTimeout, svc.httpClient.Timeout)
	}

	// t.Setenv automatically restores env vars on test cleanup
}
