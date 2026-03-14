#!/bin/bash
#
# Generate In-Proxy Keys for PsiphonNGLinux
#
# This script generates the Ed25519 private key and obfuscation secret
# required to run psiphond-ng in proxy mode.
#
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check for go
if ! command -v go &> /dev/null; then
    log_error "Go is not installed. Please install Go 1.21+"
    exit 1
fi

log_info "Generating In-Proxy keys..."

# Create temporary Go program
TMP_DIR=$(mktemp -d)
TMP_PROG="$TMP_DIR/genkeys.go"

cat > "$TMP_PROG" << 'EOF'
package main

import (
    "crypto/ed25519"
    "encoding/base64"
    "fmt"
    "math/rand"
    "time"
)

func main() {
    rand.Seed(time.Now().UnixNano())

    // Generate Ed25519 key pair
    privateKey, publicKey := ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))

    // Generate 32-byte root obfuscation secret
    secret := make([]byte, 32)
    rand.Read(secret)

    fmt.Println("=== In-Proxy Keys for PsiphonNGLinux ===")
    fmt.Println()
    fmt.Println("1. PRIVATE KEY (keep secret! add to config):")
    fmt.Println(base64.StdEncoding.EncodeToString(privateKey.Seed()))
    fmt.Println()
    fmt.Println("2. ROOT OBSCURATION SECRET (keep secret! add to config):")
    fmt.Println(base64.StdEncoding.EncodeToString(secret))
    fmt.Println()
    fmt.Println("3. PUBLIC KEY (share with broker):")
    fmt.Println(base64.StdEncoding.EncodeToString(publicKey.Bytes()))
    fmt.Println()
    fmt.Println("⚠️  IMPORTANT:")
    fmt.Println("   - Save the PRIVATE KEY and OBSCURATION SECRET in your config")
    fmt.Println("   - Share the PUBLIC KEY with your broker administrator")
    fmt.Println("   - Never share the private key or secret with anyone")
    fmt.Println("   - These keys are generated uniquely for each proxy instance")
    fmt.Println()
    fmt.Println("Add to your config (/etc/psiphon/psiphond-ng.conf):")
    fmt.Println(`  "inproxy_session_private_key": "YOUR_PRIVATE_KEY",`)
    fmt.Println(`  "inproxy_session_root_obfuscation_secret": "YOUR_SECRET"`)
}
EOF

# Run the program
go run "$TMP_PROG"

# Clean up
rm -rf "$TMP_DIR"

log_success "Keys generation complete!"
