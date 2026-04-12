// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestHealthConcurrent50Requests fires 50 concurrent GET /health requests and
// asserts that every response returns HTTP 200.
func TestHealthConcurrent50Requests(t *testing.T) {
	app, _ := newHealthApp("0.1.0", time.Now())

	const numRequests = 50
	var wg sync.WaitGroup
	var failures atomic.Int32

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			resp, err := app.Test(req)
			if err != nil {
				failures.Add(1)
				return
			}
			if resp.StatusCode != http.StatusOK {
				failures.Add(1)
			}
		}()
	}

	wg.Wait()

	if n := failures.Load(); n != 0 {
		t.Errorf("%d out of %d concurrent health requests did not return 200", n, numRequests)
	}
}

// TestHealthActiveConnsRaceFree verifies that concurrent reads of activeConns
// via ActiveConnections() do not race (exercised with -race flag).
func TestHealthActiveConnsRaceFree(t *testing.T) {
	_, ws := newHealthApp("0.1.0", time.Now())

	// Set a known initial value.
	ws.activeConns.Store(10)

	const numReaders = 20
	var wg sync.WaitGroup

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simultaneous reads — must not race.
			_ = ws.ActiveConnections()
		}()
	}

	wg.Wait()

	// Value must be unchanged.
	if got := ws.ActiveConnections(); got != 10 {
		t.Errorf("activeConns: want 10 after concurrent reads, got %d", got)
	}
}
