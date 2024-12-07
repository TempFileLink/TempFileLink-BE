package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load(".env")
}

func Config(key string) string {
	value := os.Getenv(key)

	if value == "" {
		log.Fatalf("Error loading the environment variable %s", key)
	}

	return value
}
