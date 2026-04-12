package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

// TestSecurityLogEmitsWarnLevel verifies that SecurityLog writes at WARN level.
func TestSecurityLogEmitsWarnLevel(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	old := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(old)

	SecurityLog("test_event", "key", "value")

	if buf.Len() == 0 {
		t.Fatal("SecurityLog produced no output")
	}
	var rec map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}
	if rec["level"] != "WARN" {
		t.Errorf("SecurityLog level: want WARN, got %q", rec["level"])
	}
}

// TestSecurityLogIncludesSecurityLogger verifies that every record emitted by
// SecurityLog carries the "logger"="security" attribute for easy filtering.
func TestSecurityLogIncludesSecurityLogger(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	old := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(old)

	SecurityLog("auth_failure", "ip", "1.2.3.4")

	var rec map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}
	if rec["logger"] != "security" {
		t.Errorf("SecurityLog logger attribute: want security, got %q", rec["logger"])
	}
}

// TestSecurityLogIncludesEventName verifies that the event name appears as the
// log message so queries like `jq 'select(.msg=="auth_failure")'` work.
func TestSecurityLogIncludesEventName(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	old := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(old)

	SecurityLog("rate_limit_hit", "ip_hash", "abc123")

	var rec map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}
	if rec["msg"] != "rate_limit_hit" {
		t.Errorf("SecurityLog msg: want rate_limit_hit, got %q", rec["msg"])
	}
	if rec["ip_hash"] != "abc123" {
		t.Errorf("SecurityLog ip_hash attr: want abc123, got %q", rec["ip_hash"])
	}
}
