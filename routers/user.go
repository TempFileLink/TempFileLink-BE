package routers

import (
	"github.com/TempFileLink/TempFileLink-BE/dto"
	"github.com/TempFileLink/TempFileLink-BE/handlers"
	"github.com/TempFileLink/TempFileLink-BE/middlewares"
	"github.com/gofiber/fiber/v2"
)

func setupUserRoutes(api fiber.Router) {
	userApi := api.Group("/user")

	userApi.Get("/", handlers.UserMessage)
	userApi.Post("/register", middlewares.BodyParser(dto.User{}), handlers.Register)
	userApi.Post("/login", middlewares.BodyParser(dto.User{}), handlers.Login)
	userApi.Get("/profile", middlewares.JWTWare, handlers.Profile)
}
