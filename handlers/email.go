package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/gomail.v2"
)

const SMTP_HOST = "smtp.gmail.com"
const SMTP_PORT = 587

func EmailMessage(c *fiber.Ctx) error {
	return c.SendString("Email")
}

func SendEmail(c *fiber.Ctx) error {
	presigned_url := c.FormValue("presigned_url")
	if presigned_url == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Presigned URL is missing",
		})
	}

	sender_email := c.FormValue("sender_email")
	if sender_email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Sender email is missing",
		})
	}

	// sender_password must be linked to sender_email google account
	// via app passwords (different from gmail password)
	sender_password := c.FormValue("sender_password")
	if sender_password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Sender password is missing",
		})
	}

	receiver_email := c.FormValue("receiver_email")
	if receiver_email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Receiver email is missing",
		})
	}

	message := c.FormValue("message") // assume injection-free
	body_message := "</b></p>"
	if message != "" {
		body_message = fmt.Sprintf(
			"with the following message:</b></p><p>%s</p>", message,
		)
	}

	subject := "TempFile.Link - Shared Link"

	body := fmt.Sprintf(`
	<p><b>This person has sent you a link to their <a href="%s">file</a> 
	via <a href="https://tempfile.netlify.app/">TempFile.Link</a> 
	%s
	<p><b>Ignore this email if you don't recognize the sender</b></p>
	`, presigned_url, body_message)

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

	err := dialer.DialAndSend(mailer)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to send email",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Email sent successfully",
	})
}
