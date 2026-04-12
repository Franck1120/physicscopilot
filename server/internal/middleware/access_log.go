// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

package middleware

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ApacheCommonLog returns a Fiber middleware that logs each completed request
// in Apache Common Log Format:
//
//	127.0.0.1 - - [02/Jan/2006:15:04:05 -0700] "GET /path HTTP/1.1" 200 1234
//
// Activate via APP_ACCESS_LOG_FORMAT=apache env var; if unset it is a no-op.
func ApacheCommonLog() fiber.Handler {
	if os.Getenv("APP_ACCESS_LOG_FORMAT") != "apache" {
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		ip := c.IP()
		ts := start.Format("02/Jan/2006:15:04:05 -0700")
		method := c.Method()
		path := c.OriginalURL()
		status := c.Response().StatusCode()
		size := len(c.Response().Body())
		log.Printf(`%s - - [%s] "%s %s HTTP/1.1" %d %d`, ip, ts, method, path, status, size)
		return err
	}
}
