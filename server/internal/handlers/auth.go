package handlers

import (
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
)

// WSAuthMiddleware returns a Fiber middleware that validates a JWT on the /ws
// endpoint. The token must be passed as a query parameter: /ws?token=<jwt>.
//
// When SUPABASE_JWT_SECRET is empty the middleware is a no-op (dev mode),
// so local development without Supabase still works.
func WSAuthMiddleware() fiber.Handler {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		// Dev mode: no secret configured → skip validation.
		return func(c *fiber.Ctx) error { return c.Next() }
	}

	keyBytes := []byte(secret)

	return func(c *fiber.Ctx) error {
		// Accept token from ?token= query param (standard for WS handshakes)
		// or from Authorization: Bearer <token> header.
		token := c.Query("token")
		if token == "" {
			if auth := c.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
				token = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authentication token")
		}

		parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "invalid signing method")
			}
			return keyBytes, nil
		})
		if err != nil || !parsed.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		return c.Next()
	}
}
