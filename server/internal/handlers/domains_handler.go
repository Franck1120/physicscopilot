// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// DomainsService is satisfied by *services.RAGService.
type DomainsService interface {
	KBDomains() []string
}

// domainsETag computes a weak ETag over the domains slice by hashing its JSON
// representation and returning the first 8 bytes as a hex string.
func domainsETag(domains []string) string {
	b, _ := json.Marshal(domains)
	h := sha256.Sum256(b)
	return fmt.Sprintf(`W/"%x"`, h[:8])
}

// DomainsHandler returns a Fiber handler for GET /api/domains.
// It returns the list of knowledge-base domains currently loaded.
// When rag is nil (e.g. in tests that do not configure a KB) an empty list is returned.
//
// The response carries an ETag based on the current domain list. Clients may
// send If-None-Match with the ETag to receive a 304 Not Modified when the list
// has not changed. The response is publicly cacheable for 5 minutes
// (Cache-Control: public, max-age=300).
func DomainsHandler(rag DomainsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var domains []string
		if rag != nil {
			domains = rag.KBDomains()
		}
		if domains == nil {
			domains = []string{}
		}

		etag := domainsETag(domains)
		if c.Get("If-None-Match") == etag {
			return c.SendStatus(fiber.StatusNotModified)
		}

		c.Set("ETag", etag)
		c.Set("Cache-Control", "public, max-age=300")
		return c.JSON(fiber.Map{"domains": domains})
	}
}
