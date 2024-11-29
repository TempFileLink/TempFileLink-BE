package routers

import "github.com/gofiber/fiber/v2"

func SetupRoutes(api fiber.Router) {
	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	setupUserRoutes(api)
	setupFileRoutes(api)
	setupEmailRoutes(api)
}
