#!/bin/bash

# Generate OIDC signing keys for Auth Gateway
# Usage: ./generate-keys.sh [output_dir] [key_id]

set -e

OUTPUT_DIR="${1:-./keys}"
KEY_ID="${2:-key-$(date +%Y%m%d)}"

echo "======================================"
echo "OIDC Key Generation for Auth Gateway"
echo "======================================"
echo ""
echo "Output directory: $OUTPUT_DIR"
echo "Key ID: $KEY_ID"
echo ""

mkdir -p "$OUTPUT_DIR"

echo "Generating RSA 2048-bit key pair..."
openssl genrsa -out "$OUTPUT_DIR/rsa_private_$KEY_ID.pem" 2048 2>/dev/null
if [ $? -eq 0 ]; then
    openssl rsa -in "$OUTPUT_DIR/rsa_private_$KEY_ID.pem" -pubout -out "$OUTPUT_DIR/rsa_public_$KEY_ID.pem" 2>/dev/null
    echo "✓ RSA key pair generated"
else
    echo "✗ Failed to generate RSA key pair"
    exit 1
fi

echo ""
echo "Generating ECDSA P-256 key pair..."
openssl ecparam -name prime256v1 -genkey -noout -out "$OUTPUT_DIR/ecdsa_private_$KEY_ID.pem" 2>/dev/null
if [ $? -eq 0 ]; then
    openssl ec -in "$OUTPUT_DIR/ecdsa_private_$KEY_ID.pem" -pubout -out "$OUTPUT_DIR/ecdsa_public_$KEY_ID.pem" 2>/dev/null
    echo "✓ ECDSA key pair generated"
else
    echo "✗ Failed to generate ECDSA key pair"
    exit 1
fi

echo ""
echo "======================================"
echo "Keys generated successfully!"
echo "======================================"
echo ""
echo "Generated files:"
echo "  - $OUTPUT_DIR/rsa_private_$KEY_ID.pem"
echo "  - $OUTPUT_DIR/rsa_public_$KEY_ID.pem"
echo "  - $OUTPUT_DIR/ecdsa_private_$KEY_ID.pem"
echo "  - $OUTPUT_DIR/ecdsa_public_$KEY_ID.pem"
echo ""
echo "======================================"
echo "Configuration"
echo "======================================"
echo ""
echo "Add to your .env file:"
echo ""
echo "# OIDC Signing Keys (RSA recommended)"
echo "OIDC_SIGNING_KEY_PATH=$OUTPUT_DIR/rsa_private_$KEY_ID.pem"
echo "OIDC_SIGNING_KEY_ID=$KEY_ID"
echo "OIDC_SIGNING_ALGORITHM=RS256"
echo ""
echo "# OR use ECDSA (alternative)"
echo "# OIDC_SIGNING_KEY_PATH=$OUTPUT_DIR/ecdsa_private_$KEY_ID.pem"
echo "# OIDC_SIGNING_KEY_ID=$KEY_ID"
echo "# OIDC_SIGNING_ALGORITHM=ES256"
echo ""
echo "======================================"
echo ""
echo "IMPORTANT: Keep private keys secure!"
echo "Add $OUTPUT_DIR/*.pem to .gitignore"
echo ""
