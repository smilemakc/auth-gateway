#!/bin/bash

# Script to generate code from protobuf definitions
# Generates for backend, go-sdk, and client-sdk (TypeScript) packages

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$BACKEND_DIR")"
GO_SDK_DIR="$ROOT_DIR/packages/go-sdk"
CLIENT_SDK_DIR="$ROOT_DIR/packages/client-sdk"

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

# Generate for client-sdk (TypeScript) if directory exists
if [ -d "$CLIENT_SDK_DIR" ]; then
    echo ""
    echo "3. Generating for client-sdk (packages/client-sdk/src/grpc/generated/)..."

    TS_PROTO_PLUGIN="$CLIENT_SDK_DIR/node_modules/.bin/protoc-gen-ts_proto"

    if [ ! -f "$TS_PROTO_PLUGIN" ]; then
        echo "   Warning: ts-proto not found at $TS_PROTO_PLUGIN"
        echo "   Run 'npm install' in packages/client-sdk first"
        echo "   Skipping TypeScript generation"
    else
        # Create generated directory
        mkdir -p "$CLIENT_SDK_DIR/src/grpc/generated"

        # Generate TypeScript interfaces from proto
        protoc --plugin=protoc-gen-ts_proto="$TS_PROTO_PLUGIN" \
               --ts_proto_out="$CLIENT_SDK_DIR/src/grpc/generated" \
               --ts_proto_opt=onlyTypes=true \
               --ts_proto_opt=snakeToCamel=true \
               --ts_proto_opt=useOptionals=messages \
               --ts_proto_opt=stringEnums=true \
               -I "$BACKEND_DIR/proto" \
               "$BACKEND_DIR/proto/auth.proto"

        echo "   ✓ packages/client-sdk/src/grpc/generated/auth.ts"

        # Copy auth.proto for runtime loading by @grpc/proto-loader
        cp "$BACKEND_DIR/proto/auth.proto" "$CLIENT_SDK_DIR/src/grpc/auth.proto"
        echo "   ✓ packages/client-sdk/src/grpc/auth.proto (copied)"
    fi
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
if [ -d "$CLIENT_SDK_DIR" ] && [ -f "$CLIENT_SDK_DIR/src/grpc/generated/auth.ts" ]; then
    echo "  - packages/client-sdk/     (for TypeScript SDK)"
fi