package mail

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v5"
)

type Mail struct {
	mg *mailgun.Client
}

func New() *Mail {
	b, err := os.ReadFile("env.env")
	if err != nil {
		log.Fatal(err)
	}
	apiKey := strings.TrimSpace(string(b[len("MAIL_KEY="):]))
	mg := mailgun.NewMailgun(apiKey)
	return &Mail{
		mg: mg,
	}
}

func (mail *Mail) SendMessage(recipient, link string) (string, error) {
	domain := "sandbox9628d10d1ee1475ab56628fba22071c1.mailgun.org"
	sender := "User Service <postmaster@sandbox9628d10d1ee1475ab56628fba22071c1.mailgun.org>"
	subject := "Email Confirmation"
	body := fmt.Sprintf("You need to confirm your email through this link: %s", link)

	message := mailgun.NewMessage(domain, sender, subject, body, recipient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, err := mail.mg.Send(ctx, message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID: %s Resp: %s\n", resp.ID, resp.Message)

	return resp.ID, nil
}
