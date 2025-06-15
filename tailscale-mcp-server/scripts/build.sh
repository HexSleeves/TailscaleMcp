#!/bin/bash

# Build script for Tailscale MCP Server
# Usage: ./scripts/build.sh [options]

set -euo pipefail

# Default values
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
BUILD_TIME=${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}
OUTPUT_DIR=${OUTPUT_DIR:-"bin"}
BINARY_NAME=${BINARY_NAME:-"tailscale-mcp-server"}

# Build flags
LDFLAGS="-ldflags -X github.com/hexsleeves/tailscale-mcp-server/version.Version=${VERSION} \
                  -X github.com/hexsleeves/tailscale-mcp-server/version.GitCommit=${COMMIT} \
                  -X github.com/hexsleeves/tailscale-mcp-server/version.BuildTime=${BUILD_TIME}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
CROSS_COMPILE=false
CLEAN=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --cross-compile)
            CROSS_COMPILE=true
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --cross-compile  Build for all supported platforms"
            echo "  --clean          Clean build artifacts before building"
            echo "  --verbose        Enable verbose output"
            echo "  --help           Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Clean if requested
if [[ "$CLEAN" == "true" ]]; then
    log_info "Cleaning build artifacts..."
    rm -rf "$OUTPUT_DIR"
    rm -rf dist/
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build information
log_info "Building Tailscale MCP Server"
log_info "Version: $VERSION"
log_info "Commit: $COMMIT"
log_info "Build Time: $BUILD_TIME"

if [[ "$CROSS_COMPILE" == "true" ]]; then
    # Cross-compile for multiple platforms
    platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
    )

    for platform in "${platforms[@]}"; do
        IFS='/' read -r GOOS GOARCH <<< "$platform"
        output_name="$OUTPUT_DIR/${BINARY_NAME}-${GOOS}-${GOARCH}"
        
        if [[ "$GOOS" == "windows" ]]; then
            output_name="${output_name}.exe"
        fi

        log_info "Building for $GOOS/$GOARCH..."
        
        if [[ "$VERBOSE" == "true" ]]; then
            GOOS="$GOOS" GOARCH="$GOARCH" go build $LDFLAGS -o "$output_name" ./cmd/tailscale-mcp-server
        else
            GOOS="$GOOS" GOARCH="$GOARCH" go build $LDFLAGS -o "$output_name" ./cmd/tailscale-mcp-server 2>/dev/null
        fi
        
        if [[ $? -eq 0 ]]; then
            log_info "✓ Built $output_name"
        else
            log_error "✗ Failed to build for $GOOS/$GOARCH"
            exit 1
        fi
    done
else
    # Build for current platform
    output_name="$OUTPUT_DIR/$BINARY_NAME"
    
    log_info "Building for current platform ($(go env GOOS)/$(go env GOARCH))..."
    
    if [[ "$VERBOSE" == "true" ]]; then
        go build $LDFLAGS -o "$output_name" ./cmd/tailscale-mcp-server
    else
        go build $LDFLAGS -o "$output_name" ./cmd/tailscale-mcp-server 2>/dev/null
    fi
    
    if [[ $? -eq 0 ]]; then
        log_info "✓ Built $output_name"
        
        # Make executable
        chmod +x "$output_name"
        
        # Show binary info
        if command -v file >/dev/null 2>&1; then
            log_info "Binary info: $(file "$output_name")"
        fi
        
        # Show size
        if command -v du >/dev/null 2>&1; then
            size=$(du -h "$output_name" | cut -f1)
            log_info "Binary size: $size"
        fi
    else
        log_error "✗ Build failed"
        exit 1
    fi
fi

log_info "Build completed successfully!"
