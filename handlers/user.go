package handlers

import "github.com/gofiber/fiber/v2"

func UserMessage(c *fiber.Ctx) error {
	return c.SendString("User")
}
