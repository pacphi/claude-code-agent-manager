# Prerequisites

This document provides a comprehensive guide to all prerequisites needed to build and run Agent Manager.

## Quick Check

Before proceeding, ensure you have these core requirements:

```bash
# Check Go version (requires 1.24.11+)
go version

# Check Git availability
git --version

# Check Make availability (recommended)
make --version

# Check browser availability (for marketplace features)
which google-chrome || which chromium || which chromium-browser || echo "No browser found - marketplace features will not work"
```

## Core Requirements

### 1. Go Programming Language

**Required Version**: Go 1.24.11 or later

**Installation**:

- **Download**: Get the latest version from [golang.org](https://golang.org/dl/)
- **macOS**: `brew install go`
- **Ubuntu/Debian**: `sudo apt-get install golang-go`
- **CentOS/RHEL**: `sudo yum install golang`

**Verification**:

```bash
go version
# Should show: go version go1.24.11 or later
```

### 2. Git Version Control

**Purpose**: Required for cloning repositories and version control operations

**Installation**:

- **macOS**: `xcode-select --install` or `brew install git`
- **Ubuntu/Debian**: `sudo apt-get install git`
- **CentOS/RHEL**: `sudo yum install git`
- **Windows**: Download from [git-scm.com](https://git-scm.com/)

**Verification**:

```bash
git --version
```

### 3. Make (Recommended)

**Purpose**: Simplifies build commands and development workflows

**Installation**:

- **macOS**: Included with Xcode Command Line Tools
- **Ubuntu/Debian**: `sudo apt-get install build-essential`
- **CentOS/RHEL**: `sudo yum install make`
- **Windows**: Install via MinGW or use Windows Subsystem for Linux (WSL)

**Alternative**: You can use `go build` commands directly if Make is not available

## Marketplace Integration (Optional)

### Chrome/Chromium Browser

**Purpose**: Required for subagents.sh marketplace integration using browser automation

**Supported Browsers**:

- Google Chrome
- Chromium
- Chrome Canary
- Microsoft Edge (Chromium-based)
- Brave Browser

**Installation**:

#### macOS

```bash
# Google Chrome
brew install --cask google-chrome

# Chromium
brew install chromium

# Brave Browser
brew install --cask brave-browser

# Alternative: Download directly
# https://www.google.com/chrome/
# https://brave.com/
```

#### Linux (Ubuntu/Debian)

```bash
# Google Chrome
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" | sudo tee /etc/apt/sources.list.d/google-chrome.list
sudo apt-get update
sudo apt-get install google-chrome-stable

# Chromium
sudo apt-get install chromium-browser

# Brave Browser
sudo snap install brave
```

#### Linux (CentOS/RHEL/Fedora)

```bash
# Google Chrome
sudo yum install google-chrome-stable

# Chromium
sudo dnf install chromium

# Brave Browser
sudo dnf install brave-browser
```

#### Windows

```cmd
# Via Chocolatey
choco install googlechrome
choco install chromium
choco install brave

# Via Scoop
scoop install googlechrome
scoop install chromium
scoop install brave

# Alternative: Download directly from vendors
```

**Verification**:

```bash
# Check if browser is accessible
google-chrome --version || chromium --version || brave-browser --version || echo "No supported browser found"
```

**Note**: If no browser is installed, marketplace features will be unavailable but all other functionality works normally.

## Development Tools (Optional)

### 1. GitHub CLI (gh)

**Purpose**: Enhanced GitHub repository support and authentication

**Installation**:

- **macOS**: `brew install gh`
- **Ubuntu/Debian**: `sudo apt-get install gh`
- **Windows**: `choco install gh` or `scoop install gh`
- **Alternative**: Download from [github.com/cli/cli](https://github.com/cli/cli)

### 2. golangci-lint

**Purpose**: Code linting during development

**Installation**:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Note**: Linting runs automatically in CI, so local installation is optional

### 3. entr (File Watcher)

**Purpose**: Hot reload development mode (`make dev`)

**Installation**:

- **macOS**: `brew install entr`
- **Ubuntu/Debian**: `sudo apt-get install entr`
- **CentOS/RHEL**: Available in EPEL repository

## Verification Script

Copy and run this script to verify all prerequisites:

```bash
#!/bin/bash

echo "=== Agent Manager Prerequisites Check ==="
echo

# Check Go
echo "1. Checking Go..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    echo "   ✓ Go found: $GO_VERSION"
    if [[ "$GO_VERSION" < "1.24.11" ]]; then
        echo "   ⚠ Warning: Go 1.24.11+ required, found $GO_VERSION"
    fi
else
    echo "   ✗ Go not found - REQUIRED"
fi

# Check Git
echo "2. Checking Git..."
if command -v git &> /dev/null; then
    GIT_VERSION=$(git --version | cut -d' ' -f3)
    echo "   ✓ Git found: $GIT_VERSION"
else
    echo "   ✗ Git not found - REQUIRED"
fi

# Check Make
echo "3. Checking Make..."
if command -v make &> /dev/null; then
    echo "   ✓ Make found"
else
    echo "   ⚠ Make not found - Recommended for build commands"
fi

# Check browsers
echo "4. Checking browsers (for marketplace)..."
BROWSER_FOUND=false
for browser in google-chrome chromium chromium-browser brave-browser; do
    if command -v $browser &> /dev/null; then
        echo "   ✓ $browser found"
        BROWSER_FOUND=true
        break
    fi
done

if [ "$BROWSER_FOUND" = false ]; then
    echo "   ⚠ No supported browser found - Marketplace features unavailable"
fi

# Check optional tools
echo "5. Checking optional tools..."
if command -v gh &> /dev/null; then
    echo "   ✓ GitHub CLI found"
else
    echo "   - GitHub CLI not found (optional)"
fi

if command -v golangci-lint &> /dev/null; then
    echo "   ✓ golangci-lint found"
else
    echo "   - golangci-lint not found (optional)"
fi

if command -v entr &> /dev/null; then
    echo "   ✓ entr found"
else
    echo "   - entr not found (optional)"
fi

echo
echo "=== Summary ==="
echo "✓ = Available, ✗ = Missing (required), ⚠ = Warning, - = Optional"
```

## Troubleshooting

### Go Issues

**"go command not found"**:

- Ensure Go is installed and in your PATH
- Add Go's bin directory to PATH: `export PATH=$PATH:/usr/local/go/bin`

**Version too old**:

- Update Go to 1.24.11+ from [golang.org](https://golang.org/dl/)

### Browser Issues

**"executable file not found in $PATH"**:

- Install a supported browser using instructions above
- Marketplace features will be disabled without a browser

### Build Issues

**"make command not found"**:

- Use direct Go commands: `go build -o bin/agent-manager cmd/agent-manager/main.go`
- Or install Make using instructions above

**Permission denied**:

- Use `sudo make install` for system-wide installation
- Or install to user directory: `cp bin/agent-manager ~/bin/`

## Next Steps

Once prerequisites are satisfied:

1. **Clone the repository**: `git clone https://github.com/pacphi/claude-code-agent-manager`
2. **Build**: `make build` or `./scripts/build.sh`
3. **Install**: `make install` or copy binary manually
4. **Verify**: `agent-manager --version`

For detailed build instructions, see [BUILD.md](BUILD.md).
