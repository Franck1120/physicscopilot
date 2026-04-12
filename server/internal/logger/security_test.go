package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"regexp"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// captureSecurityLog redirects slog output to a buffer, calls fn, then returns
// the parsed JSON record.
func captureSecurityLog(t *testing.T, fn func()) map[string]interface{} {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	old := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(old)

	fn()

	if buf.Len() == 0 {
		t.Fatal("SecurityLog produced no output")
	}
	var rec map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log output is not valid JSON: %v\nraw: %s", err, buf.String())
	}
	return rec
}

// ---------------------------------------------------------------------------
// Existing tests (preserved exactly)
// ---------------------------------------------------------------------------

// TestSecurityLogEmitsWarnLevel verifies that SecurityLog writes at WARN level.
func TestSecurityLogEmitsWarnLevel(t *testing.T) {
	rec := captureSecurityLog(t, func() {
		SecurityLog("test_event", "key", "value")
	})
	if rec["level"] != "WARN" {
		t.Errorf("SecurityLog level: want WARN, got %q", rec["level"])
	}
}

// TestSecurityLogIncludesSecurityLogger verifies that every record emitted by
// SecurityLog carries the "logger"="security" attribute for easy filtering.
func TestSecurityLogIncludesSecurityLogger(t *testing.T) {
	rec := captureSecurityLog(t, func() {
		SecurityLog("auth_failure", "ip", "1.2.3.4")
	})
	if rec["logger"] != "security" {
		t.Errorf("SecurityLog logger attribute: want security, got %q", rec["logger"])
	}
}

// TestSecurityLogIncludesEventName verifies that the event name appears as the
// log message so queries like `jq 'select(.msg=="auth_failure")'` work.
func TestSecurityLogIncludesEventName(t *testing.T) {
	rec := captureSecurityLog(t, func() {
		SecurityLog("rate_limit_hit", "ip_hash", "abc123")
	})
	if rec["msg"] != "rate_limit_hit" {
		t.Errorf("SecurityLog msg: want rate_limit_hit, got %q", rec["msg"])
	}
	if rec["ip_hash"] != "abc123" {
		t.Errorf("SecurityLog ip_hash attr: want abc123, got %q", rec["ip_hash"])
	}
}

// ---------------------------------------------------------------------------
// HashIP tests
// ---------------------------------------------------------------------------

// TestHashIPLength verifies that HashIP always returns exactly 8 hex characters.
func TestHashIPLength(t *testing.T) {
	inputs := []string{"1.2.3.4", "::1", "255.255.255.255", "192.168.0.1"}
	for _, ip := range inputs {
		got := HashIP(ip)
		if len(got) != 8 {
			t.Errorf("HashIP(%q) length: want 8, got %d (%q)", ip, len(got), got)
		}
	}
}

// TestHashIPDeterministic verifies that calling HashIP twice with the same
// input produces the same output.
func TestHashIPDeterministic(t *testing.T) {
	ip := "10.0.0.1"
	a := HashIP(ip)
	b := HashIP(ip)
	if a != b {
		t.Errorf("HashIP(%q) is not deterministic: first=%q, second=%q", ip, a, b)
	}
}

// TestHashIPDifferentInputs verifies that different IPs produce different hashes.
func TestHashIPDifferentInputs(t *testing.T) {
	h1 := HashIP("1.2.3.4")
	h2 := HashIP("5.6.7.8")
	if h1 == h2 {
		t.Errorf("HashIP collision: both %q and %q produced %q", "1.2.3.4", "5.6.7.8", h1)
	}
}

// TestHashIPEmptyString verifies that HashIP("") does not panic and returns 8 chars.
func TestHashIPEmptyString(t *testing.T) {
	got := HashIP("")
	if len(got) != 8 {
		t.Errorf("HashIP(\"\") length: want 8, got %d (%q)", len(got), got)
	}
}

// TestHashIPFormat verifies that the output contains only lowercase hex characters.
func TestHashIPFormat(t *testing.T) {
	hexOnly := regexp.MustCompile(`^[0-9a-f]+$`)
	inputs := []string{"1.2.3.4", "::1", "", "192.168.100.200"}
	for _, ip := range inputs {
		got := HashIP(ip)
		if !hexOnly.MatchString(got) {
			t.Errorf("HashIP(%q) = %q: contains non-hex characters", ip, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Additional SecurityLog event tests
// ---------------------------------------------------------------------------

// TestSecurityLogAuthFailure verifies that SecurityLog handles "auth_failure".
func TestSecurityLogAuthFailure(t *testing.T) {
	rec := captureSecurityLog(t, func() {
		SecurityLog("auth_failure", "ip_hash", "abc")
	})
	if rec["msg"] != "auth_failure" {
		t.Errorf("msg: want auth_failure, got %q", rec["msg"])
	}
	if rec["ip_hash"] != "abc" {
		t.Errorf("ip_hash: want abc, got %q", rec["ip_hash"])
	}
}

// TestSecurityLogRateLimitHit verifies that SecurityLog handles "rate_limit_hit".
func TestSecurityLogRateLimitHit(t *testing.T) {
	rec := captureSecurityLog(t, func() {
		SecurityLog("rate_limit_hit", "ip_hash", "abc")
	})
	if rec["msg"] != "rate_limit_hit" {
		t.Errorf("msg: want rate_limit_hit, got %q", rec["msg"])
	}
}

// TestSecurityLogIPBanned verifies that SecurityLog handles "ip_banned".
func TestSecurityLogIPBanned(t *testing.T) {
	rec := captureSecurityLog(t, func() {
		SecurityLog("ip_banned", "ip_hash", "abc")
	})
	if rec["msg"] != "ip_banned" {
		t.Errorf("msg: want ip_banned, got %q", rec["msg"])
	}
}

// TestSecurityLogJpegValidationFailure verifies "jpeg_validation_failure" event.
func TestSecurityLogJpegValidationFailure(t *testing.T) {
	rec := captureSecurityLog(t, func() {
		SecurityLog("jpeg_validation_failure")
	})
	if rec["msg"] != "jpeg_validation_failure" {
		t.Errorf("msg: want jpeg_validation_failure, got %q", rec["msg"])
	}
	if rec["logger"] != "security" {
		t.Errorf("logger: want security, got %q", rec["logger"])
	}
}

// TestSecurityLogInputSanitized verifies "input_sanitized" event with field attr.
func TestSecurityLogInputSanitized(t *testing.T) {
	rec := captureSecurityLog(t, func() {
		SecurityLog("input_sanitized", "field", "name")
	})
	if rec["msg"] != "input_sanitized" {
		t.Errorf("msg: want input_sanitized, got %q", rec["msg"])
	}
	if rec["field"] != "name" {
		t.Errorf("field: want name, got %q", rec["field"])
	}
}

// TestSecurityLogMultipleAttrs verifies that all key-value pairs appear in output.
func TestSecurityLogMultipleAttrs(t *testing.T) {
	rec := captureSecurityLog(t, func() {
		SecurityLog("test", "k1", "v1", "k2", "v2")
	})
	if rec["k1"] != "v1" {
		t.Errorf("k1: want v1, got %q", rec["k1"])
	}
	if rec["k2"] != "v2" {
		t.Errorf("k2: want v2, got %q", rec["k2"])
	}
}
