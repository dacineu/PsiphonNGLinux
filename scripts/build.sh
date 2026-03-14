#!/bin/bash
#
# Build script for PsiphonNGLinux
#
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuration
BUILD_DIR="build"
OUTPUT_BINARY="psiphond-ng"
TAGS=""
LDFLAGS="-s -w"
VERSION="1.0.0"
COMMIT=""
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -o|--output)
            OUTPUT_BINARY="$2"
            shift 2
            ;;
        -t|--tags)
            TAGS="$2"
            shift 2
            ;;
        --ldflags)
            LDFLAGS="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        --static)
            LDFLAGS="$LDFLAGS -extldflags '-static'"
            shift
            ;;
        --compress)
            COMPRESS=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Build PsiphonNGLinux daemon"
            echo ""
            echo "Options:"
            echo "  -o, --output FILE     Output binary name (default: psiphond-ng)"
            echo "  -t, --tags TAGS       Go build tags (e.g., 'tun' for TUN support)"
            echo "  --ldflags FLAGS       Additional ldflags"
            echo "  -v, --version VER     Version string (default: 1.0.0)"
            echo "  --static              Build static binary"
            echo "  --compress            Compress binary with upx (if available)"
            echo "  -h, --help            Show this help"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Get commit if available
if git rev-parse --is-inside-work-tree &>/dev/null; then
    COMMIT=$(git rev-parse --short HEAD)
fi

# Create build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

log_info "Building PsiphonNGLinux $VERSION ($COMMIT)"
log_info "Build tags: ${TAGS:-none}"
log_info "Output: $OUTPUT_BINARY"

# Build
BUILD_CMD="go build -trimpath -ldflags='$LDFLAGS' -o $BUILD_DIR/$OUTPUT_BINARY"
if [[ -n "$TAGS" ]]; then
    BUILD_CMD="$BUILD_CMD -tags '$TAGS'"
fi

# Add version info
LDFLAGS_VERSION="-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"
LDFLAGS="$LDFLAGS $LDFLAGS_VERSION"

# Execute build
log_info "Running: go build -trimpath -ldflags='$LDFLAGS' -o $BUILD_DIR/$OUTPUT_BINARY"
if [[ -n "$TAGS" ]]; then
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="$LDFLAGS" -tags "$TAGS" -o "$BUILD_DIR/$OUTPUT_BINARY" ./cmd/psiphond-ng
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/$OUTPUT_BINARY" ./cmd/psiphond-ng
fi

if [[ $? -ne 0 ]]; then
    log_error "Build failed"
    exit 1
fi

# Check binary size
BINARY_SIZE=$(du -h "$BUILD_DIR/$OUTPUT_BINARY" | cut -f1)
log_info "Binary built: $BUILD_DIR/$OUTPUT_BINARY (size: $BINARY_SIZE)"

# Optional compression
if [[ "$COMPRESS" == "true" ]]; then
    if command -v upx &>/dev/null; then
        log_info "Compressing binary with upx..."
        upx --best --lzma "$BUILD_DIR/$OUTPUT_BINARY"
        COMPRESSED_SIZE=$(du -h "$BUILD_DIR/$OUTPUT_BINARY" | cut -f1)
        log_info "Compressed size: $COMPRESSED_SIZE"
    else
        log_warn "upx not found, skipping compression"
    fi
fi

log_info "Build complete!"
log_info ""
log_info "Artifacts:"
log_info "  Binary: $BUILD_DIR/$OUTPUT_BINARY"
log_info ""
log_info "To install:"
log_info "  sudo cp $BUILD_DIR/$OUTPUT_BINARY /usr/local/bin/"
log_info "  sudo ./scripts/install.sh"
