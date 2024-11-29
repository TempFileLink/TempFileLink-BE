package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Config(key string) string {
	err := godotenv.Load(".env")
	value := os.Getenv(key)

	if err != nil && value == "" {
		log.Fatalf("Error loading the environment variable %s", key)
	}

	return value
}
