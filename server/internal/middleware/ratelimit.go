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
// Each IP gets a token-bucket limiter refilling at apiRequestsPerMinute/min
// with a burst allowance of apiLimiterBurst.
type IPRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipEntry
}

// NewIPRateLimiter creates a limiter and starts a background goroutine
// that periodically removes idle per-IP entries.
func NewIPRateLimiter() *IPRateLimiter {
	rl := &IPRateLimiter{
		limiters: make(map[string]*ipEntry),
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
				rate.Every(time.Minute/apiRequestsPerMinute),
				apiLimiterBurst,
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
