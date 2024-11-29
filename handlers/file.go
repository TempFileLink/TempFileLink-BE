package handlers

import "github.com/gofiber/fiber/v2"

func FileMessage(c *fiber.Ctx) error {
	return c.SendString("File")
}
