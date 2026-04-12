// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

// Package handlers implements the HTTP and WebSocket handlers for the
// PhysicsCopilot server.
package handlers

import "github.com/gofiber/fiber/v2"

// swaggerUIHTML is the inline HTML page that loads Swagger UI from CDN and
// points it at the /api/docs OpenAPI spec endpoint.
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>PhysicsCopilot API — Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/api/docs",
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: "BaseLayout",
      deepLinking: true
    });
  </script>
</body>
</html>`

// SwaggerUIHandler returns a Fiber handler that serves an HTML page embedding
// Swagger UI (loaded from the unpkg CDN at version 5). The UI is pre-configured
// to load the OpenAPI specification from GET /api/docs.
//
// Register as:
//
//	api.Get("/swagger", handlers.SwaggerUIHandler())
func SwaggerUIHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Set("Cache-Control", "no-cache, no-store")
		return c.SendString(swaggerUIHTML)
	}
}
