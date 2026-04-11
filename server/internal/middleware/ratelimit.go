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
