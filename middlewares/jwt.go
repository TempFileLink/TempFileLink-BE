package middlewares

import (
	"github.com/TempFileLink/TempFileLink-BE/config"
	jwt "github.com/gofiber/contrib/jwt"
)

var JWTWare = jwt.New(jwt.Config{
	SigningKey: jwt.SigningKey{Key: []byte(config.Config("JWT_SECRET"))},
})
