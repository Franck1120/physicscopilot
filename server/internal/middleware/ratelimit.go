package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

const (
	// apiRequestsPerMinute is the steady-state REST API request budget per IP.
	apiRequestsPerMinute = 60
	// apiLimiterBurst allows short bursts above the steady-state rate.
	apiLimiterBurst = 10
	// limiterExpiry removes idle per-IP limiters to prevent memory growth.
	limiterExpiry = 5 * time.Minute
)

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter enforces per-IP request rate limits on REST endpoints.
// Each IP gets a token-bucket limiter refilling at ratePerMin/min
// with a burst allowance of burst tokens.
type IPRateLimiter struct {
	mu         sync.Mutex
	limiters   map[string]*ipEntry
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
		rl.mu.Unlock()
	}
}

// Middleware returns a Fiber handler that enforces the per-IP rate limit.
// Returns HTTP 429 when the limit is exceeded.
func (rl *IPRateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !rl.getLimiter(c.IP()).Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded — try again in a moment",
			})
		}
		return c.Next()
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
