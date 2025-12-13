#!/bin/bash

# Auth Gateway gRPC Endpoint Test Script
# This script tests all gRPC endpoints using grpcurl

set -e

GRPC_SERVER="${GRPC_SERVER:-localhost:50051}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "  Auth Gateway gRPC Endpoint Tests"
echo "=========================================="
echo "Server: $GRPC_SERVER"
echo ""

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${RED}Error: grpcurl is not installed.${NC}"
    echo "Install it with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    exit 1
fi

# Function to run a test
run_test() {
    local name="$1"
    local method="$2"
    local data="$3"

    echo -e "${BLUE}Test: $name${NC}"
    echo "Method: $method"
    echo "Request: $data"
    echo "---"

    if grpcurl -plaintext -d "$data" "$GRPC_SERVER" "$method" 2>&1; then
        echo -e "${GREEN}OK${NC}"
    else
        echo -e "${YELLOW}(Expected error for test data)${NC}"
    fi
    echo ""
}

echo "Listing available services..."
echo "---"
grpcurl -plaintext "$GRPC_SERVER" list || true
echo ""

echo "Describing AuthService..."
echo "---"
grpcurl -plaintext "$GRPC_SERVER" describe auth.AuthService || true
echo ""

echo "=========================================="
echo "  Running Endpoint Tests"
echo "=========================================="
echo ""

# Test 1: ValidateToken - empty token
run_test "ValidateToken (empty token)" \
    "auth.AuthService/ValidateToken" \
    '{"access_token": ""}'

# Test 2: ValidateToken - invalid token
run_test "ValidateToken (invalid token)" \
    "auth.AuthService/ValidateToken" \
    '{"access_token": "invalid-token-12345"}'

# Test 3: ValidateToken - valid JWT format (will fail signature check)
run_test "ValidateToken (JWT format - invalid signature)" \
    "auth.AuthService/ValidateToken" \
    '{"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"}'

# Test 4: ValidateToken - API key format (will fail validation)
run_test "ValidateToken (API key format)" \
    "auth.AuthService/ValidateToken" \
    '{"access_token": "agw_test_api_key_12345"}'

# Test 5: GetUser - empty user ID
run_test "GetUser (empty user ID)" \
    "auth.AuthService/GetUser" \
    '{"user_id": ""}'

# Test 6: GetUser - invalid UUID format
run_test "GetUser (invalid UUID)" \
    "auth.AuthService/GetUser" \
    '{"user_id": "not-a-valid-uuid"}'

# Test 7: GetUser - valid UUID format (non-existent user)
run_test "GetUser (non-existent user)" \
    "auth.AuthService/GetUser" \
    '{"user_id": "00000000-0000-0000-0000-000000000000"}'

# Test 8: CheckPermission - empty user ID
run_test "CheckPermission (empty user ID)" \
    "auth.AuthService/CheckPermission" \
    '{"user_id": "", "resource": "users", "action": "read"}'

# Test 9: CheckPermission - invalid UUID
run_test "CheckPermission (invalid UUID)" \
    "auth.AuthService/CheckPermission" \
    '{"user_id": "invalid-uuid", "resource": "users", "action": "read"}'

# Test 10: CheckPermission - valid UUID (non-existent user)
run_test "CheckPermission (non-existent user)" \
    "auth.AuthService/CheckPermission" \
    '{"user_id": "00000000-0000-0000-0000-000000000000", "resource": "users", "action": "read"}'

# Test 11: IntrospectToken - empty token
run_test "IntrospectToken (empty token)" \
    "auth.AuthService/IntrospectToken" \
    '{"access_token": ""}'

# Test 12: IntrospectToken - invalid token
run_test "IntrospectToken (invalid token)" \
    "auth.AuthService/IntrospectToken" \
    '{"access_token": "invalid-token"}'

echo "=========================================="
echo "  All Tests Completed"
echo "=========================================="
echo ""
echo "Notes:"
echo "- Most tests show error responses because they use invalid test data"
echo "- To test with real data:"
echo "  1. Sign in via REST API: POST /auth/signin"
echo "  2. Use the returned access_token with ValidateToken"
echo "  3. Or create an API key: POST /api-keys"
echo ""
echo "Example with real token:"
echo "  grpcurl -plaintext -d '{\"access_token\": \"YOUR_JWT_TOKEN\"}' \\"
echo "    $GRPC_SERVER auth.AuthService/ValidateToken"
