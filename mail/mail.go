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

type CacheAPI interface {
	Add(userID, token string) error
	Get(userID string) (string, bool)
	Delete(userID string)
}

type Mail struct {
	mg    *mailgun.Client
	table CacheAPI

	domain string
	sender string
}

func New(c CacheAPI) *Mail {
	b, err := os.ReadFile("env.env")
	if err != nil {
		log.Fatal(err)
	}
	apiKey := strings.TrimSpace(string(b[len("MAIL_KEY="):]))
	mg := mailgun.NewMailgun(apiKey)
	return &Mail{
		mg:    mg,
		table: c,
		// вынести в конфиг
		domain: "sandbox9628d10d1ee1475ab56628fba22071c1.mailgun.org",
		sender: "User Service <postmaster@sandbox9628d10d1ee1475ab56628fba22071c1.mailgun.org>",
	}
}

func (m *Mail) sendMessage(subject, recipient, message string) (string, error) {
	msg := mailgun.NewMessage(m.domain, m.sender, subject, message, recipient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, err := m.mg.Send(ctx, msg)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (m *Mail) SendEmailConfirmationMessage(userID, email string, mailtoken, link string) (string, error) {
	if _, ok := m.table.Get(userID); ok {
		// чтобы не было возможности израскодовать квоту писем
		return "", ErrMsgAlreadySent
	}
	err := m.table.Add(userID, mailtoken)
	if err != nil {
		return "", err
	}
	msg := fmt.Sprintf("You need to confirm your email through this link: %s", link)
	msgID, err := m.sendMessage("Email Confirmation", email, msg)
	if err != nil {
		m.table.Delete(userID)
		return "", err
	}
	return msgID, nil
}

func (m *Mail) CheckToken(userID, token string) bool {
	mtoken, ok := m.table.Get(userID)
	if !ok || mtoken != token {
		return false
	}
	m.table.Delete(userID)
	return true
}
