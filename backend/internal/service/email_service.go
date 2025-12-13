package service

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strconv"

	"github.com/smilemakc/auth-gateway/internal/config"
)

// EmailService handles email sending
type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.SMTPConfig) *EmailService {
	return &EmailService{
		smtpHost:     cfg.Host,
		smtpPort:     strconv.Itoa(cfg.Port),
		smtpUsername: cfg.Username,
		smtpPassword: cfg.Password,
		fromEmail:    cfg.FromEmail,
		fromName:     cfg.FromName,
	}
}

// SendOTP sends an OTP code via email
func (s *EmailService) SendOTP(to, code, otpType string) error {
	subject := "Your Verification Code"
	bodyTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { background: #f9fafb; padding: 30px; }
        .code { font-size: 32px; font-weight: bold; color: #4F46E5; text-align: center; padding: 20px; background: white; border-radius: 8px; letter-spacing: 5px; }
        .footer { text-align: center; padding: 20px; color: #6b7280; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        <div class="content">
            <p>Hello,</p>
            <p>{{.Message}}</p>
            <div class="code">{{.Code}}</div>
            <p><strong>This code will expire in 10 minutes.</strong></p>
            <p>If you didn't request this code, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply.</p>
            <p>&copy; 2024 Auth Gateway. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	data := map[string]string{
		"Code": code,
	}

	switch otpType {
	case "verification":
		data["Title"] = "Email Verification"
		data["Message"] = "Please use the following code to verify your email address:"
	case "password_reset":
		data["Title"] = "Password Reset"
		data["Message"] = "Please use the following code to reset your password:"
		subject = "Password Reset Code"
	case "2fa":
		data["Title"] = "Two-Factor Authentication"
		data["Message"] = "Please use the following code to complete your login:"
		subject = "2FA Code"
	case "login":
		data["Title"] = "Login Code"
		data["Message"] = "Please use the following code to log in:"
		subject = "Login Code"
	default:
		data["Title"] = "Verification Code"
		data["Message"] = "Please use the following code:"
	}

	tmpl, err := template.New("email").Parse(bodyTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	return s.Send(to, subject, body.String())
}

// Send sends an email
func (s *EmailService) Send(to, subject, htmlBody string) error {
	// If SMTP is not configured, just log (for development)
	if s.smtpUsername == "" || s.smtpPassword == "" {
		fmt.Printf("\n=== EMAIL (SMTP not configured, logging instead) ===\n")
		fmt.Printf("To: %s\n", to)
		fmt.Printf("Subject: %s\n", subject)
		fmt.Printf("Body: %s\n", htmlBody)
		fmt.Printf("===============================================\n\n")
		return nil
	}

	// Prepare message
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			htmlBody,
	)

	// SMTP authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Send email
	addr := s.smtpHost + ":" + s.smtpPort
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendWelcome sends a welcome email
func (s *EmailService) SendWelcome(to, username string) error {
	subject := "Welcome to Auth Gateway!"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { background: #f9fafb; padding: 30px; }
        .footer { text-align: center; padding: 20px; color: #6b7280; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to Auth Gateway!</h1>
        </div>
        <div class="content">
            <p>Hello %s,</p>
            <p>Welcome to Auth Gateway! Your account has been successfully created.</p>
            <p>You can now enjoy all the features of our platform.</p>
            <p>If you have any questions, feel free to contact our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Auth Gateway. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, username)

	return s.Send(to, subject, body)
}
