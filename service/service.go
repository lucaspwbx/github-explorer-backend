package service

import (
	"log"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendEmail(username, email string) error {
	from := mail.NewEmail("Trending Repos", "trendingrepos@xyz.com")
	subject := "Welcome to Trending Repos"
	to := mail.NewEmail(username, email)
	plainTextContent := "Welcome do Trending Repos!"
	htmlContent := "<strong>Welcome to Trending Repos!</strong>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
