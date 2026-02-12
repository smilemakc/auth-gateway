// Example: Admin operations with the Auth Gateway Go SDK
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

func main() {
	// Create client and authenticate as admin
	client := authgateway.NewClient(authgateway.Config{
		BaseURL: "http://localhost:8811",
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	// Sign in as admin
	_, err := client.Auth.SignInWithEmail(ctx, "admin@example.com", "adminpassword")
	if err != nil {
		log.Fatalf("Failed to sign in as admin: %v", err)
	}
	fmt.Println("Signed in as admin")

	// Example 1: Get system stats
	fmt.Println("\n=== System Statistics ===")
	stats, err := client.Admin.GetStats(ctx)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
	} else {
		fmt.Printf("Total Users: %d\n", stats.TotalUsers)
		fmt.Printf("Active Users: %d\n", stats.ActiveUsers)
		fmt.Printf("Total Sessions: %d\n", stats.TotalSessions)
		fmt.Printf("Total API Keys: %d\n", stats.TotalAPIKeys)
		fmt.Printf("2FA Enabled Users: %d\n", stats.TwoFAEnabledUsers)
	}

	// Example 2: List users with pagination
	fmt.Println("\n=== List Users ===")
	usersResp, err := client.Admin.ListUsers(ctx, &models.ListUsersParams{
		Page:   1,
		Limit:  10,
		Search: "",
	})
	if err != nil {
		log.Printf("Failed to list users: %v", err)
	} else {
		fmt.Printf("Total users: %d (page %d of %d)\n",
			usersResp.Pagination.Total,
			usersResp.Pagination.Page,
			usersResp.Pagination.TotalPages)
		for _, user := range usersResp.Items {
			status := "active"
			if !user.IsActive {
				status = "inactive"
			}
			fmt.Printf("- %s: %s (%s)\n", user.ID[:8], user.Email, status)
		}
	}

	// Example 3: Create a new user
	fmt.Println("\n=== Create User ===")
	newUser, err := client.Admin.CreateUser(ctx, &models.CreateUserRequest{
		Email:    "newuser@example.com",
		Username: "newuser123",
		Password: "securepassword",
		FullName: "New User",
	})
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	} else {
		fmt.Printf("Created user: %s (%s)\n", newUser.Username, newUser.Email)
	}

	// Example 4: List roles
	fmt.Println("\n=== List Roles ===")
	roles, err := client.Admin.ListRoles(ctx)
	if err != nil {
		log.Printf("Failed to list roles: %v", err)
	} else {
		for _, role := range roles {
			system := ""
			if role.IsSystemRole {
				system = " [system]"
			}
			fmt.Printf("- %s: %s%s\n", role.Name, role.DisplayName, system)
		}
	}

	// Example 5: Create a new role
	fmt.Println("\n=== Create Role ===")
	newRole, err := client.Admin.CreateRole(ctx, &models.CreateRoleRequest{
		Name:        "custom_role",
		DisplayName: "Custom Role",
		Description: "A custom role for demonstration",
	})
	if err != nil {
		log.Printf("Failed to create role: %v", err)
	} else {
		fmt.Printf("Created role: %s\n", newRole.DisplayName)
	}

	// Example 6: List audit logs
	fmt.Println("\n=== Recent Audit Logs ===")
	logsResp, err := client.Admin.ListAuditLogs(ctx, &models.ListAuditLogsParams{
		Page:  1,
		Limit: 5,
	})
	if err != nil {
		log.Printf("Failed to list audit logs: %v", err)
	} else {
		for _, logEntry := range logsResp.Items {
			fmt.Printf("- [%s] %s: %s (%s)\n",
				logEntry.Timestamp.Format("2006-01-02 15:04"),
				logEntry.Action,
				logEntry.Resource,
				logEntry.Status)
		}
	}

	// Example 7: Get session statistics
	fmt.Println("\n=== Session Statistics ===")
	sessionStats, err := client.Admin.GetSessionStats(ctx)
	if err != nil {
		log.Printf("Failed to get session stats: %v", err)
	} else {
		fmt.Printf("Active Sessions: %d\n", sessionStats.TotalActiveSessions)
		fmt.Println("By Device:")
		for device, count := range sessionStats.SessionsByDevice {
			fmt.Printf("  - %s: %d\n", device, count)
		}
	}

	// Example 8: IP Filters
	fmt.Println("\n=== IP Filters ===")
	filters, err := client.Admin.ListIPFilters(ctx)
	if err != nil {
		log.Printf("Failed to list IP filters: %v", err)
	} else {
		if len(filters) == 0 {
			fmt.Println("No IP filters configured")
		} else {
			for _, filter := range filters {
				fmt.Printf("- %s: %s (%s)\n", filter.Type, filter.IPAddress, filter.Description)
			}
		}
	}

	// Example 9: Create IP filter
	fmt.Println("\n=== Create IP Filter ===")
	newFilter, err := client.Admin.CreateIPFilter(ctx, &models.CreateIPFilterRequest{
		IPAddress:   "192.168.1.100",
		Type:        "whitelist",
		Description: "Office IP",
	})
	if err != nil {
		log.Printf("Failed to create IP filter: %v", err)
	} else {
		fmt.Printf("Created IP filter: %s (%s)\n", newFilter.IPAddress, newFilter.Type)
	}

	// Example 10: Get geo distribution
	fmt.Println("\n=== Geographic Distribution ===")
	geoDist, err := client.Admin.GetGeoDistribution(ctx)
	if err != nil {
		log.Printf("Failed to get geo distribution: %v", err)
	} else {
		for _, geo := range geoDist {
			fmt.Printf("- %s: %d users\n", geo.Country, geo.Count)
		}
	}
}
