package inertia

import (
	"github.com/gofiber/fiber/v2"
)

func (i *Inertia) Middleware(c *fiber.Ctx) error {
	if c.Get("X-Inertia") == "" {
		return c.Next()
	}

	if c.Method() == "GET" && c.Get("X-Inertia-Version") != i.version {
		c.Set("X-Inertia-Location", i.url+c.OriginalURL())

		return c.SendStatus(fiber.StatusConflict)
	}

	// Go to next middleware:
	return c.Next()
}
