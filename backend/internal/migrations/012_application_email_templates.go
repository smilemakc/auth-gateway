package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Adding application-scoped email templates...")
		return addApplicationEmailTemplates(ctx, db)
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Removing application email templates support...")
		return removeApplicationEmailTemplates(ctx, db)
	})
}

func addApplicationEmailTemplates(ctx context.Context, db *bun.DB) error {
	// ============================================================
	// 1. Add application_id column to email_templates
	// ============================================================
	_, err := db.ExecContext(ctx, `
		ALTER TABLE email_templates
		ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id) ON DELETE CASCADE
	`)
	if err != nil {
		return fmt.Errorf("failed to add application_id to email_templates: %w", err)
	}

	// ============================================================
	// 2. Add application_id column to email_template_versions
	// ============================================================
	_, err = db.ExecContext(ctx, `
		ALTER TABLE email_template_versions
		ADD COLUMN IF NOT EXISTS application_id UUID
	`)
	if err != nil {
		return fmt.Errorf("failed to add application_id to email_template_versions: %w", err)
	}

	// ============================================================
	// 3. Create indexes for efficient lookups
	// ============================================================
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_email_templates_application_id ON email_templates(application_id) WHERE application_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_email_templates_app_type ON email_templates(application_id, type) WHERE application_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_email_template_versions_app ON email_template_versions(application_id) WHERE application_id IS NOT NULL",
	}

	for _, indexSQL := range indexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w (SQL: %s)", err, indexSQL)
		}
	}

	// ============================================================
	// 4. Add table and column comments
	// ============================================================
	comments := []string{
		"COMMENT ON COLUMN email_templates.application_id IS 'Links email template to a specific application for multi-app support'",
		"COMMENT ON COLUMN email_template_versions.application_id IS 'Application scope for template version history'",
	}

	for _, commentSQL := range comments {
		if _, err := db.ExecContext(ctx, commentSQL); err != nil {
			return fmt.Errorf("failed to add comment: %w (SQL: %s)", err, commentSQL)
		}
	}

	// ============================================================
	// 5. Seed default templates for each existing application
	// ============================================================

	// Define default templates for each type
	templateDefinitions := []struct {
		Type      string
		Name      string
		Subject   string
		HTMLBody  string
		Variables []string
	}{
		{
			Type:    "verification",
			Name:    "Email Verification",
			Subject: "Verify Your Email Address",
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #3B82F6; }
        .header h1 { color: #3B82F6; margin: 0; font-size: 28px; }
        .content { margin-bottom: 30px; }
        .code-box { background-color: #f8f9fa; border: 2px dashed #3B82F6; border-radius: 6px; padding: 20px; text-align: center; margin: 25px 0; }
        .code { font-size: 32px; font-weight: bold; color: #3B82F6; letter-spacing: 5px; font-family: 'Courier New', monospace; }
        .expiry { color: #6b7280; font-size: 14px; margin-top: 10px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
        .button { display: inline-block; padding: 12px 30px; background-color: #3B82F6; color: #ffffff; text-decoration: none; border-radius: 6px; margin: 20px 0; font-weight: bold; }
        .warning { background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Email Verification</h1>
        </div>
        <div class="content">
            <p>Hello <strong>{{.username}}</strong>,</p>
            <p>Thank you for signing up! Please verify your email address by using the verification code below:</p>
            <div class="code-box">
                <div class="code">{{.code}}</div>
                <div class="expiry">This code expires in {{.expiry_minutes}} minutes</div>
            </div>
            <div class="warning">
                <strong>Security Notice:</strong> If you didn't request this verification code, please ignore this email or contact support if you have concerns.
            </div>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>&copy; 2024 Auth Gateway. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: []string{"username", "email", "code", "expiry_minutes"},
		},
		{
			Type:    "password_reset",
			Name:    "Password Reset",
			Subject: "Reset Your Password",
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #ef4444; }
        .header h1 { color: #ef4444; margin: 0; font-size: 28px; }
        .content { margin-bottom: 30px; }
        .code-box { background-color: #fef2f2; border: 2px dashed #ef4444; border-radius: 6px; padding: 20px; text-align: center; margin: 25px 0; }
        .code { font-size: 32px; font-weight: bold; color: #ef4444; letter-spacing: 5px; font-family: 'Courier New', monospace; }
        .expiry { color: #6b7280; font-size: 14px; margin-top: 10px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
        .warning { background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <p>Hello <strong>{{.username}}</strong>,</p>
            <p>We received a request to reset your password. Use the code below to complete the reset process:</p>
            <div class="code-box">
                <div class="code">{{.code}}</div>
                <div class="expiry">This code expires in {{.expiry_minutes}} minutes</div>
            </div>
            <div class="warning">
                <strong>Security Warning:</strong> If you didn't request a password reset, please ignore this email and ensure your account is secure. Consider changing your password if you suspect unauthorized access.
            </div>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>&copy; 2024 Auth Gateway. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: []string{"username", "email", "code", "expiry_minutes"},
		},
		{
			Type:    "welcome",
			Name:    "Welcome",
			Subject: "Welcome to Our Platform!",
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #10b981; }
        .header h1 { color: #10b981; margin: 0; font-size: 28px; }
        .content { margin-bottom: 30px; }
        .welcome-box { background: linear-gradient(135deg, #10b981 0%, #059669 100%); border-radius: 8px; padding: 30px; text-align: center; color: #ffffff; margin: 25px 0; }
        .welcome-box h2 { margin: 0 0 10px 0; font-size: 24px; }
        .feature-list { background-color: #f8f9fa; border-radius: 6px; padding: 20px; margin: 20px 0; }
        .feature-list ul { margin: 0; padding-left: 20px; }
        .feature-list li { margin: 10px 0; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
        .button { display: inline-block; padding: 12px 30px; background-color: #10b981; color: #ffffff; text-decoration: none; border-radius: 6px; margin: 20px 0; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome Aboard!</h1>
        </div>
        <div class="content">
            <div class="welcome-box">
                <h2>Hello {{.username}}!</h2>
                <p>We're excited to have you join our community</p>
            </div>
            <p>Thank you for creating an account with us. You're now part of our growing community!</p>
            <div class="feature-list">
                <h3>Getting Started:</h3>
                <ul>
                    <li>Complete your profile to personalize your experience</li>
                    <li>Explore our features and discover what we offer</li>
                    <li>Connect with other members of our community</li>
                    <li>Check out our help center for tips and tutorials</li>
                </ul>
            </div>
            <p>If you have any questions or need assistance, our support team is here to help!</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>&copy; 2024 Auth Gateway. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: []string{"username", "email", "full_name"},
		},
		{
			Type:    "2fa",
			Name:    "Two-Factor Authentication",
			Subject: "Your Two-Factor Authentication Code",
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #8b5cf6; }
        .header h1 { color: #8b5cf6; margin: 0; font-size: 28px; }
        .content { margin-bottom: 30px; }
        .code-box { background-color: #faf5ff; border: 2px dashed #8b5cf6; border-radius: 6px; padding: 20px; text-align: center; margin: 25px 0; }
        .code { font-size: 32px; font-weight: bold; color: #8b5cf6; letter-spacing: 5px; font-family: 'Courier New', monospace; }
        .expiry { color: #6b7280; font-size: 14px; margin-top: 10px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
        .security-icon { text-align: center; font-size: 48px; margin: 20px 0; }
        .warning { background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Two-Factor Authentication</h1>
        </div>
        <div class="content">
            <div class="security-icon">🔐</div>
            <p>Hello <strong>{{.username}}</strong>,</p>
            <p>Here is your two-factor authentication code to complete your login:</p>
            <div class="code-box">
                <div class="code">{{.code}}</div>
                <div class="expiry">This code expires in {{.expiry_minutes}} minutes</div>
            </div>
            <div class="warning">
                <strong>Security Notice:</strong> Never share this code with anyone. Our team will never ask you for this code.
            </div>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>&copy; 2024 Auth Gateway. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: []string{"username", "email", "code", "expiry_minutes"},
		},
		{
			Type:    "otp_login",
			Name:    "Login OTP Code",
			Subject: "Your One-Time Password for Login",
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #0ea5e9; }
        .header h1 { color: #0ea5e9; margin: 0; font-size: 28px; }
        .content { margin-bottom: 30px; }
        .code-box { background-color: #f0f9ff; border: 2px dashed #0ea5e9; border-radius: 6px; padding: 20px; text-align: center; margin: 25px 0; }
        .code { font-size: 32px; font-weight: bold; color: #0ea5e9; letter-spacing: 5px; font-family: 'Courier New', monospace; }
        .expiry { color: #6b7280; font-size: 14px; margin-top: 10px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
        .warning { background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Login Verification</h1>
        </div>
        <div class="content">
            <p>Hello <strong>{{.username}}</strong>,</p>
            <p>Use the following one-time password to complete your login:</p>
            <div class="code-box">
                <div class="code">{{.code}}</div>
                <div class="expiry">This code expires in {{.expiry_minutes}} minutes</div>
            </div>
            <div class="warning">
                <strong>Security Notice:</strong> If you didn't attempt to log in, please secure your account immediately and contact support.
            </div>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>&copy; 2024 Auth Gateway. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: []string{"username", "email", "code", "expiry_minutes"},
		},
		{
			Type:    "otp_registration",
			Name:    "Registration OTP Code",
			Subject: "Complete Your Registration",
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #f59e0b; }
        .header h1 { color: #f59e0b; margin: 0; font-size: 28px; }
        .content { margin-bottom: 30px; }
        .code-box { background-color: #fffbeb; border: 2px dashed #f59e0b; border-radius: 6px; padding: 20px; text-align: center; margin: 25px 0; }
        .code { font-size: 32px; font-weight: bold; color: #f59e0b; letter-spacing: 5px; font-family: 'Courier New', monospace; }
        .expiry { color: #6b7280; font-size: 14px; margin-top: 10px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
        .info-box { background-color: #dbeafe; border-left: 4px solid #3b82f6; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Complete Registration</h1>
        </div>
        <div class="content">
            <p>Hello <strong>{{.username}}</strong>,</p>
            <p>Welcome! Please use the verification code below to complete your registration:</p>
            <div class="code-box">
                <div class="code">{{.code}}</div>
                <div class="expiry">This code expires in {{.expiry_minutes}} minutes</div>
            </div>
            <div class="info-box">
                <strong>Next Steps:</strong> After verifying your email, you'll have full access to all features and can start exploring our platform.
            </div>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>&copy; 2024 Auth Gateway. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: []string{"username", "email", "code", "expiry_minutes"},
		},
	}

	// Fetch all existing applications
	type AppRow struct {
		ID   string `bun:"id"`
		Name string `bun:"name"`
	}

	var applications []AppRow
	err = db.NewSelect().
		Table("applications").
		Column("id", "name").
		Scan(ctx, &applications)
	if err != nil {
		return fmt.Errorf("failed to fetch applications: %w", err)
	}

	// Insert templates for each application
	for _, app := range applications {
		for _, tmpl := range templateDefinitions {
			// Prepare variables as JSONB
			variablesJSON := "["
			for i, v := range tmpl.Variables {
				if i > 0 {
					variablesJSON += ","
				}
				variablesJSON += `"` + v + `"`
			}
			variablesJSON += "]"

			// Insert template for this application
			_, err = db.ExecContext(ctx, `
				INSERT INTO email_templates (application_id, type, name, subject, html_body, variables, is_active, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?::jsonb, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				ON CONFLICT DO NOTHING
			`, app.ID, tmpl.Type, tmpl.Name, tmpl.Subject, tmpl.HTMLBody, variablesJSON)
			if err != nil {
				return fmt.Errorf("failed to seed template %s for application %s: %w", tmpl.Type, app.Name, err)
			}
		}
	}

	fmt.Println(" OK")
	return nil
}

func removeApplicationEmailTemplates(ctx context.Context, db *bun.DB) error {
	// ============================================================
	// 1. Drop indexes
	// ============================================================
	indexes := []string{
		"DROP INDEX IF EXISTS idx_email_template_versions_app",
		"DROP INDEX IF EXISTS idx_email_templates_app_type",
		"DROP INDEX IF EXISTS idx_email_templates_application_id",
	}

	for _, indexSQL := range indexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to drop index: %w (SQL: %s)", err, indexSQL)
		}
	}

	// ============================================================
	// 2. Delete application-scoped templates
	// ============================================================
	_, err := db.ExecContext(ctx, `
		DELETE FROM email_templates WHERE application_id IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to delete application-scoped templates: %w", err)
	}

	// ============================================================
	// 3. Remove application_id columns
	// ============================================================
	_, err = db.ExecContext(ctx, `
		ALTER TABLE email_template_versions
		DROP COLUMN IF EXISTS application_id
	`)
	if err != nil {
		return fmt.Errorf("failed to drop application_id from email_template_versions: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		ALTER TABLE email_templates
		DROP COLUMN IF EXISTS application_id
	`)
	if err != nil {
		return fmt.Errorf("failed to drop application_id from email_templates: %w", err)
	}

	fmt.Println(" OK")
	return nil
}
