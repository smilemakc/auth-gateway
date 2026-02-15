package cmd

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

const adminRoleName = "admin"

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

func runAdminCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate email
	email := utils.NormalizeEmail(adminEmail)
	if err := utils.ValidateEmail(email); err != nil {
		return fmt.Errorf("invalid email: %w", err)
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
	exists, err := userRepo.EmailExists(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return fmt.Errorf("email already exists: %s", email)
	}

	// Check if username already exists
	exists, err = userRepo.UsernameExists(ctx, username)
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
		AccountType:   string(models.AccountTypeHuman),
		EmailVerified: true,
		IsActive:      true,
	}

	if err := userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Assign admin role to user
	if err := rbacRepo.AssignRoleToUser(ctx, user.ID, adminRole.ID, user.ID); err != nil {
		return fmt.Errorf("failed to assign admin role: %w", err)
	}

	fmt.Println("\nAdmin user created successfully!")
	fmt.Println("================================")
	fmt.Printf("ID:       %s\n", user.ID)
	fmt.Printf("Email:    %s\n", user.Email)
	fmt.Printf("Username: %s\n", user.Username)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Role:     %s (%s)\n", adminRole.DisplayName, adminRole.Name)
	fmt.Printf("Permissions: %d\n", len(adminRole.Permissions))
	fmt.Println("\nThe admin user can now sign in and manage all entities in the system.")

	return nil
}

// ensureAdminRoleExists checks if the admin role exists, creates it if not
func ensureAdminRoleExists(ctx context.Context, rbacRepo *repository.RBACRepository) (*models.Role, error) {
	// Reload the role with permissions
	adminRole, err := rbacRepo.GetRoleByName(ctx, adminRoleName)
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
