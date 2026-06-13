package email

import (
	"fmt"
	"net/smtp"
	"os"
)

type SMTPSender struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSMTPSender() *SMTPSender {
	return &SMTPSender{
		host: os.Getenv("SMTP_HOST"),
		port: os.Getenv("SMTP_PORT"),
		user: os.Getenv("SMTP_USER"),
		pass: os.Getenv("SMTP_PASS"),
		from: os.Getenv("SMTP_FROM"),
	}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	if s.port == "" {
		s.port = "587"
	}

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s\r\n", s.from, to, subject, body))

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
