#!/bin/bash

# Install script for agent-manager
set -e

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Installation directory
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="agent-manager"

# Check if running as root when needed
check_permissions() {
    if [ ! -w "$INSTALL_DIR" ]; then
        if [ "$EUID" -ne 0 ]; then
            echo -e "${YELLOW}Installation to $INSTALL_DIR requires sudo privileges${NC}"
            echo "Please run: sudo $0"
            exit 1
        fi
    fi
}

# Build the binary
echo -e "${GREEN}Building agent-manager...${NC}"
./scripts/build.sh

if [ ! -f "bin/$BINARY_NAME" ]; then
    echo -e "${RED}Build failed or binary not found${NC}"
    exit 1
fi

# Check permissions for installation directory
check_permissions

# Install the binary
echo -e "${GREEN}Installing to $INSTALL_DIR...${NC}"
cp "bin/$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Verify installation
if command -v $BINARY_NAME &> /dev/null; then
    VERSION=$($BINARY_NAME version 2>/dev/null | head -n1)
    echo -e "${GREEN}✓ Installation successful!${NC}"
    echo ""
    echo "Installed: $INSTALL_DIR/$BINARY_NAME"
    echo "Version: $VERSION"
    echo ""
    echo "Run 'agent-manager --help' to get started"
else
    echo -e "${RED}✗ Installation verification failed${NC}"
    echo "Binary was copied but cannot be executed"
    echo "Check your PATH environment variable"
    exit 1
fi

# Offer to install shell completions
echo ""
echo -e "${YELLOW}Would you like to install shell completions? (y/n)${NC}"
read -r response

if [[ "$response" =~ ^[Yy]$ ]]; then
    # Detect shell
    SHELL_NAME=$(basename "$SHELL")

    case $SHELL_NAME in
        bash)
            COMPLETION_DIR="${HOME}/.bash_completion.d"
            mkdir -p "$COMPLETION_DIR"
            $BINARY_NAME completion bash > "$COMPLETION_DIR/$BINARY_NAME"
            echo "source $COMPLETION_DIR/$BINARY_NAME" >> "${HOME}/.bashrc"
            echo -e "${GREEN}Bash completions installed. Restart your shell or run: source ~/.bashrc${NC}"
            ;;
        zsh)
            COMPLETION_DIR="${HOME}/.zsh/completions"
            mkdir -p "$COMPLETION_DIR"
            $BINARY_NAME completion zsh > "$COMPLETION_DIR/_$BINARY_NAME"
            echo "fpath=($COMPLETION_DIR \$fpath)" >> "${HOME}/.zshrc"
            echo -e "${GREEN}Zsh completions installed. Restart your shell or run: source ~/.zshrc${NC}"
            ;;
        fish)
            COMPLETION_DIR="${HOME}/.config/fish/completions"
            mkdir -p "$COMPLETION_DIR"
            $BINARY_NAME completion fish > "$COMPLETION_DIR/$BINARY_NAME.fish"
            echo -e "${GREEN}Fish completions installed. Restart your shell${NC}"
            ;;
        *)
            echo -e "${YELLOW}Unsupported shell: $SHELL_NAME${NC}"
            echo "You can generate completions manually with: $BINARY_NAME completion <bash|zsh|fish>"
            ;;
    esac
fi