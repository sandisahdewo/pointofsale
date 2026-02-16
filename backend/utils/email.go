package utils

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

//go:embed templates/welcome.html
var welcomeTemplate string

//go:embed templates/password_reset.html
var passwordResetTemplate string

//go:embed templates/account_approved.html
var accountApprovedTemplate string

// EmailService handles email sending operations.
type EmailService struct {
	host string
	port string
	from string
}

// NewEmailService creates a new email service instance.
func NewEmailService(host, port, from string) *EmailService {
	return &EmailService{
		host: host,
		port: port,
		from: from,
	}
}

// SendWelcomeEmail sends registration pending notification.
func (s *EmailService) SendWelcomeEmail(toEmail, userName string) error {
	subject := "Welcome to Point of Sale — Registration Pending"
	data := map[string]string{
		"UserName": userName,
	}
	return s.sendEmail(toEmail, subject, welcomeTemplate, data)
}

// SendPasswordResetEmail sends password reset link.
func (s *EmailService) SendPasswordResetEmail(toEmail, userName, resetLink string) error {
	subject := "Point of Sale — Password Reset"
	data := map[string]string{
		"UserName":  userName,
		"ResetLink": resetLink,
	}
	return s.sendEmail(toEmail, subject, passwordResetTemplate, data)
}

// SendAccountApprovedEmail sends account approval notification.
func (s *EmailService) SendAccountApprovedEmail(toEmail, userName string) error {
	subject := "Point of Sale — Account Approved"
	data := map[string]string{
		"UserName": userName,
	}
	return s.sendEmail(toEmail, subject, accountApprovedTemplate, data)
}

// sendEmail is a generic email sending function.
func (s *EmailService) sendEmail(to, subject, templateStr string, data map[string]string) error {
	// Parse template
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Build message
	message := s.buildMessage(to, subject, body.String())

	// For Mailpit in development, no authentication is needed
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	err = smtp.SendMail(addr, nil, s.from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildMessage constructs the email message with headers and body.
func (s *EmailService) buildMessage(to, subject, htmlBody string) string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	return msg.String()
}
