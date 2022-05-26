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

	sess, _ := i.store.Get(c)

	message := sess.Get("message")

	if message != nil {
		i.Share("message", sess.Get("message"))
		sess.Delete("message")
	}

	// Go to next middleware:
	return c.Next()
}
