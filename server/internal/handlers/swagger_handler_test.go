// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
	"strings"
	"testing"
)

func TestSwaggerUIHTMLDarkTheme(t *testing.T) {
	html := SwaggerUIHTML("/openapi.yaml")

	// Must contain the dark background colour.
	if !strings.Contains(html, "background: #1a1a2e") {
		t.Error("SwaggerUIHTML: dark background 'background: #1a1a2e' not found in HTML")
	}

	// Topbar must be hidden.
	if !strings.Contains(html, "display: none") {
		t.Error("SwaggerUIHTML: topbar 'display: none' not found in HTML")
	}

	// Must NOT contain the Swagger default white/light background.
	if strings.Contains(html, "background: #fff") || strings.Contains(html, "background-color: #fff") ||
		strings.Contains(html, "background: white") {
		t.Error("SwaggerUIHTML: Swagger default light background found — dark theme not applied")
	}
}

func TestSwaggerUIHTMLContainsSpecURL(t *testing.T) {
	const specURL = "/openapi.yaml"
	html := SwaggerUIHTML(specURL)

	if !strings.Contains(html, specURL) {
		t.Errorf("SwaggerUIHTML: spec URL %q not found in HTML", specURL)
	}
}

func TestSwaggerUIHTMLContainsThemeColorMeta(t *testing.T) {
	html := SwaggerUIHTML("/openapi.yaml")

	if !strings.Contains(html, `name="theme-color"`) {
		t.Error("SwaggerUIHTML: <meta name=\"theme-color\"> not found in HTML")
	}
	if !strings.Contains(html, "#1a1a2e") {
		t.Error("SwaggerUIHTML: theme-color value '#1a1a2e' not found in HTML")
	}
}

func TestSwaggerUIHTMLIsValidHTML(t *testing.T) {
	html := SwaggerUIHTML("/openapi.yaml")

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("SwaggerUIHTML: missing DOCTYPE declaration")
	}
	if !strings.Contains(html, "<html") {
		t.Error("SwaggerUIHTML: missing <html> tag")
	}
	if !strings.Contains(html, "</html>") {
		t.Error("SwaggerUIHTML: missing closing </html> tag")
	}
	if !strings.Contains(html, "swagger-ui") {
		t.Error("SwaggerUIHTML: missing swagger-ui reference")
	}
}
