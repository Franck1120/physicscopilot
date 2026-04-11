package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// defaultJWTSecret is used when SUPABASE_JWT_SECRET is not set.
// Replace with a real secret in production via the environment variable.
const defaultJWTSecret = "super-secret-jwt-token-with-at-least-32-characters-long"

func jwtSecret() []byte {
	if s := os.Getenv("SUPABASE_JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte(defaultJWTSecret)
}

// JWTAuth validates a Bearer JWT from the Authorization header.
// On success it stores the Supabase user UUID in c.Locals("user_id").
// Returns HTTP 401 JSON on missing or invalid tokens.
func JWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := tokenFromHeader(c)
		token, err := parseToken(raw)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		userID, _ := token.Claims.GetSubject()
		c.Locals("user_id", userID)
		return c.Next()
	}
}

// WSJWTAuth validates a JWT passed as the ?token= query param.
// Used for the WebSocket upgrade where browser clients cannot set
// custom Authorization headers.
// On success it stores the user UUID in c.Locals("user_id").
// Returns HTTP 401 JSON on missing or invalid tokens.
func WSJWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Query("token")
		token, err := parseToken(raw)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		userID, _ := token.Claims.GetSubject()
		c.Locals("user_id", userID)
		return c.Next()
	}
}

// tokenFromHeader extracts the raw JWT from "Authorization: Bearer <token>".
// Returns an empty string if the header is absent or malformed.
func tokenFromHeader(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	after, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok {
		return ""
	}
	return after
}

// parseToken validates a raw JWT string using HMAC-HS256 and the configured secret.
// It requires the exp claim to be present and not expired.
func parseToken(raw string) (*jwt.Token, error) {
	if raw == "" {
		return nil, jwt.ErrTokenMalformed
	}
	return jwt.Parse(raw,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret(), nil
		},
		jwt.WithExpirationRequired(),
	)
}
