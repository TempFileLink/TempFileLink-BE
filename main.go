package main

import (
	"fmt"

	"github.com/TempFileLink/TempFileLink-BE/config"
	"github.com/TempFileLink/TempFileLink-BE/database"
	"github.com/TempFileLink/TempFileLink-BE/middlewares"
	"github.com/TempFileLink/TempFileLink-BE/routers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func init() {
	database.ConnectDB()
}

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: middlewares.ErrorHandler,
	})

	api := app.Group("/api")
	apiV1 := api.Group("/v1")

	// health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK!")
	})

	// Middlewares
	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.Config("FRONTEND_URL"),
		AllowCredentials: true,
	}))

	// Routers
	routers.SetupRoutes(apiV1)

	port := config.Config("PORT")
	port = fmt.Sprintf(":%s", port)

	app.Listen(port)
}
