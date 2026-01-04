package common

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
	"time"

	"github.com/multimarket-labs/event-pod-services/config"
)

type EmailMessage struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []string
}

type EmailService struct {
	config *config.EmailConfig
}

func NewEmailService(cfg *config.EmailConfig) (*EmailService, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid email config: %w", err)
	}
	return &EmailService{config: cfg}, nil
}

func (s *EmailService) SendEmail(msg *EmailMessage) error {
	if err := validateMessage(msg); err != nil {
		return fmt.Errorf("invalid email message: %w", err)
	}

	message := s.buildMessage(msg)

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	var auth smtp.Auth
	if s.config.SMTPUser != "" && s.config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	}

	recipients := append(msg.To, msg.Cc...)
	recipients = append(recipients, msg.Bcc...)

	if s.config.UseSSL {
		return s.sendMailTLS(addr, auth, s.config.FromEmail, recipients, []byte(message))
	}

	return smtp.SendMail(addr, auth, s.config.FromEmail, recipients, []byte(message))
}

func (s *EmailService) sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName:         s.config.SMTPHost,
		InsecureSkipVerify: false,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Quit()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to add recipient: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

func (s *EmailService) buildMessage(msg *EmailMessage) string {
	var builder strings.Builder

	fromName := s.config.FromName
	if fromName == "" {
		fromName = s.config.FromEmail
	}
	builder.WriteString(fmt.Sprintf("From: %s <%s>\r\n", mimeQEncode(fromName), s.config.FromEmail))

	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	if len(msg.Cc) > 0 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.Cc, ", ")))
	}

	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", mimeQEncode(msg.Subject)))

	builder.WriteString("MIME-Version: 1.0\r\n")

	if msg.HTMLBody != "" {
		boundary := "boundary-" + generateBoundary()
		builder.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

		// Plain text part
		if msg.Body != "" {
			builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
			builder.WriteString(msg.Body)
			builder.WriteString("\r\n\r\n")
		}

		// HTML part
		builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		builder.WriteString(msg.HTMLBody)
		builder.WriteString("\r\n\r\n")

		builder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		builder.WriteString(msg.Body)
	}

	return builder.String()
}

func generateBoundary() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func validateConfig(config *config.EmailConfig) error {
	if config.SMTPHost == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if config.SMTPPort <= 0 || config.SMTPPort > 65535 {
		return fmt.Errorf("invalid SMTP port")
	}
	if config.FromEmail == "" {
		return fmt.Errorf("sender email is required")
	}
	return nil
}

func validateMessage(msg *EmailMessage) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if msg.Subject == "" {
		return fmt.Errorf("email subject is required")
	}
	if msg.Body == "" && msg.HTMLBody == "" {
		return fmt.Errorf("email body is required")
	}
	return nil
}

func mimeQEncode(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}
