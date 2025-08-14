# Building Agent Manager

This guide covers how to build, install, and develop Agent Manager.

## Prerequisites

### Required Tools

- **Go 1.23 or later**: Download from [golang.org](https://golang.org/dl/)
- **Git**: For version control and cloning repositories
- **Make** (optional): For convenient build commands

### Optional Tools

- **GitHub CLI (gh)**: For enhanced GitHub repository support
- **golangci-lint**: For code linting during development

## Quick Build

The fastest way to build Agent Manager:

```bash
# Using the convenience script
./scripts/build.sh

# Or using Make
make build

# Or manually with Go
go build -o bin/agent-manager cmd/agent-manager/main.go
```

## Installation

### System-wide Installation

Install Agent Manager to `/usr/local/bin` (requires sudo):

```bash
# Using the install script
./scripts/install.sh

# Or using Make
make install
```

### Manual Installation

```bash
# Build first
make build

# Copy to desired location
cp bin/agent-manager /usr/local/bin/
chmod +x /usr/local/bin/agent-manager
```

### Verify Installation

```bash
# Check if installed correctly
agent-manager version

# Validate your configuration
agent-manager validate
```

## Development

### Setting Up Development Environment

1. **Clone the repository**:

   ```bash
   git clone https://github.com/pacphi/claude-code-agent-manager
   cd claude-code-agent-template
   ```

2. **Install dependencies**:

   ```bash
   make deps
   ```

3. **Build and test**:

   ```bash
   make build
   make test
   ```

### Development Workflow

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests with coverage
make test-coverage

# Build and run
make run

# Development mode with hot reload (requires entr)
make dev
```

### Cross-Platform Building

Build for multiple platforms:

```bash
# Build for all supported platforms
make cross-compile

# Create release artifacts
make release
```

Supported platforms:

- macOS (AMD64, ARM64)
- Linux (AMD64, ARM64)
- Windows (AMD64)

## Build Options

### Makefile Targets

| Target | Description |
|--------|-------------|
| `build` | Build the binary for current platform |
| `install` | Install to `/usr/local/bin` |
| `test` | Run all tests |
| `test-coverage` | Run tests with coverage report |
| `clean` | Remove build artifacts |
| `fmt` | Format Go code |
| `lint` | Run Go linter |
| `vet` | Run Go vet |
| `cross-compile` | Build for multiple platforms |
| `release` | Create release artifacts |
| `deps` | Download and tidy dependencies |

### Build Variables

Customize the build with environment variables:

```bash
# Custom version
VERSION=v1.0.0 make build

# Custom binary name
BINARY_NAME=my-agent-manager make build

# Enable verbose output
GOFLAGS=-v make build
```

### LDFLAGS

The build includes version information:

```bash
go build -ldflags "
  -X main.version=$(git describe --tags --always --dirty)
  -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)
" cmd/agent-manager/main.go
```

## Docker Build

Build a Docker image:

```bash
# Build Docker image
make docker-build

# Run in container
docker run --rm -v $(pwd):/workspace agent-manager:latest validate
```

## Troubleshooting

### Common Build Issues

**Go not found**:

```bash
# macOS
brew install go

# Ubuntu/Debian
sudo apt-get install golang-go

# Or download from golang.org
```

**Permission denied during install**:

```bash
# Use sudo for system-wide install
sudo make install

# Or install to user directory
mkdir -p ~/bin
cp bin/agent-manager ~/bin/
export PATH="$HOME/bin:$PATH"
```

**Git command not found** (when using Git sources):

```bash
# macOS
xcode-select --install

# Ubuntu/Debian
sudo apt-get install git

# CentOS/RHEL
sudo yum install git
```

**GitHub CLI not found** (optional):

```bash
# macOS
brew install gh

# Ubuntu/Debian
sudo apt-get install gh

# Or download from github.com/cli/cli
```

### Build Performance

Speed up builds:

```bash
# Use Go module proxy
export GOPROXY=https://proxy.golang.org,direct

# Enable module caching
export GOMODCACHE=$HOME/.cache/go-mod

# Parallel builds
export GOMAXPROCS=$(nproc)
```

### Development Tips

**Enable debug logging**:

```bash
export AGENT_MANAGER_LOG_LEVEL=debug
./bin/agent-manager install
```

**Use custom config for testing**:

```bash
./bin/agent-manager --config test-config.yaml validate
```

**Dry run mode**:

```bash
./bin/agent-manager install --dry-run
```

## IDE Integration

### VS Code

Recommended extensions:

- Go (official Google extension)
- Go Test Explorer
- golangci-lint

### GoLand/IntelliJ

The project includes standard Go module structure and should work out of the box.

## Testing

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/config

# With coverage
make test-coverage

# Verbose output
go test -v ./...
```

### Writing Tests

Follow Go testing conventions:

- Test files end with `_test.go`
- Test functions start with `Test`
- Use table-driven tests where appropriate

Example:

```go
func TestConfigLoad(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected *Config
        wantErr  bool
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```
