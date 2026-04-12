// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import "github.com/gofiber/fiber/v2"

// DomainsService is satisfied by *services.RAGService.
type DomainsService interface {
	KBDomains() []string
	// DomainEntryCount returns the number of knowledge-base entries loaded for
	// the given domain. Returns 0 when the domain is not found.
	DomainEntryCount(domain string) int
}

// DomainDetail is the per-domain object returned when ?detailed=true.
type DomainDetail struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// DomainsHandler returns a Fiber handler for GET /api/domains.
//
// Default (detailed=false): {"domains":["hvac","printer"]}
// With ?detailed=true:      {"domains":[{"name":"hvac","count":20},{"name":"printer","count":8}]}
//
// When rag is nil (e.g. in tests that do not configure a KB) an empty list is returned.
func DomainsHandler(rag DomainsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		detailed := c.Query("detailed") == "true"

		if detailed {
			var details []DomainDetail
			if rag != nil {
				for _, d := range rag.KBDomains() {
					details = append(details, DomainDetail{
						Name:  d,
						Count: rag.DomainEntryCount(d),
					})
				}
			}
			if details == nil {
				details = []DomainDetail{}
			}
			return c.JSON(fiber.Map{"domains": details})
		}

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
