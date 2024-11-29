package routers

import (
	"github.com/TempFileLink/TempFileLink-BE/handlers"
	"github.com/gofiber/fiber/v2"
)

func setupUserRoutes(api fiber.Router) {
	userApi := api.Group("/user")

	userApi.Get("/", handlers.UserMessage)
}
