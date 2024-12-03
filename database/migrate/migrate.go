package main

import (
	"log"

	"github.com/TempFileLink/TempFileLink-BE/database"
)

func init() {
	database.ConnectDB()
}

func main() {
	database.DB.AutoMigrate()

	log.Println("Migration completed")
}
