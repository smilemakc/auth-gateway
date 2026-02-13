#!/bin/bash

# Generate self-signed TLS certificates for gRPC development
# Usage: ./generate-grpc-tls.sh [output_dir]
#
# For production, use certificates from a trusted CA (Let's Encrypt, etc.)

set -e

OUTPUT_DIR="${1:-./certs/grpc}"
DAYS=365
CN="${2:-localhost}"

echo "======================================"
echo "gRPC TLS Certificate Generation"
echo "======================================"
echo ""
echo "Output directory: $OUTPUT_DIR"
echo "Common Name (CN): $CN"
echo "Validity: $DAYS days"
echo ""

mkdir -p "$OUTPUT_DIR"

# Generate CA private key
echo "1. Generating CA private key..."
openssl genrsa -out "$OUTPUT_DIR/ca-key.pem" 4096 2>/dev/null
echo "   ✓ CA private key generated"

# Generate CA certificate
echo "2. Generating CA certificate..."
openssl req -new -x509 -days "$DAYS" \
    -key "$OUTPUT_DIR/ca-key.pem" \
    -out "$OUTPUT_DIR/ca-cert.pem" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=Auth Gateway/OU=Development/CN=Auth Gateway CA" \
    2>/dev/null
echo "   ✓ CA certificate generated"

# Generate server private key
echo "3. Generating server private key..."
openssl genrsa -out "$OUTPUT_DIR/server-key.pem" 2048 2>/dev/null
echo "   ✓ Server private key generated"

# Generate server CSR
echo "4. Generating server certificate signing request..."
openssl req -new \
    -key "$OUTPUT_DIR/server-key.pem" \
    -out "$OUTPUT_DIR/server.csr" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=Auth Gateway/OU=gRPC Server/CN=$CN" \
    2>/dev/null
echo "   ✓ Server CSR generated"

# Create extensions file for SAN
cat > "$OUTPUT_DIR/server-ext.cnf" << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = auth-gateway
DNS.3 = $CN
IP.1 = 127.0.0.1
IP.2 = 0.0.0.0
EOF

# Sign server certificate with CA
echo "5. Signing server certificate with CA..."
openssl x509 -req -days "$DAYS" \
    -in "$OUTPUT_DIR/server.csr" \
    -CA "$OUTPUT_DIR/ca-cert.pem" \
    -CAkey "$OUTPUT_DIR/ca-key.pem" \
    -CAcreateserial \
    -out "$OUTPUT_DIR/server-cert.pem" \
    -extfile "$OUTPUT_DIR/server-ext.cnf" \
    2>/dev/null
echo "   ✓ Server certificate signed"

# Generate client private key (for mTLS)
echo "6. Generating client private key (for mTLS)..."
openssl genrsa -out "$OUTPUT_DIR/client-key.pem" 2048 2>/dev/null
echo "   ✓ Client private key generated"

# Generate client CSR
echo "7. Generating client certificate signing request..."
openssl req -new \
    -key "$OUTPUT_DIR/client-key.pem" \
    -out "$OUTPUT_DIR/client.csr" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=Auth Gateway/OU=gRPC Client/CN=grpc-client" \
    2>/dev/null
echo "   ✓ Client CSR generated"

# Sign client certificate with CA
echo "8. Signing client certificate with CA..."
openssl x509 -req -days "$DAYS" \
    -in "$OUTPUT_DIR/client.csr" \
    -CA "$OUTPUT_DIR/ca-cert.pem" \
    -CAkey "$OUTPUT_DIR/ca-key.pem" \
    -CAcreateserial \
    -out "$OUTPUT_DIR/client-cert.pem" \
    2>/dev/null
echo "   ✓ Client certificate signed"

# Cleanup CSR and temp files
rm -f "$OUTPUT_DIR/server.csr" "$OUTPUT_DIR/client.csr" "$OUTPUT_DIR/server-ext.cnf" "$OUTPUT_DIR/ca-cert.srl"

echo ""
echo "======================================"
echo "Certificates generated successfully!"
echo "======================================"
echo ""
echo "Files:"
echo "  CA:     $OUTPUT_DIR/ca-cert.pem"
echo "  Server: $OUTPUT_DIR/server-cert.pem, $OUTPUT_DIR/server-key.pem"
echo "  Client: $OUTPUT_DIR/client-cert.pem, $OUTPUT_DIR/client-key.pem"
echo ""
echo "======================================"
echo "Configuration"
echo "======================================"
echo ""
echo "Add to your .env file:"
echo ""
echo "  GRPC_TLS_ENABLED=true"
echo "  GRPC_TLS_CERT_FILE=$OUTPUT_DIR/server-cert.pem"
echo "  GRPC_TLS_KEY_FILE=$OUTPUT_DIR/server-key.pem"
echo ""
echo "For gRPC clients, use:"
echo "  CA cert: $OUTPUT_DIR/ca-cert.pem"
echo ""
echo "Example (Go SDK):"
echo '  grpcClient, _ := authgateway.NewGRPCClient(authgateway.GRPCConfig{'
echo '      Address:  "localhost:50051",'
echo "      TLSCert:  \"$OUTPUT_DIR/ca-cert.pem\","
echo '      APIKey:   "agw_YOUR_API_KEY",'
echo '  })'
echo ""
echo "IMPORTANT:"
echo "  - These are SELF-SIGNED certificates for DEVELOPMENT only"
echo "  - For production, use certificates from a trusted CA"
echo "  - Add $OUTPUT_DIR/ to .gitignore"
echo ""
