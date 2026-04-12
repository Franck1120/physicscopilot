// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
)

// HashIP returns an 8-character hex string derived from sha256(ip).
// Use this instead of logging raw IP addresses to avoid storing PII in logs
// while still allowing correlation of events from the same source.
func HashIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:4])
}

// SecurityLog emits a security audit event at WARN level via the global slog
// logger. Every record carries the "logger"="security" attribute so security
// events can be filtered independently from application logs:
//
//	grep '"logger":"security"' app.log
//	jq 'select(.logger=="security")' app.log
//
// event is the machine-readable event name (e.g. "auth_failure",
// "rate_limit_hit", "jpeg_validation_failure"). Additional key-value pairs
// can be passed as attrs.
func SecurityLog(event string, attrs ...any) {
	slog.Warn(event, append([]any{"logger", "security"}, attrs...)...)
}
