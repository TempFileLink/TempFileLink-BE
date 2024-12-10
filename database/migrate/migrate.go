package main

import (
	"log"

	"github.com/TempFileLink/TempFileLink-BE/database"
	"github.com/TempFileLink/TempFileLink-BE/models"
)

func init() {
	database.ConnectDB()
}

func main() {
	database.DB.AutoMigrate(&models.Model{}, &models.User{}, &models.FileMetadata{})

	log.Println("Migration completed")
}
