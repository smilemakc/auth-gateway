package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/smilemakc/auth-gateway/pkg/grpcclient"
)

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "localhost:50051", "gRPC server address")
	apiKey := flag.String("api-key", "", "API key for gRPC authentication (agw_*)")
	token := flag.String("token", "", "JWT token or API key (agw_*) for validation")
	userID := flag.String("user-id", "", "User ID for GetUser/CheckPermission")
	resource := flag.String("resource", "", "Resource name for CheckPermission")
	action := flag.String("action", "", "Action name for CheckPermission")
	testAll := flag.Bool("test-all", false, "Run all test examples")
	flag.Parse()

	fmt.Println("======================================")
	fmt.Println("  Auth Gateway gRPC Client Example")
	fmt.Println("======================================")
	fmt.Printf("Server: %s\n\n", *serverAddr)

	// Create client with options
	opts := []grpcclient.Option{grpcclient.WithTimeout(10 * time.Second)}
	if *apiKey != "" {
		opts = append(opts, grpcclient.WithAPIKey(*apiKey))
	}
	client, err := grpcclient.NewClient(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected to gRPC server successfully!")
	fmt.Println()

	// If test-all flag is set, run all examples
	if *testAll {
		runAllTests(client)
		return
	}

	// Run specific commands based on flags
	ctx := context.Background()

	if *token != "" {
		fmt.Println("--- ValidateToken ---")
		resp, err := client.ValidateToken(ctx, *token)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			printJSON("Response", resp)
		}
		fmt.Println()

		fmt.Println("--- IntrospectToken ---")
		introspect, err := client.IntrospectToken(ctx, *token)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			printJSON("Response", introspect)
		}
		fmt.Println()
	}

	if *userID != "" {
		fmt.Println("--- GetUser ---")
		resp, err := client.GetUser(ctx, *userID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			printJSON("Response", resp)
		}
		fmt.Println()

		if *resource != "" && *action != "" {
			fmt.Println("--- CheckPermission ---")
			fmt.Printf("Checking if user %s can %s on %s\n", *userID, *action, *resource)
			resp, err := client.CheckPermission(ctx, *userID, *resource, *action)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				printJSON("Response", resp)
			}
			fmt.Println()
		}
	}

	// If no flags provided, show usage
	if *token == "" && *userID == "" && !*testAll {
		fmt.Println("Usage Examples:")
		fmt.Println("---------------")
		fmt.Println()
		fmt.Println("1. Validate a JWT token:")
		fmt.Printf("   %s -token=<jwt_token>\n", os.Args[0])
		fmt.Println()
		fmt.Println("2. Validate an API key:")
		fmt.Printf("   %s -token=agw_<api_key>\n", os.Args[0])
		fmt.Println()
		fmt.Println("3. Get user by ID:")
		fmt.Printf("   %s -user-id=<uuid>\n", os.Args[0])
		fmt.Println()
		fmt.Println("4. Check permission:")
		fmt.Printf("   %s -user-id=<uuid> -resource=users -action=read\n", os.Args[0])
		fmt.Println()
		fmt.Println("5. Run all test examples (with mock data):")
		fmt.Printf("   %s -test-all\n", os.Args[0])
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -server     gRPC server address (default: localhost:50051)")
		fmt.Println("  -api-key    API key for authentication (agw_* format, required)")
		fmt.Println("  -token      JWT token or API key for validation")
		fmt.Println("  -user-id    User UUID for GetUser or CheckPermission")
		fmt.Println("  -resource   Resource name for CheckPermission (e.g., users, orders)")
		fmt.Println("  -action     Action name for CheckPermission (e.g., read, write, delete)")
		fmt.Println("  -test-all   Run all test examples with sample data")
	}
}

func runAllTests(client *grpcclient.Client) {
	ctx := context.Background()

	fmt.Println("==========================================")
	fmt.Println("  Running All gRPC Endpoint Tests")
	fmt.Println("==========================================")
	fmt.Println()

	// Test 1: ValidateToken with invalid token
	fmt.Println("Test 1: ValidateToken (invalid token)")
	fmt.Println("--------------------------------------")
	resp1, err := client.ValidateToken(ctx, "invalid-token")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp1)
	}
	fmt.Println()

	// Test 2: ValidateToken with empty token
	fmt.Println("Test 2: ValidateToken (empty token)")
	fmt.Println("------------------------------------")
	resp2, err := client.ValidateToken(ctx, "")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp2)
	}
	fmt.Println()

	// Test 3: ValidateToken with mock JWT format
	fmt.Println("Test 3: ValidateToken (malformed JWT)")
	fmt.Println("--------------------------------------")
	mockJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	resp3, err := client.ValidateToken(ctx, mockJWT)
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp3)
	}
	fmt.Println()

	// Test 4: ValidateToken with API key format
	fmt.Println("Test 4: ValidateToken (API key format)")
	fmt.Println("---------------------------------------")
	resp4, err := client.ValidateToken(ctx, "agw_test_api_key_12345")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp4)
	}
	fmt.Println()

	// Test 5: GetUser with invalid UUID
	fmt.Println("Test 5: GetUser (invalid UUID format)")
	fmt.Println("--------------------------------------")
	resp5, err := client.GetUser(ctx, "not-a-uuid")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp5)
	}
	fmt.Println()

	// Test 6: GetUser with valid UUID format (non-existent user)
	fmt.Println("Test 6: GetUser (non-existent user)")
	fmt.Println("------------------------------------")
	resp6, err := client.GetUser(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp6)
	}
	fmt.Println()

	// Test 7: GetUser with empty user ID
	fmt.Println("Test 7: GetUser (empty user ID)")
	fmt.Println("--------------------------------")
	resp7, err := client.GetUser(ctx, "")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp7)
	}
	fmt.Println()

	// Test 8: CheckPermission with valid UUID format
	fmt.Println("Test 8: CheckPermission (non-existent user)")
	fmt.Println("--------------------------------------------")
	resp8, err := client.CheckPermission(ctx, "00000000-0000-0000-0000-000000000000", "users", "read")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp8)
	}
	fmt.Println()

	// Test 9: CheckPermission with empty user ID
	fmt.Println("Test 9: CheckPermission (empty user ID)")
	fmt.Println("----------------------------------------")
	resp9, err := client.CheckPermission(ctx, "", "users", "read")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp9)
	}
	fmt.Println()

	// Test 10: IntrospectToken with invalid token
	fmt.Println("Test 10: IntrospectToken (invalid token)")
	fmt.Println("-----------------------------------------")
	resp10, err := client.IntrospectToken(ctx, "invalid-token")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp10)
	}
	fmt.Println()

	// Test 11: IntrospectToken with empty token
	fmt.Println("Test 11: IntrospectToken (empty token)")
	fmt.Println("---------------------------------------")
	resp11, err := client.IntrospectToken(ctx, "")
	if err != nil {
		fmt.Printf("gRPC Error: %v\n", err)
	} else {
		printJSON("Response", resp11)
	}
	fmt.Println()

	fmt.Println("==========================================")
	fmt.Println("  All Tests Completed!")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("Note: Most tests show error responses because")
	fmt.Println("they use invalid/non-existent test data.")
	fmt.Println("To test with real data:")
	fmt.Println("  1. Sign up/sign in via REST API to get a JWT token")
	fmt.Println("  2. Create an API key via /api-keys endpoint with required scopes")
	fmt.Println("  3. Use the -api-key flag with your API key")
	fmt.Println("  4. Use the -token flag with your real token")
}

func printJSON(label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("%s: %+v\n", label, v)
		return
	}
	fmt.Printf("%s:\n%s\n", label, string(data))
}
