package handlers

import (
	"fmt"
	"time"

	"github.com/TempFileLink/TempFileLink-BE/config"
	"github.com/TempFileLink/TempFileLink-BE/database"
	"github.com/TempFileLink/TempFileLink-BE/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gofiber/fiber/v2"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

const SMTP_HOST = "smtp.gmail.com"
const SMTP_PORT = 587

func EmailMessage(c *fiber.Ctx) error {
	return c.SendString("Email")
}

func SendEmail(c *fiber.Ctx) error {
	fileId := c.Params("fileId")
	if fileId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File ID is missing",
		})
	}

	var metadata models.FileMetadata
	if err := database.DB.Where("id = ?", fileId).First(&metadata).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	if time.Now().After(metadata.ExpiryTime) {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{
			"error": "File has expired",
		})
	}

	if metadata.IsPassword {
		file_password := c.FormValue("file_password")
		if err := bcrypt.CompareHashAndPassword([]byte(metadata.Password), []byte(file_password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid password",
			})
		}
	}

	sess, err := newSession()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	s3Client := s3.New(sess)
	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(config.Config("AWS_BUCKET_NAME")),
		Key:    aws.String(metadata.S3Key),
	})

	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate download URL",
		})
	}

	// sender_password must be linked to sender_email google account
	// via app passwords (different from gmail password)
	sender_password := c.FormValue("sender_password")
	sender_email := c.FormValue("sender_email")
	receiver_email := c.FormValue("receiver_email")
	filename := metadata.Filename     // assume injection-free
	message := c.FormValue("message") // assume injection-free
	subject := fmt.Sprintf("TempFile.Link - Shared Link to %s", filename)

	body := fmt.Sprintf(`
	<p><b>This person has sent you a link to their file <a href="%s">%s</a> 
	via <a href="https://tempfile.netlify.app/">TempFile.Link</a> 
	with the following message:</b></p>
	%s
	<p><b>Ignore this email if you don't recognize the sender</b></p>
	`, urlStr, filename, message)

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", sender_email)
	mailer.SetHeader("To", receiver_email)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body)

	dialer := gomail.NewDialer(
		SMTP_HOST,
		SMTP_PORT,
		sender_email,
		sender_password,
	)

	err = dialer.DialAndSend(mailer)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to send email",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Email sent successfully",
	})
}
