package handlers

import (
	"time"

	"github.com/TempFileLink/TempFileLink-BE/config"
	"github.com/TempFileLink/TempFileLink-BE/database"
	"github.com/TempFileLink/TempFileLink-BE/dto"
	"github.com/TempFileLink/TempFileLink-BE/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func UserMessage(c *fiber.Ctx) error {
	return c.SendString("User")
}

func Register(c *fiber.Ctx) error {
	userDto := c.Locals("body").(*dto.User)

	err := dto.Validate(userDto)

	if err != nil {
		return err
	}

	var existingUser models.User
	database.DB.Where("email = ?", userDto.Email).First(&existingUser)

	if existingUser != (models.User{}) {
		return fiber.NewError(fiber.StatusBadRequest, "An account already exists with this email address.")
	}

	hash, hashErr := bcrypt.GenerateFromPassword([]byte(userDto.Password), bcrypt.DefaultCost)

	if hashErr != nil {
		return err
	}

	user := models.User{
		Email:    userDto.Email,
		Password: string(hash),
	}

	result := database.DB.Create(&user)

	if result.Error != nil {
		return result.Error
	}

	return c.JSON(user)
}

func Login(c *fiber.Ctx) error {
	userDto := c.Locals("body").(*dto.User)

	// Check if the user exists
	var user models.User
	database.DB.Where("email = ?", userDto.Email).First(&user)

	if user == (models.User{}) {
		return fiber.NewError(fiber.StatusBadRequest, "Account is invalid")
	}

	// Check if the password is correct
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userDto.Password))

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Account is invalid")
	}

	// Create a JWT token
	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(config.Config("JWT_SECRET")))

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}
