# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based command-line tool called **Agent Manager** for managing Claude Code subagents from multiple sources. It provides YAML-driven configuration with intelligent conflict resolution, file transformations, installation tracking, marketplace integration, and advanced query capabilities for agent discovery and management.

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

### Release & Deployment
- `make release` - Create release artifacts for all platforms
- `make docker-build` - Build Docker image

### Benchmarking & Performance
- `make benchmark` - Run performance benchmarks
- `make benchmark-quick` - Run quick performance benchmarks
- `make benchmark-profile` - Run benchmarks with detailed profiling
- `make benchmark-clean` - Clean benchmark results

### Using the Tool
- `./bin/agent-manager install` - Install agents from all sources
- `./bin/agent-manager install --source <name>` - Install specific source
- `./bin/agent-manager list` - List installed agents
- `./bin/agent-manager update` - Update agents
- `./bin/agent-manager validate` - Validate agents-config.yaml
- `./bin/agent-manager query` - Query and discover agents using advanced search
- `./bin/agent-manager marketplace` - Browse and install from marketplace

## Architecture

The codebase follows a modular Go architecture:

```text
cmd/agent-manager/    # CLI interface using cobra
internal/
├── cli/             # Command line interface and command registry
├── config/          # YAML configuration parsing with variable substitution
├── installer/       # Core installation logic with source handlers (GitHub, Git, Local)
├── transformer/     # File transformation engine (remove prefixes, extract docs)
├── tracker/         # Installation state management with JSON tracking
├── conflict/        # Conflict resolution strategies (backup, overwrite, skip, merge)
├── query/           # Advanced query engine for agent discovery and search
├── marketplace/     # Marketplace integration for subagents.sh
├── progress/        # Progress tracking and display
├── util/            # Utility functions and helpers
└── types/           # Common type definitions
```

### Key Components

- **Source Handlers**: Support GitHub repositories, Git repositories, and local file systems
- **Transformations**: Built-in transformations for removing numeric prefixes and extracting documentation
- **Conflict Resolution**: Intelligent handling of existing files with multiple strategies (backup, overwrite, skip, merge)
- **Installation Tracking**: Complete file-level tracking with hashes and timestamps
- **YAML Configuration**: `agents-config.yaml` drives all behavior with variable substitution support
- **Query Engine**: Advanced search and filtering capabilities for agent discovery
- **Marketplace Integration**: Browser-based marketplace browsing and installation from subagents.sh
- **Progress Tracking**: Real-time progress indication for long-running operations

### Configuration Structure

The main configuration file `agents-config.yaml` defines:

- Global settings (base directories, conflict strategies, timeouts)
- Source definitions with type, repository, authentication, and filters
- File transformations and post-install actions
- Query configurations for agent discovery and search
- Marketplace settings and browser preferences
- Metadata for tracking and logging

### Variable Substitution

The configuration supports variable substitution:

- `${settings.key}` - References settings values
- `${env.VAR}` - References environment variables

## Prerequisites

### Required

- **Go 1.24.11+** - For building and running Agent Manager
- **Git** - For version control and repository operations
- **Make** - For build commands (recommended, but optional)

### Marketplace Integration (Optional)

- **Chrome/Chromium/Brave browser** - Required for subagents.sh marketplace integration
  - Supported browsers: Google Chrome, Chromium, Chrome Canary, Microsoft Edge (Chromium-based), Brave Browser
  - See [Marketplace Guide](docs/commands/MARKETPLACE.md) for installation instructions

### Development (Optional)

- **GitHub CLI (gh)** - Enhanced GitHub integration
- **golangci-lint** - Code linting (runs automatically in CI)
- **entr** - Hot reload development mode

## Dependencies

Key Go dependencies: cobra (CLI), go-git (Git operations), yaml.v3 (configuration), chromedp (browser automation), ristretto (caching), progressbar (progress tracking), diff3 (merge operations), securejoin (path security)

## Testing Strategy

The project uses Go's standard testing framework. Run tests with coverage to ensure code quality. The Makefile provides convenient test targets for both basic testing and coverage reporting.
