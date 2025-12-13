package service

import (
	"errors"
	"net/smtp"
	"testing"

	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestEmailService_SendOTP(t *testing.T) {
	cfg := &config.SMTPConfig{
		Host:      "smtp.example.com",
		Port:      587,
		Username:  "user",
		Password:  "pass",
		FromEmail: "from@example.com",
		FromName:  "Auth Gateway",
	}
	svc := NewEmailService(cfg)

	t.Run("Success", func(t *testing.T) {
		svc.sendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			assert.Equal(t, "smtp.example.com:587", addr)
			assert.Equal(t, "from@example.com", from)
			assert.Equal(t, []string{"to@example.com"}, to)
			assert.Contains(t, string(msg), "Subject: Your Verification Code")
			return nil
		}

		err := svc.SendOTP("to@example.com", "123456", "verification")
		assert.NoError(t, err)
	})
}

func TestEmailService_SendWelcome(t *testing.T) {
	cfg := &config.SMTPConfig{
		Host:      "smtp.example.com",
		Port:      587,
		Username:  "user",
		Password:  "pass",
		FromEmail: "from@example.com",
		FromName:  "Auth Gateway",
	}
	svc := NewEmailService(cfg)

	t.Run("Success", func(t *testing.T) {
		svc.sendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			assert.Contains(t, string(msg), "Subject: Welcome to Auth Gateway!")
			return nil
		}

		err := svc.SendWelcome("to@example.com", "testuser")
		assert.NoError(t, err)
	})
}

func TestEmailService_Send(t *testing.T) {
	cfg := &config.SMTPConfig{
		Host:      "smtp.example.com",
		Port:      587,
		Username:  "user",
		Password:  "pass",
		FromEmail: "from@example.com",
		FromName:  "Auth Gateway",
	}
	svc := NewEmailService(cfg)

	t.Run("Success", func(t *testing.T) {
		svc.sendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			return nil
		}
		err := svc.Send("to@example.com", "Subject", "Body")
		assert.NoError(t, err)
	})

	t.Run("SMTPError", func(t *testing.T) {
		svc.sendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			return errors.New("smtp error")
		}
		err := svc.Send("to@example.com", "Subject", "Body")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send email")
	})

	t.Run("NoConfig", func(t *testing.T) {
		svcNoConfig := NewEmailService(&config.SMTPConfig{}) // Empty config
		// Should log and return nil
		err := svcNoConfig.Send("to@example.com", "Subject", "Body")
		assert.NoError(t, err)
	})
}
