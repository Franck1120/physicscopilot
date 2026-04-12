package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"

	applogger "github.com/Franck1120/physicscopilot/server/internal/logger"
)

const (
	// apiRequestsPerMinute is the steady-state REST API request budget per IP.
	apiRequestsPerMinute = 60
	// apiLimiterBurst allows short bursts above the steady-state rate.
	apiLimiterBurst = 10
	// limiterExpiry removes idle per-IP limiters to prevent memory growth.
	limiterExpiry = 5 * time.Minute

	// banViolationThreshold is the number of rate-limit violations within
	// banViolationWindow that triggers a temporary IP ban.
	banViolationThreshold = 10
	// banViolationWindow is the sliding window over which violations are counted.
	banViolationWindow = 1 * time.Minute
	// banDuration is how long a banned IP is blocked before being retried.
	banDuration = 5 * time.Minute
)

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// violationEntry tracks how many times an IP exceeded the rate limit within
// the current banViolationWindow.
type violationEntry struct {
	count       int
	windowStart time.Time
}

// IPRateLimiter enforces per-IP request rate limits on REST endpoints.
// Each IP gets a token-bucket limiter refilling at ratePerMin/min
// with a burst allowance of burst tokens. IPs that exceed the rate limit
// banViolationThreshold times within banViolationWindow are temporarily
// banned for banDuration (HTTP 403 instead of 429).
type IPRateLimiter struct {
	mu         sync.Mutex
	limiters   map[string]*ipEntry
	violations map[string]*violationEntry
	bans       map[string]time.Time
	ratePerMin int
	burst      int
}

// NewIPRateLimiter creates a limiter with the production defaults
// (60 req/min, burst 10) and starts a background cleanup goroutine.
func NewIPRateLimiter() *IPRateLimiter {
	return newIPRateLimiterWith(apiRequestsPerMinute, apiLimiterBurst)
}

// newIPRateLimiterWith creates a limiter with custom rate and burst values.
// Intended for use in tests to exercise the 429 path without waiting a full minute.
func newIPRateLimiterWith(perMin, burst int) *IPRateLimiter {
	rl := &IPRateLimiter{
		limiters:   make(map[string]*ipEntry),
		violations: make(map[string]*violationEntry),
		bans:       make(map[string]time.Time),
		ratePerMin: perMin,
		burst:      burst,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	e, ok := rl.limiters[ip]
	if !ok {
		e = &ipEntry{
			limiter: rate.NewLimiter(
				rate.Every(time.Minute/time.Duration(rl.ratePerMin)),
				rl.burst,
			),
		}
		rl.limiters[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

func (rl *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(limiterExpiry)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, e := range rl.limiters {
			if time.Since(e.lastSeen) > limiterExpiry {
				delete(rl.limiters, ip)
			}
		}
		// Remove expired violation windows — they will be reset on the next hit anyway,
		// but cleaning them prevents unbounded growth for long-idle IPs.
		for ip, v := range rl.violations {
			if time.Since(v.windowStart) > banViolationWindow {
				delete(rl.violations, ip)
			}
		}
		// Remove bans that have already expired (isBanned also cleans on read,
		// but IPs that stop connecting would never trigger that path).
		for ip, until := range rl.bans {
			if time.Now().After(until) {
				delete(rl.bans, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns a Fiber handler that enforces the per-IP rate limit.
// Returns HTTP 403 when the IP is temporarily banned, HTTP 429 otherwise.
func (rl *IPRateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if rl.isBanned(ip) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "temporarily blocked due to repeated rate limit abuse — retry after 5 minutes",
			})
		}
		if !rl.getLimiter(ip).Allow() {
			rl.recordViolation(ip)
			applogger.SecurityLog("rate_limit_hit", "ip_hash", applogger.HashIP(ip))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded — try again in a moment",
			})
		}
		return c.Next()
	}
}

// isBanned returns true if ip is currently under a temporary ban.
func (rl *IPRateLimiter) isBanned(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	until, ok := rl.bans[ip]
	if !ok {
		return false
	}
	if time.Now().After(until) {
		delete(rl.bans, ip)
		return false
	}
	return true
}

// recordViolation increments the violation counter for ip. If the counter
// reaches banViolationThreshold within banViolationWindow, the IP is banned
// for banDuration.
func (rl *IPRateLimiter) recordViolation(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, ok := rl.violations[ip]
	if !ok || time.Since(v.windowStart) > banViolationWindow {
		rl.violations[ip] = &violationEntry{count: 1, windowStart: time.Now()}
		return
	}
	v.count++
	if v.count >= banViolationThreshold {
		rl.bans[ip] = time.Now().Add(banDuration)
		delete(rl.violations, ip)
		applogger.SecurityLog("ip_banned",
			"ip_hash", applogger.HashIP(ip),
			"duration_min", int(banDuration.Minutes()),
		)
	}
}

// ── UserRateLimiter ───────────────────────────────────────────────────────────

const (
	// userMessagesPerMinute is the maximum WebSocket API messages (frame/text)
	// a single authenticated user may trigger per minute.
	userMessagesPerMinute = 30
	// userLimiterBurst allows short bursts for each user.
	userLimiterBurst = 5
	// userLimiterExpiry removes idle per-user limiters to prevent memory growth.
	userLimiterExpiry = 10 * time.Minute
)

// UserRateLimiter enforces per-user rate limits keyed by the JWT subject claim
// (user ID). It is applied inside WebSocket handlers to cap expensive AI calls
// on a per-user basis, independently of the per-IP connection limit.
//
// When no user ID is available (unauthenticated / dev mode) Allow returns true
// to preserve the no-auth development experience.
type UserRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipEntry // reuse ipEntry (limiter + lastSeen)
	perMin   int
	burst    int
}

// NewUserRateLimiter creates a UserRateLimiter with production defaults
// (30 messages/min, burst 5) and starts a background cleanup goroutine.
func NewUserRateLimiter() *UserRateLimiter {
	return newUserRateLimiterWith(userMessagesPerMinute, userLimiterBurst)
}

// newUserRateLimiterWith creates a UserRateLimiter with custom values.
// Intended for tests.
func newUserRateLimiterWith(perMin, burst int) *UserRateLimiter {
	ul := &UserRateLimiter{
		limiters: make(map[string]*ipEntry),
		perMin:   perMin,
		burst:    burst,
	}
	go ul.cleanupLoop()
	return ul
}

// Allow returns true if the user identified by userID is within their rate
// limit. When userID is empty (unauthenticated request), it always returns true.
func (ul *UserRateLimiter) Allow(userID string) bool {
	if userID == "" {
		return true
	}
	ul.mu.Lock()
	defer ul.mu.Unlock()
	e, ok := ul.limiters[userID]
	if !ok {
		e = &ipEntry{
			limiter: rate.NewLimiter(
				rate.Every(time.Minute/time.Duration(ul.perMin)),
				ul.burst,
			),
		}
		ul.limiters[userID] = e
	}
	e.lastSeen = time.Now()
	return e.limiter.Allow()
}

func (ul *UserRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(userLimiterExpiry)
	defer ticker.Stop()
	for range ticker.C {
		ul.mu.Lock()
		for uid, e := range ul.limiters {
			if time.Since(e.lastSeen) > userLimiterExpiry {
				delete(ul.limiters, uid)
			}
		}
		ul.mu.Unlock()
	}
}
