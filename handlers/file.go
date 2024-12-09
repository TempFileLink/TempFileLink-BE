package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/TempFileLink/TempFileLink-BE/config"
	"github.com/TempFileLink/TempFileLink-BE/database"
	"github.com/TempFileLink/TempFileLink-BE/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

/*
Functionality
*/

func newSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Config("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			config.Config("AWS_ACCESS_KEY_ID"),
			config.Config("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})

	if err != nil {
		return nil, err
	}

	return sess, nil
}

func listObjects(client *s3.S3, prefix string) (*s3.ListObjectsV2Output, error) {
	res, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(config.Config("AWS_BUCKET_NAME")),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func uploadObject(uploader *s3manager.Uploader, prefix string, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	filePath := fmt.Sprintf("%s%s", prefix, fileHeader.Filename)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:  aws.String(config.Config("AWS_BUCKET_NAME")),
		Key:     aws.String(filePath),
		Body:    file,
		Tagging: aws.String("expiry=true"),
	})

	return err
}

func presignUrl(client *s3.S3, fileName string) (string, error) {
	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(config.Config("AWS_BUCKET_NAME")),
		Key:    aws.String(fileName),
	})

	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", err
	}

	return urlStr, nil
}

func getUserInfo(jwtUser *jwt.Token) string {
	claims := jwtUser.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	var user models.User
	database.DB.Where("email = ?", email).First(&user)

	// if user == (models.User{}) {
	// 	return fiber.NewError(fiber.StatusBadRequest, "Account is invalid")
	// }

	return fmt.Sprintf("%s/", user.ID.String())
}

/*
For Handling API call in routers/file.go
*/

func FileMessage(c *fiber.Ctx) error {
	return c.SendString("File")
}

func GetListFile(c *fiber.Ctx) error {
	/*
		contoh return format
		{
			"data": [
				{
					"name": "820d074d-a33c-4b03-b165-eb9c559bc621/file2.txt",
					"size": 49
				}
			]
		}
	*/
	prefix := getUserInfo(c.Locals("user").(*jwt.Token))

	sess, err := newSession()
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	s3Client := s3.New(sess)
	objects, err := listObjects(s3Client, prefix)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Change to Slice
	var response []fiber.Map
	for _, object := range objects.Contents {
		response = append(response, fiber.Map{"name": *object.Key, "size": *object.Size})
	}

	return c.JSON(fiber.Map{"data": response})
}

func UploadFile(c *fiber.Ctx) error {
	prefix := getUserInfo(c.Locals("user").(*jwt.Token))

	sess, err := newSession()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	uploader := s3manager.NewUploader(sess)
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File upload failed",
		})
	}

	// Get password if provided
	password := c.FormValue("password")
	var hashedPassword string
	isPassword := false

	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Password processing failed",
			})
		}
		hashedPassword = string(hash)
		isPassword = true
	}

	// Upload to S3
	s3Key := fmt.Sprintf("%s%s", prefix, file.Filename)
	err = uploadObject(uploader, prefix, file)
	if err != nil {
		log.Printf("Failed to upload file %s to S3: %v", s3Key, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Upload to S3 failed",
		})
	}

	// Save metadata to DB
	metadata := models.FileMetadata{
		UserID:     uuid.MustParse(prefix[0 : len(prefix)-1]),
		Filename:   file.Filename,
		S3Key:      s3Key,
		IsPassword: isPassword,
		Password:   hashedPassword,
		ExpiryTime: time.Now().Add(24 * time.Hour), // Default 24 jam
	}

	if err := database.DB.Create(&metadata).Error; err != nil {
		// Rollback S3 upload if DB fails
		s3Client := s3.New(sess)
		_, deleteErr := s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(config.Config("AWS_BUCKET_NAME")),
			Key:    aws.String(s3Key),
		})
		if deleteErr != nil {
			log.Printf("Failed to cleanup S3 after DB error: %v", deleteErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file metadata",
		})
	}

	return c.JSON(fiber.Map{
		"message":    "File uploaded successfully",
		"fileId":     metadata.ID,
		"expiryTime": metadata.ExpiryTime,
	})
}

func GetFile(c *fiber.Ctx) error {
	fileId := c.Params("fileId")
	if fileId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File ID is missing",
		})
	}

	// Check file metadata
	var metadata models.FileMetadata
	if err := database.DB.Where("id = ?", fileId).First(&metadata).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	// Check if file is expired
	if time.Now().After(metadata.ExpiryTime) {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{
			"error": "File has expired",
		})
	}

	// Check password if required
	if metadata.IsPassword {
		password := c.FormValue("password")
		if err := bcrypt.CompareHashAndPassword([]byte(metadata.Password), []byte(password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid password",
			})
		}
	}

	// Generate presigned URL
	sess, err := newSession()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	s3Client := s3.New(sess)
	urlStr, err := presignUrl(s3Client, metadata.S3Key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate download URL",
		})
	}

	return c.Redirect(urlStr)
}

func DeleteFile(c *fiber.Ctx) error {
	prefix := getUserInfo(c.Locals("user").(*jwt.Token))

	fileName := c.Params("fileId")
	if fileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File ID is missing",
		})
	}

	s3Key := fmt.Sprintf("%s%s", prefix, fileName)

	// Check file metadata dan ownership
	var metadata models.FileMetadata
	if err := database.DB.Where("s3_key = ?", s3Key).First(&metadata).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	// Delete dari S3
	sess, err := newSession()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	s3Client := s3.New(sess)
	_, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.Config("AWS_BUCKET_NAME")),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete file from S3",
		})
	}

	// Delete dari database
	if err := database.DB.Delete(&metadata).Error; err != nil {
		log.Printf("Failed to delete metadata for file %s: %v", s3Key, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete file metadata",
		})
	}

	return c.JSON(fiber.Map{
		"message": "File deleted successfully",
	})
}
