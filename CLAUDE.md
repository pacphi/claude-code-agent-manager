# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based command-line tool called **Agent Manager** for managing Claude Code agents from multiple sources. It provides YAML-driven configuration with intelligent conflict resolution, file transformations, and installation tracking.

## Common Commands

### Building
- `make build` - Build the agent-manager binary
- `make cross-compile` - Build for multiple platforms
- `./scripts/build.sh` - Alternative build script

### Testing
- `make test` - Run all tests with verbose output
- `make test-coverage` - Generate coverage report (creates coverage.html)

### Development
- `make fmt` - Format Go code
- `make lint` - Run golangci-lint (requires installation)
- `make vet` - Run go vet
- `make deps` - Download and tidy dependencies
- `make dev` - Hot reload development mode (requires entr)

### Installation & Management
- `make install` - Install to /usr/local/bin (requires sudo)
- `make run` - Build and run with install command
- `make validate` - Validate configuration file
- `make clean` - Remove build artifacts

### Using the Tool
- `./bin/agent-manager install` - Install agents from all sources
- `./bin/agent-manager install --source <name>` - Install specific source
- `./bin/agent-manager list` - List installed agents
- `./bin/agent-manager update` - Update agents
- `./bin/agent-manager validate` - Validate agents-config.yaml

## Architecture

The codebase follows a modular Go architecture:

```
cmd/agent-manager/    # CLI interface using cobra
internal/
├── config/          # YAML configuration parsing with variable substitution
├── installer/       # Core installation logic with source handlers (GitHub, Git, Local)
├── transformer/     # File transformation engine (remove prefixes, extract docs)
├── tracker/         # Installation state management with JSON tracking
└── conflict/        # Conflict resolution strategies (backup, overwrite, skip, merge)
```

### Key Components

- **Source Handlers**: Support GitHub repositories, Git repositories, and local file systems
- **Transformations**: Built-in transformations for removing numeric prefixes and extracting documentation
- **Conflict Resolution**: Intelligent handling of existing files with multiple strategies
- **Installation Tracking**: Complete file-level tracking with hashes and timestamps
- **YAML Configuration**: `agents-config.yaml` drives all behavior with variable substitution support

### Configuration Structure

The main configuration file `agents-config.yaml` defines:
- Global settings (base directories, conflict strategies, timeouts)
- Source definitions with type, repository, authentication, and filters
- File transformations and post-install actions
- Metadata for tracking and logging

### Variable Substitution

The configuration supports variable substitution:
- `${settings.key}` - References settings values
- `${env.VAR}` - References environment variables

## Dependencies

- Go 1.24.6+ required
- Key dependencies: cobra (CLI), go-git (Git operations), yaml.v3 (configuration), chromedp (browser automation)
- **Chrome/Chromium/Brave browser required** for subagents.sh marketplace integration (see MARKETPLACE.md)
- Optional: GitHub CLI for enhanced GitHub integration
- Optional: golangci-lint for linting
- Optional: entr for development hot reload

## Testing Strategy

The project uses Go's standard testing framework. Run tests with coverage to ensure code quality. The Makefile provides convenient test targets for both basic testing and coverage reporting.