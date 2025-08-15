#!/bin/bash

# Build script for agent-manager
set -e

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${GREEN}Building agent-manager for ${OS}/${ARCH}...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.24.6"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo -e "${YELLOW}Warning: Go version $GO_VERSION is installed, but $REQUIRED_VERSION or later is recommended${NC}"
fi

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Build the binary
echo "Building binary..."
make build

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Build successful!${NC}"
    echo ""
    echo "Binary location: bin/agent-manager"
    echo ""
    echo "To install system-wide, run:"
    echo "  make install"
    echo ""
    echo "To run directly:"
    echo "  ./bin/agent-manager --help"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi