package routers

import (
	"github.com/TempFileLink/TempFileLink-BE/handlers"
	"github.com/gofiber/fiber/v2"
)

/*
Di DB isinya
fileName		| fileId 		| isPassword	| password	| username
Nama file asli	| Nama di S3	| true/false	| hashed	| Username dari yg punya
*/

func setupFileRoutes(api fiber.Router) {
	fileApi := api.Group("/file")

	fileApi.Get("/", handlers.GetListFile)
	fileApi.Get("/get/:fileId", handlers.GetFile)
	fileApi.Post("/upload", handlers.UploadFile)
}
