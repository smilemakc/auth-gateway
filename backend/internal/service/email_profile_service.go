package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/smtp"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
)

type EmailProfileService struct {
	providerRepo    *repository.EmailProviderRepository
	profileRepo     *repository.EmailProfileRepository
	templateService *TemplateService
	auditService    AuditLogger
	cfg             *config.Config
	fallbackEmail   EmailSender
}

func NewEmailProfileService(
	providerRepo *repository.EmailProviderRepository,
	profileRepo *repository.EmailProfileRepository,
	templateService *TemplateService,
	auditService AuditLogger,
	cfg *config.Config,
	fallbackEmail EmailSender,
) *EmailProfileService {
	return &EmailProfileService{
		providerRepo:    providerRepo,
		profileRepo:     profileRepo,
		templateService: templateService,
		auditService:    auditService,
		cfg:             cfg,
		fallbackEmail:   fallbackEmail,
	}
}

func (s *EmailProfileService) CreateProvider(ctx context.Context, req *models.CreateEmailProviderRequest) (*models.EmailProvider, error) {
	provider := &models.EmailProvider{
		Name:               req.Name,
		Type:               req.Type,
		IsActive:           req.IsActive,
		SMTPHost:           req.SMTPHost,
		SMTPPort:           req.SMTPPort,
		SMTPUsername:       req.SMTPUsername,
		SMTPPassword:       req.SMTPPassword,
		SMTPUseTLS:         req.SMTPUseTLS,
		SendGridAPIKey:     req.SendGridAPIKey,
		SESRegion:          req.SESRegion,
		SESAccessKeyID:     req.SESAccessKeyID,
		SESSecretAccessKey: req.SESSecretAccessKey,
		MailgunDomain:      req.MailgunDomain,
		MailgunAPIKey:      req.MailgunAPIKey,
	}

	if err := s.providerRepo.Create(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionCreate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProvider,
			"provider_id":   provider.ID.String(),
			"provider_name": provider.Name,
			"provider_type": provider.Type,
		},
	})

	return provider, nil
}

func (s *EmailProfileService) GetProvider(ctx context.Context, id uuid.UUID) (*models.EmailProviderResponse, error) {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.maskProviderSecrets(provider), nil
}

func (s *EmailProfileService) ListProviders(ctx context.Context) ([]*models.EmailProviderResponse, error) {
	providers, err := s.providerRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	responses := make([]*models.EmailProviderResponse, len(providers))
	for i, provider := range providers {
		responses[i] = s.maskProviderSecrets(provider)
	}

	return responses, nil
}

func (s *EmailProfileService) UpdateProvider(ctx context.Context, id uuid.UUID, req *models.UpdateEmailProviderRequest) error {
	existing, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.SMTPHost != nil {
		existing.SMTPHost = req.SMTPHost
	}
	if req.SMTPPort != nil {
		existing.SMTPPort = req.SMTPPort
	}
	if req.SMTPUsername != nil {
		existing.SMTPUsername = req.SMTPUsername
	}
	if req.SMTPPassword != nil {
		existing.SMTPPassword = req.SMTPPassword
	}
	if req.SMTPUseTLS != nil {
		existing.SMTPUseTLS = req.SMTPUseTLS
	}
	if req.SendGridAPIKey != nil {
		existing.SendGridAPIKey = req.SendGridAPIKey
	}
	if req.SESRegion != nil {
		existing.SESRegion = req.SESRegion
	}
	if req.SESAccessKeyID != nil {
		existing.SESAccessKeyID = req.SESAccessKeyID
	}
	if req.SESSecretAccessKey != nil {
		existing.SESSecretAccessKey = req.SESSecretAccessKey
	}
	if req.MailgunDomain != nil {
		existing.MailgunDomain = req.MailgunDomain
	}
	if req.MailgunAPIKey != nil {
		existing.MailgunAPIKey = req.MailgunAPIKey
	}

	if err := s.providerRepo.Update(ctx, id, existing); err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionUpdate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProvider,
			"provider_id":   id.String(),
		},
	})

	return nil
}

func (s *EmailProfileService) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.providerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionDelete,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProvider,
			"provider_id":   id.String(),
			"provider_name": provider.Name,
		},
	})

	return nil
}

func (s *EmailProfileService) TestProvider(ctx context.Context, id uuid.UUID, testEmail string) error {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !provider.IsActive {
		return models.NewAppError(http.StatusBadRequest, "provider is not active")
	}

	subject := "Test Email from Auth Gateway"
	htmlBody := `<html><body><h1>Test Email</h1><p>This is a test email from your email provider configuration.</p></body></html>`

	var sendErr error
	switch provider.Type {
	case models.EmailProviderTypeSMTP:
		sendErr = s.sendViaSMTP(provider, testEmail, testEmail, "Auth Gateway", subject, htmlBody)
	case models.EmailProviderTypeSendGrid:
		sendErr = s.sendViaSendGrid(provider, testEmail, testEmail, "Auth Gateway", subject, htmlBody)
	case models.EmailProviderTypeMailgun:
		sendErr = s.sendViaMailgun(provider, testEmail, testEmail, "Auth Gateway", subject, htmlBody)
	case models.EmailProviderTypeSES:
		sendErr = s.sendViaSES(provider, testEmail, testEmail, "Auth Gateway", subject, htmlBody)
	default:
		return models.NewAppError(http.StatusBadRequest, "unsupported provider type")
	}

	if sendErr != nil {
		s.auditService.Log(AuditLogParams{
			Action: models.ActionTest,
			Status: models.StatusFailure,
			Details: map[string]interface{}{
				"resource_type": models.ResourceEmailProvider,
				"provider_id":   id.String(),
				"test_email":    testEmail,
				"error":         sendErr.Error(),
			},
		})
		return fmt.Errorf("failed to send test email: %w", sendErr)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionTest,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProvider,
			"provider_id":   id.String(),
			"test_email":    testEmail,
		},
	})

	return nil
}

func (s *EmailProfileService) CreateProfile(ctx context.Context, req *models.CreateEmailProfileRequest) (*models.EmailProfile, error) {
	_, err := s.providerRepo.GetByID(ctx, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	if req.IsDefault {
		if err := s.profileRepo.ClearDefault(ctx); err != nil {
			return nil, fmt.Errorf("failed to clear default flags: %w", err)
		}
	}

	profile := &models.EmailProfile{
		Name:       req.Name,
		ProviderID: req.ProviderID,
		FromEmail:  req.FromEmail,
		FromName:   req.FromName,
		ReplyTo:    req.ReplyTo,
		IsDefault:  req.IsDefault,
		IsActive:   req.IsActive,
	}

	if err := s.profileRepo.Create(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionCreate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProfile,
			"profile_id":    profile.ID.String(),
			"profile_name":  profile.Name,
			"is_default":    profile.IsDefault,
		},
	})

	return profile, nil
}

func (s *EmailProfileService) GetProfile(ctx context.Context, id uuid.UUID) (*models.EmailProfile, error) {
	profile, err := s.profileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *EmailProfileService) ListProfiles(ctx context.Context) ([]*models.EmailProfile, error) {
	profiles, err := s.profileRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}
	return profiles, nil
}

func (s *EmailProfileService) UpdateProfile(ctx context.Context, id uuid.UUID, req *models.UpdateEmailProfileRequest) error {
	existing, err := s.profileRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if req.ProviderID != nil {
		_, err := s.providerRepo.GetByID(ctx, *req.ProviderID)
		if err != nil {
			return fmt.Errorf("provider not found: %w", err)
		}
		existing.ProviderID = *req.ProviderID
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.FromEmail != nil {
		existing.FromEmail = *req.FromEmail
	}
	if req.FromName != nil {
		existing.FromName = *req.FromName
	}
	if req.ReplyTo != nil {
		existing.ReplyTo = req.ReplyTo
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if req.IsDefault != nil && *req.IsDefault {
		if err := s.profileRepo.ClearDefault(ctx); err != nil {
			return fmt.Errorf("failed to clear default flags: %w", err)
		}
		existing.IsDefault = true
	}

	if err := s.profileRepo.Update(ctx, id, existing); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionUpdate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProfile,
			"profile_id":    id.String(),
		},
	})

	return nil
}

func (s *EmailProfileService) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	profile, err := s.profileRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.profileRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionDelete,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProfile,
			"profile_id":    id.String(),
			"profile_name":  profile.Name,
		},
	})

	return nil
}

func (s *EmailProfileService) SetDefaultProfile(ctx context.Context, id uuid.UUID) error {
	if err := s.profileRepo.SetDefault(ctx, id); err != nil {
		return fmt.Errorf("failed to set default profile: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionUpdate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProfile,
			"profile_id":    id.String(),
			"action":        "set_default",
		},
	})

	return nil
}

func (s *EmailProfileService) GetProfileTemplates(ctx context.Context, profileID uuid.UUID) ([]*models.EmailProfileTemplate, error) {
	templates, err := s.profileRepo.GetTemplatesForProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile templates: %w", err)
	}
	return templates, nil
}

func (s *EmailProfileService) SetProfileTemplate(ctx context.Context, profileID uuid.UUID, otpType string, templateID uuid.UUID) error {
	_, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	_, err = s.templateService.GetEmailTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	if err := s.profileRepo.SetTemplateForOTPType(ctx, profileID, otpType, templateID); err != nil {
		return fmt.Errorf("failed to set profile template: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionUpdate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProfileTemplate,
			"profile_id":    profileID.String(),
			"otp_type":      otpType,
			"template_id":   templateID.String(),
		},
	})

	return nil
}

func (s *EmailProfileService) RemoveProfileTemplate(ctx context.Context, profileID uuid.UUID, otpType string) error {
	if err := s.profileRepo.RemoveTemplateForOTPType(ctx, profileID, otpType); err != nil {
		return fmt.Errorf("failed to remove profile template: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionDelete,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailProfileTemplate,
			"profile_id":    profileID.String(),
			"otp_type":      otpType,
		},
	})

	return nil
}

func (s *EmailProfileService) SendOTPEmail(ctx context.Context, profileID *uuid.UUID, toEmail string, otpType models.OTPType, code string) error {
	var profile *models.EmailProfile
	var err error

	if profileID == nil {
		profile, err = s.profileRepo.GetDefault(ctx)
		if err != nil {
			if err == models.ErrNotFound {
				if s.fallbackEmail != nil {
					return s.fallbackEmail.SendOTP(toEmail, code, string(otpType))
				}
				return models.NewAppError(http.StatusNotFound, "no default email profile configured")
			}
			return fmt.Errorf("failed to get default profile: %w", err)
		}
	} else {
		profile, err = s.profileRepo.GetByID(ctx, *profileID)
		if err != nil {
			return fmt.Errorf("profile not found: %w", err)
		}
	}

	if !profile.IsActive {
		return models.NewAppError(http.StatusBadRequest, "profile is not active")
	}

	if profile.Provider == nil {
		provider, err := s.providerRepo.GetByID(ctx, profile.ProviderID)
		if err != nil {
			return fmt.Errorf("provider not found: %w", err)
		}
		profile.Provider = provider
	}

	if !profile.Provider.IsActive {
		return models.NewAppError(http.StatusBadRequest, "provider is not active")
	}

	var subject, htmlBody string

	profileTemplate, err := s.profileRepo.GetTemplateForOTPType(ctx, profile.ID, string(otpType))
	if err != nil && err != models.ErrNotFound {
		return fmt.Errorf("failed to get template: %w", err)
	}

	variables := map[string]interface{}{
		"code":           code,
		"expiry_minutes": 10,
	}

	if profileTemplate != nil && profileTemplate.Template != nil {
		var textBody string
		subject, htmlBody, textBody, err = s.templateService.RenderTemplate(ctx, string(otpType), variables)
		_ = textBody
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
	} else {
		var textBody string
		subject, htmlBody, textBody, err = s.templateService.RenderTemplate(ctx, string(otpType), variables)
		_ = textBody
		if err != nil {
			subject = s.getDefaultSubject(otpType)
			htmlBody = s.getDefaultHTMLBody(otpType, code)
		}
	}

	var sendErr error
	switch profile.Provider.Type {
	case models.EmailProviderTypeSMTP:
		sendErr = s.sendViaSMTP(profile.Provider, toEmail, profile.FromEmail, profile.FromName, subject, htmlBody)
	case models.EmailProviderTypeSendGrid:
		sendErr = s.sendViaSendGrid(profile.Provider, toEmail, profile.FromEmail, profile.FromName, subject, htmlBody)
	case models.EmailProviderTypeMailgun:
		sendErr = s.sendViaMailgun(profile.Provider, toEmail, profile.FromEmail, profile.FromName, subject, htmlBody)
	case models.EmailProviderTypeSES:
		sendErr = s.sendViaSES(profile.Provider, toEmail, profile.FromEmail, profile.FromName, subject, htmlBody)
	default:
		sendErr = fmt.Errorf("unsupported provider type: %s", profile.Provider.Type)
	}

	now := time.Now()
	logStatus := models.EmailStatusSent
	var errorMsg *string

	if sendErr != nil {
		logStatus = models.EmailStatusFailed
		errStr := sendErr.Error()
		errorMsg = &errStr
	}

	log := &models.EmailLog{
		ProfileID:      profile.ID,
		RecipientEmail: toEmail,
		Subject:        subject,
		TemplateType:   string(otpType),
		ProviderType:   profile.Provider.Type,
		Status:         logStatus,
		ErrorMessage:   errorMsg,
		CreatedAt:      now,
	}

	if sendErr == nil {
		log.SentAt = &now
	}

	if createLogErr := s.profileRepo.CreateLog(ctx, log); createLogErr != nil {
		s.auditService.Log(AuditLogParams{
			Action: models.ActionCreate,
			Status: models.StatusFailure,
			Details: map[string]interface{}{
				"resource_type": models.ResourceEmailLog,
				"error":         createLogErr.Error(),
			},
		})
	}

	if sendErr != nil {
		s.auditService.Log(AuditLogParams{
			Action: models.ActionSend,
			Status: models.StatusFailure,
			Details: map[string]interface{}{
				"resource_type": models.ResourceEmail,
				"profile_id":    profile.ID.String(),
				"recipient":     toEmail,
				"otp_type":      otpType,
				"error":         sendErr.Error(),
			},
		})
		return fmt.Errorf("failed to send email: %w", sendErr)
	}

	s.auditService.Log(AuditLogParams{
		Action: models.ActionSend,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmail,
			"profile_id":    profile.ID.String(),
			"recipient":     toEmail,
			"otp_type":      otpType,
		},
	})

	return nil
}

func (s *EmailProfileService) sendViaSMTP(provider *models.EmailProvider, to, from, fromName, subject, htmlBody string) error {
	if provider.SMTPHost == nil || provider.SMTPPort == nil {
		return fmt.Errorf("SMTP configuration incomplete")
	}

	host := *provider.SMTPHost
	port := *provider.SMTPPort
	addr := fmt.Sprintf("%s:%d", host, port)

	var auth smtp.Auth
	if provider.SMTPUsername != nil && provider.SMTPPassword != nil {
		auth = smtp.PlainAuth("", *provider.SMTPUsername, *provider.SMTPPassword, host)
	}

	fromHeader := fmt.Sprintf("%s <%s>", fromName, from)
	msg := []byte(
		"From: " + fromHeader + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			htmlBody,
	)

	useTLS := provider.SMTPUseTLS != nil && *provider.SMTPUseTLS

	if useTLS {
		tlsConfig := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: false,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect with TLS: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP auth failed: %w", err)
			}
		}

		if err := client.Mail(from); err != nil {
			return fmt.Errorf("SMTP MAIL failed: %w", err)
		}

		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("SMTP RCPT failed: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("SMTP DATA failed: %w", err)
		}

		_, err = w.Write(msg)
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close data writer: %w", err)
		}

		return nil
	}

	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func (s *EmailProfileService) sendViaSendGrid(provider *models.EmailProvider, to, from, fromName, subject, htmlBody string) error {
	return fmt.Errorf("SendGrid integration not yet implemented")
}

func (s *EmailProfileService) sendViaMailgun(provider *models.EmailProvider, to, from, fromName, subject, htmlBody string) error {
	return fmt.Errorf("Mailgun integration not yet implemented")
}

func (s *EmailProfileService) sendViaSES(provider *models.EmailProvider, to, from, fromName, subject, htmlBody string) error {
	return fmt.Errorf("AWS SES integration not yet implemented")
}

func (s *EmailProfileService) maskProviderSecrets(provider *models.EmailProvider) *models.EmailProviderResponse {
	response := &models.EmailProviderResponse{
		ID:             provider.ID,
		Name:           provider.Name,
		Type:           provider.Type,
		IsActive:       provider.IsActive,
		CreatedAt:      provider.CreatedAt,
		UpdatedAt:      provider.UpdatedAt,
		CreatedBy:      provider.CreatedBy,
		SMTPHost:       provider.SMTPHost,
		SMTPPort:       provider.SMTPPort,
		SMTPUsername:   provider.SMTPUsername,
		SMTPUseTLS:     provider.SMTPUseTLS,
		SESRegion:      provider.SESRegion,
		SESAccessKeyID: provider.SESAccessKeyID,
		MailgunDomain:  provider.MailgunDomain,
	}

	response.HasSMTPPassword = provider.SMTPPassword != nil && *provider.SMTPPassword != ""
	response.HasSendGridAPIKey = provider.SendGridAPIKey != nil && *provider.SendGridAPIKey != ""
	response.HasSESSecretAccessKey = provider.SESSecretAccessKey != nil && *provider.SESSecretAccessKey != ""
	response.HasMailgunAPIKey = provider.MailgunAPIKey != nil && *provider.MailgunAPIKey != ""

	return response
}

func (s *EmailProfileService) getDefaultSubject(otpType models.OTPType) string {
	switch otpType {
	case models.OTPTypeVerification:
		return "Email Verification Code"
	case models.OTPTypePasswordReset:
		return "Password Reset Code"
	case models.OTPType2FA:
		return "Two-Factor Authentication Code"
	default:
		return "Verification Code"
	}
}

func (s *EmailProfileService) getDefaultHTMLBody(otpType models.OTPType, code string) string {
	var title, message string

	switch otpType {
	case models.OTPTypeVerification:
		title = "Email Verification"
		message = "Please use the following code to verify your email address:"
	case models.OTPTypePasswordReset:
		title = "Password Reset"
		message = "Please use the following code to reset your password:"
	case models.OTPType2FA:
		title = "Two-Factor Authentication"
		message = "Please use the following code to complete your login:"
	default:
		title = "Verification Code"
		message = "Please use the following code:"
	}

	return fmt.Sprintf(`
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
            <h1>%s</h1>
        </div>
        <div class="content">
            <p>Hello,</p>
            <p>%s</p>
            <div class="code">%s</div>
            <p><strong>This code will expire in 10 minutes.</strong></p>
            <p>If you didn't request this code, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply.</p>
            <p>&copy; 2025 Auth Gateway. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, title, message, code)
}

// GetProfileStats retrieves email sending statistics for a profile
func (s *EmailProfileService) GetProfileStats(ctx context.Context, profileID uuid.UUID) (*models.EmailStatsResponse, error) {
	// Verify profile exists
	profile, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	if profile == nil {
		return nil, models.NewAppError(http.StatusNotFound, "Email profile not found")
	}

	// Get stats from repository (global stats, could be filtered by profile in future)
	return s.profileRepo.GetStats(ctx)
}

// TestProfile sends a test email using the specified profile
func (s *EmailProfileService) TestProfile(ctx context.Context, profileID uuid.UUID, testEmail string) error {
	// Get profile with provider
	profile, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}
	if profile == nil {
		return models.NewAppError(http.StatusNotFound, "Email profile not found")
	}

	// Get the provider
	provider, err := s.providerRepo.GetByID(ctx, profile.ProviderID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	if provider == nil {
		return models.NewAppError(http.StatusNotFound, "Email provider not found")
	}

	subject := "Test Email from Auth Gateway"
	htmlBody := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; }
        .content { background: #f9fafb; padding: 30px; }
        .success { font-size: 24px; color: #10B981; text-align: center; }
        .footer { text-align: center; padding: 20px; color: #6b7280; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Test Email</h1>
        </div>
        <div class="content">
            <p class="success">✓ Email Profile Test Successful</p>
            <p>This is a test email to verify your email profile configuration.</p>
            <p><strong>Profile:</strong> ` + profile.Name + `</p>
            <p><strong>From:</strong> ` + profile.FromName + ` &lt;` + profile.FromEmail + `&gt;</p>
        </div>
        <div class="footer">
            <p>This is an automated test message from Auth Gateway.</p>
        </div>
    </div>
</body>
</html>`

	var sendErr error
	switch provider.Type {
	case "smtp":
		sendErr = s.sendViaSMTP(provider, testEmail, profile.FromEmail, profile.FromName, subject, htmlBody)
	case "sendgrid":
		sendErr = s.sendViaSendGrid(provider, testEmail, profile.FromEmail, profile.FromName, subject, htmlBody)
	case "mailgun":
		sendErr = s.sendViaMailgun(provider, testEmail, profile.FromEmail, profile.FromName, subject, htmlBody)
	case "ses":
		sendErr = s.sendViaSES(provider, testEmail, profile.FromEmail, profile.FromName, subject, htmlBody)
	default:
		return models.NewAppError(http.StatusBadRequest, "Unsupported provider type: "+provider.Type)
	}

	// Log the test email
	status := models.EmailStatusSent
	var errMsg *string
	if sendErr != nil {
		status = models.EmailStatusFailed
		errStr := sendErr.Error()
		errMsg = &errStr
	}

	now := time.Now()
	log := &models.EmailLog{
		ProfileID:      profileID,
		RecipientEmail: testEmail,
		Subject:        subject,
		TemplateType:   "test",
		ProviderType:   provider.Type,
		Status:         status,
		ErrorMessage:   errMsg,
		SentAt:         &now,
	}
	_ = s.profileRepo.CreateLog(ctx, log)

	// Audit log
	s.auditService.Log(AuditLogParams{
		Action: models.ActionTest,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"profile_id": profileID.String(),
			"email":      testEmail,
		},
	})

	return sendErr
}
