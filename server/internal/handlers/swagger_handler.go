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

// SwaggerUIHandler returns a Fiber handler that serves an inline HTML page
// embedding Swagger UI (loaded from the unpkg CDN at version 5). The UI is
// pre-configured to point at GET /api/docs for the OpenAPI 3.0 specification.
//
// CDN dependency: the response references two assets from
// https://unpkg.com/swagger-ui-dist@5. Clients require internet access to
// load the interactive UI; in air-gapped environments bundle the assets
// locally and update the HTML template accordingly.
//
// Cache policy: the page itself is served with Cache-Control: no-cache,
// no-store so that the browser always fetches the latest route registration.
// The CDN assets are cached by the browser according to unpkg's own headers.
//
// Security: no user-supplied data is rendered in the HTML; the response is
// safe from injection attacks. The Content-Security-Policy header set by the
// global middleware must allow https://unpkg.com for the UI to function.
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
