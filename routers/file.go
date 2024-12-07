package routers

import (
	"github.com/TempFileLink/TempFileLink-BE/handlers"
	"github.com/TempFileLink/TempFileLink-BE/middlewares"
	"github.com/gofiber/fiber/v2"
)

/*
Di DB isinya
fileId 				| isPassword	| password
Nama file asli		| true/false	| hashed
contoh:				|				|
<USER-ID>/<FILE>	|				|

Nanti redirect langsung aja ke signedURL filename
*/

func setupFileRoutes(api fiber.Router) {
	fileApi := api.Group("/file")

	fileApi.Get("/", handlers.FileMessage)
	fileApi.Get("/all", middlewares.JWTWare, handlers.GetListFile)
	fileApi.Get("/get/:fileId", middlewares.JWTWare, handlers.GetFile)
	fileApi.Post("/upload", middlewares.JWTWare, handlers.UploadFile)
}
