#!/bin/bash

# Script to generate Go code from protobuf definitions
# Generates for both backend and go-sdk packages

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$BACKEND_DIR")"
GO_SDK_DIR="$ROOT_DIR/packages/go-sdk"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed"
    echo "Install it from: https://protobuf.dev/installation/"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

echo "=== Generating Proto Files ==="
echo ""

# Generate for backend
echo "1. Generating for backend (backend/proto/)..."
cd "$BACKEND_DIR"
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/auth.proto
echo "   ✓ backend/proto/auth.pb.go"
echo "   ✓ backend/proto/auth_grpc.pb.go"

# Generate for go-sdk (if directory exists)
if [ -d "$GO_SDK_DIR" ]; then
    echo ""
    echo "2. Generating for go-sdk (packages/go-sdk/proto/)..."

    # Create proto directory in go-sdk if it doesn't exist
    mkdir -p "$GO_SDK_DIR/proto"

    # Copy the proto file to go-sdk and modify go_package for SDK
    # We need different go_package for the SDK module
    sed 's|option go_package = "github.com/smilemakc/auth-gateway/proto";|option go_package = "github.com/smilemakc/auth-gateway/packages/go-sdk/proto";|' \
        "$BACKEND_DIR/proto/auth.proto" > "$GO_SDK_DIR/proto/auth.proto"

    cd "$GO_SDK_DIR"
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           proto/auth.proto

    # Remove the temporary proto file from go-sdk (keep only generated .go files)
    rm -f "$GO_SDK_DIR/proto/auth.proto"

    echo "   ✓ packages/go-sdk/proto/auth.pb.go"
    echo "   ✓ packages/go-sdk/proto/auth_grpc.pb.go"
fi

echo ""
echo "=== Proto Generation Completed Successfully! ==="
echo ""
echo "Source: backend/proto/auth.proto"
echo ""
echo "Generated locations:"
echo "  - backend/proto/           (for backend server)"
if [ -d "$GO_SDK_DIR" ]; then
    echo "  - packages/go-sdk/proto/   (for Go SDK)"
fi