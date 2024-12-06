package handlers

import (
	"github.com/TempFileLink/TempFileLink-BE/database"
	"github.com/TempFileLink/TempFileLink-BE/dto"
	"github.com/TempFileLink/TempFileLink-BE/models"
	"github.com/gofiber/fiber/v2"
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
