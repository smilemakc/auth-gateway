package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Admin role ID is a well-known UUID from the database migrations
var adminRoleID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin management commands",
	Long:  `Commands for managing administrator accounts and permissions.`,
}

var (
	adminEmail    string
	adminUsername string
	adminPassword string
	adminFullName string
)

var adminCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an admin user",
	Long: `Create a new administrator user with full access to all entities.

The admin user will have:
- The "admin" role with all permissions
- Email marked as verified
- Account marked as active

If the admin role does not exist, it will be created automatically along
with all system permissions.

Example:
  auth-gateway-cli admin create --email admin@example.com --username admin
  auth-gateway-cli admin create --email admin@example.com --username admin --password mysecurepass
  auth-gateway-cli admin create --email admin@example.com --username admin --full-name "System Admin"`,
	RunE: runAdminCreate,
}

func init() {
	adminCmd.AddCommand(adminCreateCmd)

	adminCreateCmd.Flags().StringVarP(&adminEmail, "email", "e", "", "Admin email address (required)")
	adminCreateCmd.Flags().StringVarP(&adminUsername, "username", "u", "", "Admin username (required)")
	adminCreateCmd.Flags().StringVarP(&adminPassword, "password", "p", "", "Admin password (will prompt if not provided)")
	adminCreateCmd.Flags().StringVarP(&adminFullName, "full-name", "n", "", "Admin full name (optional)")

	adminCreateCmd.MarkFlagRequired("email")
	adminCreateCmd.MarkFlagRequired("username")
}

// systemPermissions defines all system permissions that should exist
var systemPermissions = []struct {
	Name        string
	Resource    string
	Action      string
	Description string
}{
	// User management permissions
	{"users.create", "users", "create", "Create new users"},
	{"users.read", "users", "read", "View user information"},
	{"users.update", "users", "update", "Update user information"},
	{"users.delete", "users", "delete", "Delete users"},
	{"users.list", "users", "list", "List all users"},

	// Role management permissions
	{"roles.create", "roles", "create", "Create new roles"},
	{"roles.read", "roles", "read", "View role information"},
	{"roles.update", "roles", "update", "Update roles"},
	{"roles.delete", "roles", "delete", "Delete roles"},
	{"roles.list", "roles", "list", "List all roles"},

	// Permission management
	{"permissions.create", "permissions", "create", "Create permissions"},
	{"permissions.read", "permissions", "read", "View permissions"},
	{"permissions.update", "permissions", "update", "Update permissions"},
	{"permissions.delete", "permissions", "delete", "Delete permissions"},
	{"permissions.list", "permissions", "list", "List all permissions"},

	// API Key permissions
	{"api_keys.create", "api_keys", "create", "Create API keys"},
	{"api_keys.read", "api_keys", "read", "View API keys"},
	{"api_keys.update", "api_keys", "update", "Update API keys"},
	{"api_keys.delete", "api_keys", "delete", "Delete API keys"},
	{"api_keys.revoke", "api_keys", "revoke", "Revoke API keys"},
	{"api_keys.list", "api_keys", "list", "List all API keys"},

	// Session management permissions
	{"sessions.read", "sessions", "read", "View active sessions"},
	{"sessions.revoke", "sessions", "revoke", "Revoke user sessions"},
	{"sessions.list", "sessions", "list", "List all sessions"},

	// Audit log permissions
	{"audit_logs.read", "audit_logs", "read", "View audit logs"},
	{"audit_logs.list", "audit_logs", "list", "List audit logs"},

	// IP filter permissions
	{"ip_filters.create", "ip_filters", "create", "Create IP filters"},
	{"ip_filters.read", "ip_filters", "read", "View IP filters"},
	{"ip_filters.update", "ip_filters", "update", "Update IP filters"},
	{"ip_filters.delete", "ip_filters", "delete", "Delete IP filters"},
	{"ip_filters.list", "ip_filters", "list", "List IP filters"},

	// Webhook permissions
	{"webhooks.create", "webhooks", "create", "Create webhooks"},
	{"webhooks.read", "webhooks", "read", "View webhooks"},
	{"webhooks.update", "webhooks", "update", "Update webhooks"},
	{"webhooks.delete", "webhooks", "delete", "Delete webhooks"},
	{"webhooks.list", "webhooks", "list", "List webhooks"},
	{"webhooks.test", "webhooks", "test", "Test webhook delivery"},

	// Email template permissions
	{"email_templates.create", "email_templates", "create", "Create email templates"},
	{"email_templates.read", "email_templates", "read", "View email templates"},
	{"email_templates.update", "email_templates", "update", "Update email templates"},
	{"email_templates.delete", "email_templates", "delete", "Delete email templates"},
	{"email_templates.list", "email_templates", "list", "List email templates"},

	// Branding permissions
	{"branding.read", "branding", "read", "View branding settings"},
	{"branding.update", "branding", "update", "Update branding settings"},

	// System settings permissions
	{"system.read", "system", "read", "View system settings"},
	{"system.update", "system", "update", "Update system settings"},
	{"system.health", "system", "health", "View system health metrics"},
	{"system.maintenance", "system", "maintenance", "Control maintenance mode"},

	// Statistics permissions
	{"stats.view", "stats", "view", "View system statistics"},
	{"stats.export", "stats", "export", "Export statistics data"},
}

func runAdminCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate email
	email := utils.NormalizeEmail(adminEmail)
	if !utils.IsValidEmail(email) {
		return fmt.Errorf("invalid email format: %s", adminEmail)
	}

	// Validate username
	username := utils.NormalizeUsername(adminUsername)
	if !utils.IsValidUsername(username) {
		return fmt.Errorf("invalid username format: %s (must be 3-100 characters, alphanumeric with underscores/hyphens)", adminUsername)
	}

	// Get password - prompt if not provided
	password := adminPassword
	if password == "" {
		var err error
		password, err = promptPassword()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
	}

	// Validate password
	if !utils.IsPasswordValid(password) {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	rbacRepo := repository.NewRBACRepository(db)

	// Check if email already exists
	exists, err := userRepo.EmailExists(email)
	if err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return fmt.Errorf("email already exists: %s", email)
	}

	// Check if username already exists
	exists, err = userRepo.UsernameExists(username)
	if err != nil {
		return fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return fmt.Errorf("username already exists: %s", username)
	}

	// Ensure admin role exists (create if it doesn't)
	adminRole, err := ensureAdminRoleExists(ctx, rbacRepo)
	if err != nil {
		return fmt.Errorf("failed to setup admin role: %w", err)
	}

	// Hash password
	passwordHash, err := utils.HashPassword(password, cfg.Security.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	user := &models.User{
		ID:            uuid.New(),
		Email:         email,
		Username:      username,
		PasswordHash:  passwordHash,
		FullName:      adminFullName,
		Role:          string(models.RoleAdmin),
		RoleID:        &adminRoleID,
		AccountType:   string(models.AccountTypeHuman),
		EmailVerified: true,
		IsActive:      true,
	}

	if err := userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	fmt.Println("\nAdmin user created successfully!")
	fmt.Println("================================")
	fmt.Printf("ID:       %s\n", user.ID)
	fmt.Printf("Email:    %s\n", user.Email)
	fmt.Printf("Username: %s\n", user.Username)
	fmt.Printf("Role:     %s (%s)\n", adminRole.DisplayName, adminRole.Name)
	fmt.Printf("Permissions: %d\n", len(adminRole.Permissions))
	fmt.Println("\nThe admin user can now sign in and manage all entities in the system.")

	return nil
}

// ensureAdminRoleExists checks if the admin role exists, creates it if not
func ensureAdminRoleExists(ctx context.Context, rbacRepo *repository.RBACRepository) (*models.Role, error) {
	// Try to get existing admin role
	adminRole, err := rbacRepo.GetRoleByID(ctx, adminRoleID)
	if err == nil {
		// Admin role exists, return it
		return adminRole, nil
	}

	// Admin role doesn't exist, create it along with permissions
	fmt.Println("Admin role not found. Creating admin role and permissions...")

	// First, ensure all permissions exist
	permissionIDs := make([]uuid.UUID, 0, len(systemPermissions))
	for _, perm := range systemPermissions {
		permission, err := ensurePermissionExists(ctx, rbacRepo, perm.Name, perm.Resource, perm.Action, perm.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to create permission %s: %w", perm.Name, err)
		}
		permissionIDs = append(permissionIDs, permission.ID)
	}
	fmt.Printf("  Created/verified %d permissions\n", len(permissionIDs))

	// Create the admin role
	adminRole = &models.Role{
		ID:           adminRoleID,
		Name:         "admin",
		DisplayName:  "Administrator",
		Description:  "Full system access with all permissions",
		IsSystemRole: true,
	}

	if err := rbacRepo.CreateRole(ctx, adminRole); err != nil {
		return nil, fmt.Errorf("failed to create admin role: %w", err)
	}
	fmt.Println("  Created admin role")

	// Assign all permissions to admin role
	if err := rbacRepo.SetRolePermissions(ctx, adminRoleID, permissionIDs); err != nil {
		return nil, fmt.Errorf("failed to assign permissions to admin role: %w", err)
	}
	fmt.Printf("  Assigned %d permissions to admin role\n", len(permissionIDs))

	// Reload the role with permissions
	adminRole, err = rbacRepo.GetRoleByID(ctx, adminRoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload admin role: %w", err)
	}

	return adminRole, nil
}

// ensurePermissionExists checks if a permission exists, creates it if not
func ensurePermissionExists(ctx context.Context, rbacRepo *repository.RBACRepository, name, resource, action, description string) (*models.Permission, error) {
	// Try to get existing permission
	permission, err := rbacRepo.GetPermissionByName(ctx, name)
	if err == nil {
		return permission, nil
	}

	// Permission doesn't exist, create it
	permission = &models.Permission{
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
	}

	if err := rbacRepo.CreatePermission(ctx, permission); err != nil {
		return nil, err
	}

	return permission, nil
}

// promptPassword securely prompts for password input
func promptPassword() (string, error) {
	fmt.Print("Enter admin password: ")

	// Try to read password without echo
	if term.IsTerminal(int(syscall.Stdin)) {
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // Add newline after password input
		if err != nil {
			return "", err
		}

		password := string(passwordBytes)

		// Confirm password
		fmt.Print("Confirm admin password: ")
		confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return "", err
		}

		if password != string(confirmBytes) {
			return "", fmt.Errorf("passwords do not match")
		}

		return password, nil
	}

	// Fallback for non-terminal input (piped input)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(password), nil
}
