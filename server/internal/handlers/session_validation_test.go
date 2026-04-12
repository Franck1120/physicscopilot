package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSessionValidationTableDriven covers additional edge cases for the
// session creation endpoint, complementing the basic tests in session_handler_test.go.
func TestSessionValidationTableDriven(t *testing.T) {
	tests := []struct {
		name       string
		brand      string
		model      string
		wantStatus int
	}{
		{
			name:       "unicode device_brand is valid",
			brand:      "普鲁萨", // Chinese characters — valid Unicode, no HTML
			model:      "MK4",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "device_model with only whitespace is accepted",
			brand:      "Prusa",
			model:      "   ", // whitespace allowed — no validation rejects it
			wantStatus: http.StatusCreated,
		},
		{
			name:       "both fields empty is accepted",
			brand:      "",
			model:      "",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "device_brand exactly 100 chars is valid",
			brand:      strings.Repeat("x", maxDeviceFieldLen),
			model:      "MK4",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "device_brand exactly 101 chars is invalid",
			brand:      strings.Repeat("x", maxDeviceFieldLen+1),
			model:      "MK4",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newSessionTestApp(t)

			body := `{"device_brand":` + jsonString(tt.brand) + `,"device_model":` + jsonString(tt.model) + `}`
			req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

// jsonString wraps s in JSON double-quotes, escaping any embedded quotes.
// This avoids pulling in encoding/json for simple string construction in tests.
func jsonString(s string) string {
	// Use Go's JSON marshaller indirectly via strings.Builder to stay dependency-free.
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
