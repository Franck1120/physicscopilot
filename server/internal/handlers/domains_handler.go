// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import "github.com/gofiber/fiber/v2"

// DomainsService is satisfied by *services.RAGService.
type DomainsService interface {
	KBDomains() []string
}

// DomainsHandler returns a Fiber handler for GET /api/domains.
// It returns the list of knowledge-base domains currently loaded.
// When rag is nil (e.g. in tests that do not configure a KB) an empty list is returned.
func DomainsHandler(rag DomainsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var domains []string
		if rag != nil {
			domains = rag.KBDomains()
		}
		if domains == nil {
			domains = []string{}
		}
		return c.JSON(fiber.Map{"domains": domains})
	}
}
