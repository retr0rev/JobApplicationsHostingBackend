package email

import (
	"fmt"
	"log"
)

type Sender interface {
	Send(to, subject, body string) error
}

type ConsoleSender struct{}

func NewConsoleSender() *ConsoleSender {
	return &ConsoleSender{}
}

func (s *ConsoleSender) Send(to, subject, body string) error {
	log.Printf("[EMAIL] To: %s", to)
	log.Printf("[EMAIL] Subject: %s", subject)
	log.Printf("[EMAIL] Body:\n%s\n", body)
	return nil
}

func BuildVerifyEmail(verifyURL string) (subject, body string) {
	return "Verify your email",
		fmt.Sprintf(`Welcome! Please verify your email by clicking the link below:

%s

If you did not create this account, ignore this email.`, verifyURL)
}

func BuildResetEmail(resetURL string) (subject, body string) {
	return "Password reset request",
		fmt.Sprintf(`You requested a password reset. Click the link below to reset your password:

%s

This link expires in 1 hour.
If you did not request this, ignore this email.`, resetURL)
}
