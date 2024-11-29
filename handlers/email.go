package handlers

import "github.com/gofiber/fiber/v2"

func EmailMessage(c *fiber.Ctx) error {
	return c.SendString("Email")
}
