package routers

import (
	"github.com/TempFileLink/TempFileLink-BE/handlers"
	"github.com/gofiber/fiber/v2"
)

func setupFileRoutes(api fiber.Router) {
	fileApi := api.Group("/file")

	fileApi.Get("/", handlers.FileMessage)
}
