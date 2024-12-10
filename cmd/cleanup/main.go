package main

import (
	"context"
	"log"

	"github.com/TempFileLink/TempFileLink-BE/database"
	"github.com/TempFileLink/TempFileLink-BE/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		if record.EventName != "ObjectRemoved:Delete" {
			continue
		}

		s3Key := record.S3.Object.Key

		// Delete dari database
		if err := database.DB.Where("s3_key = ?", s3Key).Delete(&models.FileMetadata{}).Error; err != nil {
			log.Printf("Failed to delete metadata for file %s: %v", s3Key, err)
			return err
		}

		log.Printf("Successfully deleted metadata for file %s", s3Key)
	}
	return nil
}

func main() {
	database.ConnectDB()
	lambda.Start(handleRequest)
}
