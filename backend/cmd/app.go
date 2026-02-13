package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Application management commands",
	Long:  `Commands for managing applications (create, list).`,
}

var (
	appName         string
	appDisplayName  string
	appDescription  string
	appHomepageURL  string
	appCallbackURLs string
	appAuthMethods  string
)

var appCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new application",
	Long: `Create a new application in the Auth Gateway system.

The application will be created with a generated secret that is shown only once.
Make sure to save the secret immediately after creation.

Example:
  auth-gateway-cli app create --name my-app --display-name "My Application"
  auth-gateway-cli app create --name my-app --display-name "My App" --auth-methods "password,oauth_google"
  auth-gateway-cli app create --name my-app --display-name "My App" --callback-urls "https://example.com/callback"`,
	RunE: runAppCreate,
}

var (
	appListJSON bool
)

var appListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications",
	Long: `List all applications registered in the Auth Gateway system.

By default outputs a human-readable table. Use --json for machine-readable output.

Example:
  auth-gateway-cli app list
  auth-gateway-cli app list --json`,
	RunE: runAppList,
}

func init() {
	appCmd.AddCommand(appCreateCmd)
	appCmd.AddCommand(appListCmd)

	appCreateCmd.Flags().StringVar(&appName, "name", "", "Application slug name (required, e.g. my-app)")
	appCreateCmd.Flags().StringVar(&appDisplayName, "display-name", "", "Human-readable display name (required)")
	appCreateCmd.Flags().StringVar(&appDescription, "description", "", "Application description")
	appCreateCmd.Flags().StringVar(&appHomepageURL, "homepage-url", "", "Application homepage URL")
	appCreateCmd.Flags().StringVar(&appCallbackURLs, "callback-urls", "", "Comma-separated list of OAuth callback URLs")
	appCreateCmd.Flags().StringVar(&appAuthMethods, "auth-methods", "password", "Comma-separated auth methods (password,oauth_google,oauth_github,oauth_yandex,oauth_telegram,otp_email,otp_sms,totp,api_key)")

	appCreateCmd.MarkFlagRequired("name")
	appCreateCmd.MarkFlagRequired("display-name")

	appListCmd.Flags().BoolVar(&appListJSON, "json", false, "Output in JSON format")
}

func newApplicationService() *service.ApplicationService {
	log := logger.New("cli", logger.InfoLevel, false)
	appRepo := repository.NewApplicationRepository(db)
	appOAuthRepo := repository.NewAppOAuthProviderRepository(db)
	return service.NewApplicationService(appRepo, appOAuthRepo, log)
}

func runAppCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	req := &models.CreateApplicationRequest{
		Name:        appName,
		DisplayName: appDisplayName,
		Description: appDescription,
		HomepageURL: appHomepageURL,
	}

	if appCallbackURLs != "" {
		req.CallbackURLs = strings.Split(appCallbackURLs, ",")
	}

	if appAuthMethods != "" {
		req.AllowedAuthMethods = strings.Split(appAuthMethods, ",")
	}

	appSvc := newApplicationService()
	app, secret, err := appSvc.CreateApplication(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	fmt.Println("\nApplication created successfully!")
	fmt.Println("================================")
	fmt.Printf("ID:           %s\n", app.ID)
	fmt.Printf("Name:         %s\n", app.Name)
	fmt.Printf("Display Name: %s\n", app.DisplayName)
	if app.Description != "" {
		fmt.Printf("Description:  %s\n", app.Description)
	}
	fmt.Printf("Auth Methods: %s\n", strings.Join(app.AllowedAuthMethods, ", "))
	fmt.Printf("\nSecret:       %s\n", secret)
	fmt.Println("\n⚠  Save this secret now — it won't be shown again.")

	return nil
}

func runAppList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	appSvc := newApplicationService()
	resp, err := appSvc.ListApplications(ctx, 1, 100, nil)
	if err != nil {
		return fmt.Errorf("failed to list applications: %w", err)
	}

	if appListJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp.Applications)
	}

	if len(resp.Applications) == 0 {
		fmt.Println("No applications found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDISPLAY NAME\tACTIVE\tCREATED")
	for _, a := range resp.Applications {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%s\n",
			a.ID,
			a.Name,
			a.DisplayName,
			a.IsActive,
			a.CreatedAt.Format("2006-01-02"),
		)
	}
	return w.Flush()
}
