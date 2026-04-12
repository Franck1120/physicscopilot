package handlers

import (
	_ "embed"

	"github.com/gofiber/fiber/v2"
)

//go:embed openapi.yaml
var openapiSpecYAML []byte

// OpenAPIHandler serves GET /api/docs — the embedded OpenAPI 3.0 spec.
func OpenAPIHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/yaml; charset=utf-8")
		return c.Send(openapiSpecYAML)
	}
}
