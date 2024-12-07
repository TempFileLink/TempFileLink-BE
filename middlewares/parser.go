package middlewares

import (
	"github.com/gofiber/fiber/v2"
)

func BodyParser[T any](t T) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := new(T)

		if err := c.BodyParser(body); err != nil {
			return err
		}

		c.Locals("body", body)

		return c.Next()
	}
}
