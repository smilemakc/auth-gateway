package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Adding notification email templates (password_changed, login_alert, 2fa_enabled, 2fa_disabled)...")
		return addNotificationTemplates(ctx, db)
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Removing notification email templates...")
		return removeNotificationTemplates(ctx, db)
	})
}

func addNotificationTemplates(ctx context.Context, db *bun.DB) error {
	// Define the 4 new notification template types
	templateDefinitions := []struct {
		Type      string
		Name      string
		Subject   string
		HTMLBody  string
		Variables string // JSON array
	}{
		{
			Type:    "password_changed",
			Name:    "Password Changed",
			Subject: "Your Password Has Been Changed",
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
        .info-box { background-color: #f8f9fa; border-radius: 6px; padding: 20px; margin: 20px 0; }
        .info-box p { margin: 5px 0; }
        .warning { background-color: #fef2f2; border-left: 4px solid #ef4444; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Changed</h1>
        </div>
        <div class="content">
            <p>Hello <strong>{{.username}}</strong>,</p>
            <p>Your password was successfully changed.</p>
            <div class="info-box">
                <p><strong>IP Address:</strong> {{.ip_address}}</p>
                <p><strong>Time:</strong> {{.timestamp}}</p>
            </div>
            <div class="warning">
                <strong>Security Alert:</strong> If you did not make this change, please contact support immediately and change your password.
            </div>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: `["username", "email", "ip_address", "user_agent", "timestamp"]`,
		},
		{
			Type:    "login_alert",
			Name:    "Login Alert",
			Subject: "New Login to Your Account",
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #3b82f6; }
        .header h1 { color: #3b82f6; margin: 0; font-size: 28px; }
        .content { margin-bottom: 30px; }
        .info-box { background-color: #f8f9fa; border-radius: 6px; padding: 20px; margin: 20px 0; }
        .info-box p { margin: 5px 0; }
        .warning { background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>New Login Detected</h1>
        </div>
        <div class="content">
            <p>Hello <strong>{{.username}}</strong>,</p>
            <p>A new login to your account was detected.</p>
            <div class="info-box">
                <p><strong>IP Address:</strong> {{.ip_address}}</p>
                <p><strong>Device:</strong> {{.device_type}}</p>
                <p><strong>Time:</strong> {{.timestamp}}</p>
            </div>
            <div class="warning">
                <strong>Not you?</strong> If you didn't log in, please change your password immediately and enable two-factor authentication.
            </div>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: `["username", "email", "ip_address", "user_agent", "device_type", "location", "timestamp"]`,
		},
		{
			Type:    "2fa_enabled",
			Name:    "2FA Enabled",
			Subject: "Two-Factor Authentication Enabled",
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
        .success-box { background: linear-gradient(135deg, #10b981 0%, #059669 100%); border-radius: 8px; padding: 25px; text-align: center; color: #ffffff; margin: 25px 0; }
        .info-box { background-color: #dbeafe; border-left: 4px solid #3b82f6; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>2FA Enabled</h1>
        </div>
        <div class="content">
            <p>Hello <strong>{{.username}}</strong>,</p>
            <div class="success-box">
                <h2>Your account is now more secure!</h2>
                <p>Two-factor authentication has been enabled.</p>
            </div>
            <p>From now on, you will need your authenticator app to sign in to your account.</p>
            <div class="info-box">
                <strong>Reminder:</strong> Make sure to save your backup codes in a secure location. You'll need them if you lose access to your authenticator app.
            </div>
            <p><strong>Time:</strong> {{.timestamp}}</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: `["username", "email", "timestamp"]`,
		},
		{
			Type:    "2fa_disabled",
			Name:    "2FA Disabled",
			Subject: "Two-Factor Authentication Disabled",
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
        .warning { background-color: #fef2f2; border-left: 4px solid #ef4444; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .action-box { background-color: #fef3c7; border-radius: 6px; padding: 20px; margin: 20px 0; text-align: center; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>2FA Disabled</h1>
        </div>
        <div class="content">
            <p>Hello <strong>{{.username}}</strong>,</p>
            <p>Two-factor authentication has been disabled on your account.</p>
            <div class="warning">
                <strong>Security Warning:</strong> Your account is now less secure without two-factor authentication. We strongly recommend re-enabling it.
            </div>
            <div class="action-box">
                <p>To re-enable 2FA, go to your account security settings.</p>
            </div>
            <p><strong>Time:</strong> {{.timestamp}}</p>
            <p>If you did not disable 2FA, please contact support immediately.</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`,
			Variables: `["username", "email", "timestamp"]`,
		},
	}

	// ============================================================
	// 1. Insert global templates for the 4 new types
	// ============================================================
	for _, tmpl := range templateDefinitions {
		_, err := db.ExecContext(ctx, `
			INSERT INTO email_templates (type, name, subject, html_body, variables, is_active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?::jsonb, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT DO NOTHING
		`, tmpl.Type, tmpl.Name, tmpl.Subject, tmpl.HTMLBody, tmpl.Variables)
		if err != nil {
			return fmt.Errorf("failed to seed global template %s: %w", tmpl.Type, err)
		}
	}

	// ============================================================
	// 2. Insert app-scoped templates for each existing application
	// ============================================================
	type AppRow struct {
		ID   string `bun:"id"`
		Name string `bun:"name"`
	}

	var applications []AppRow
	err := db.NewSelect().
		Table("applications").
		Column("id", "name").
		Scan(ctx, &applications)
	if err != nil {
		return fmt.Errorf("failed to fetch applications: %w", err)
	}

	for _, app := range applications {
		for _, tmpl := range templateDefinitions {
			_, err = db.ExecContext(ctx, `
				INSERT INTO email_templates (application_id, type, name, subject, html_body, variables, is_active, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?::jsonb, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				ON CONFLICT DO NOTHING
			`, app.ID, tmpl.Type, tmpl.Name, tmpl.Subject, tmpl.HTMLBody, tmpl.Variables)
			if err != nil {
				return fmt.Errorf("failed to seed template %s for application %s: %w", tmpl.Type, app.Name, err)
			}
		}
	}

	fmt.Println(" OK")
	return nil
}

func removeNotificationTemplates(ctx context.Context, db *bun.DB) error {
	notificationTypes := []string{"password_changed", "login_alert", "2fa_enabled", "2fa_disabled"}

	for _, templateType := range notificationTypes {
		_, err := db.ExecContext(ctx, `DELETE FROM email_templates WHERE type = ?`, templateType)
		if err != nil {
			return fmt.Errorf("failed to delete %s templates: %w", templateType, err)
		}
	}

	fmt.Println(" OK")
	return nil
}
