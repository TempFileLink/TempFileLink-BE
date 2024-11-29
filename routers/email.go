package routers

import (
	"github.com/TempFileLink/TempFileLink-BE/handlers"
	"github.com/gofiber/fiber/v2"
)

func setupEmailRoutes(api fiber.Router) {
	emailApi := api.Group("/email")

	emailApi.Get("/", handlers.EmailMessage)
}
