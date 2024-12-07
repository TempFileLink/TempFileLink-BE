package handlers

import (
	"fmt"
	"mime/multipart"
	"time"

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
		Bucket: aws.String(config.Config("AWS_BUCKET_NAME")),
		Key:    aws.String(filePath),
		Body:   file,
	})

	return err
}

func presignUrl(client *s3.S3, prefix string, name string) (string, error) {
	fileName := fmt.Sprintf("%s%s", prefix, name)

	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(config.Config("AWS_BUCKET_NAME")),
		Key:    aws.String(fileName),
	})

	urlStr, err := req.Presign(5 * time.Minute)
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
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	uploader := s3manager.NewUploader(sess)
	file, err := c.FormFile("file")

	err = uploadObject(uploader, prefix, file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Upload file failed")
	}

	/* Handle buat bikin password ke DB */

	// Return berhasil
	return c.SendString("File uploaded")
}

func GetFile(c *fiber.Ctx) error {
	prefix := getUserInfo(c.Locals("user").(*jwt.Token))

	sess, err := newSession()
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	fileName := c.Params("fileId")
	if fileName == "" {
		return c.Status(fiber.StatusBadRequest).SendString("File ID is missing")
	}

	s3Client := s3.New(sess)

	/* Handle buat cek password, klo benar baru proses */
	// Isi disini

	// Untuk presign + redirect URL
	urlStr, err := presignUrl(s3Client, prefix, fileName)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Redirect
	return c.Redirect(urlStr)
}
