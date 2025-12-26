# Installation Guide

A step-by-step guide to installing and setting up Agent Manager.

## Prerequisites Check

Before installation, verify you have the required tools:

```bash
# Check Go version (requires 1.24.11+)
go version

# Check Git installation
git --version

# Optional: Check for browser (for marketplace features)
which chrome || which chromium || which brave
```

See the [Prerequisites Guide](PREREQUISITES.md) for detailed requirements.

## Installation Steps

### Step 1: Clone the Repository

```bash
# Using GitHub CLI (recommended)
gh repo clone pacphi/claude-code-agent-manager

# Or using Git
git clone https://github.com/pacphi/claude-code-agent-manager.git

# Enter the directory
cd claude-code-agent-manager
```

### Step 2: Build Agent Manager

```bash
# Using Make (recommended)
make build

# Or using the build script
./scripts/build.sh

# Or directly with Go
go build -o bin/agent-manager cmd/agent-manager/main.go
```

You should see the binary created at `bin/agent-manager`.

### Step 3: Verify Installation

```bash
# Check the binary works
./bin/agent-manager --help

# Verify version
./bin/agent-manager version
```

### Step 4: System-wide Installation (Optional)

To use `agent-manager` from anywhere:

```bash
# Install to /usr/local/bin (requires sudo)
sudo make install

# Or manually copy
sudo cp bin/agent-manager /usr/local/bin/

# Verify system-wide installation
agent-manager --help
```

## Configuration Setup

### Step 1: Review Default Configuration

The repository includes a default `agents-config.yaml`:

```bash
# View the configuration
cat agents-config.yaml
```

### Step 2: Customize Configuration (Optional)

Create your own configuration:

```bash
# Copy default config
cp agents-config.yaml my-config.yaml

# Edit with your preferences
nano my-config.yaml
```

### Step 3: Validate Configuration

```bash
# Validate default config
./bin/agent-manager validate

# Or validate custom config
./bin/agent-manager validate --config my-config.yaml
```

## First Installation

### Install Default Agents

```bash
# Install from all configured sources
./bin/agent-manager install
```

Expected output:

```text
Installing from source: awesome-claude-code-subagents
✓ Fetched repository
✓ Applied transformations
✓ Installed 25 agents
Installation complete!
```

### Verify Installation

```bash
# List installed agents
./bin/agent-manager list

# Check installation directory
ls -la ~/.claude/agents/
```

## Troubleshooting Installation

### Go Not Found

If `go version` fails:

1. Install Go from [go.dev](https://go.dev/dl/)
2. Add Go to your PATH:

   ```bash
   export PATH=$PATH:/usr/local/go/bin
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   ```

### Build Failures

If the build fails:

```bash
# Clean and retry
make clean
make deps
make build
```

### Permission Errors

For permission issues during installation:

```bash
# Ensure correct permissions
chmod +x bin/agent-manager
chmod +x scripts/*.sh

# For system installation, use sudo
sudo make install
```

### Configuration Issues

If validation fails:

```bash
# Check YAML syntax
./bin/agent-manager validate --verbose

# Use default config
cp agents-config.yaml.example agents-config.yaml
```

## Platform-Specific Notes

### macOS

- Ensure Xcode Command Line Tools are installed:

  ```bash
  xcode-select --install
  ```

### Linux

- May need to install build-essential:

  ```bash
  sudo apt-get install build-essential  # Debian/Ubuntu
  sudo yum groupinstall "Development Tools"  # RHEL/CentOS
  ```

### Windows

- Use WSL2 for best compatibility
- Or use Git Bash for basic functionality

## Next Steps

Once installation is complete:

1. Read [First Steps](FIRST-STEPS.md) for basic usage
2. Explore [Configuration Guide](../guides/CONFIGURATION.md)
3. Learn about [Workflows](../guides/WORKFLOWS.md)

## Getting Help

- Run `agent-manager --help` for command help
- Check [Troubleshooting Guide](../guides/TROUBLESHOOTING.md)
- Report issues at [GitHub Issues](https://github.com/pacphi/claude-code-agent-manager/issues)
