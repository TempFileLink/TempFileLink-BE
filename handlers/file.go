package handlers

import (
	"fmt"
	"mime/multipart"
	"time"

	"github.com/TempFileLink/TempFileLink-BE/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gofiber/fiber/v2"
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

func listObjects(client *s3.S3, bucketName string, prefix string) (*s3.ListObjectsV2Output, error) {
	res, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: &prefix,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func uploadFile(uploader *s3manager.Uploader, bucketName string, prefix string, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	filePath := fmt.Sprintf("%s%s", prefix, fileHeader.Filename)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filePath),
		Body:   file,
	})

	return err
}

func presignUrl(client *s3.S3, bucketName string, prefix string, name string) (string, error) {
	fileName := fmt.Sprintf("%s%s", prefix, name)

	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})

	urlStr, err := req.Presign(5 * time.Minute)
	if err != nil {
		return "", err
	}

	return urlStr, nil
}

/*
For Handling API call in routers/file.go
*/

func GetListFile(c *fiber.Ctx) error {
	sess, err := newSession()
	if err != nil {
		return c.SendString("Failed to create AWS session")
	}

	bucketName := "my-unique-bucket-kowan"
	prefix := "user1/" // Tinggal cara dapet prefix
	s3Client := s3.New(sess)

	objects, err := listObjects(s3Client, bucketName, prefix)
	if err != nil {
		return c.SendString("Failed")
	}

	for _, object := range objects.Contents {
		fmt.Printf("Found object: %s, size: %d\n", *object.Key, *object.Size)
	}

	return c.SendString("S3 session & client initialized")
}

func UploadFile(c *fiber.Ctx) error {
	sess, err := newSession()
	uploader := s3manager.NewUploader(sess)

	bucketName := "my-unique-bucket-kowan"
	prefix := "user1/" // Tinggal cara dapet prefix

	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	err = uploadFile(uploader, bucketName, prefix, file)
	if err != nil {
		c.SendString("Failed File")
	}

	return c.SendString("Upload File")
}

func GetFile(c *fiber.Ctx) error {
	sess, err := newSession()
	if err != nil {
		return c.SendString("Failed to create AWS session")
	}

	fileId := c.Params("fileId")
	if fileId == "" {
		return c.SendString("Error name")
	}

	// Handle mapping fileId ke fileName
	fileName := fileId

	bucketName := "my-unique-bucket-kowan"
	prefix := "user1/" // Tinggal cara dapet prefix
	s3Client := s3.New(sess)

	urlStr, err := presignUrl(s3Client, bucketName, prefix, fileName)
	if err != nil {
		fmt.Printf("Couldn't presign url: %v", err)
		return c.SendString("Failed to create AWS session")
	}

	return c.SendString(urlStr)
}
