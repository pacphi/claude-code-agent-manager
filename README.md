# Claude Code Agent Manager

[![Version](https://img.shields.io/github/v/release/pacphi/claude-code-agent-manager?include_prereleases)](https://github.com/pacphi/claude-code-agent-manager/releases)
[![License](https://img.shields.io/github/license/pacphi/claude-code-agent-manager)](LICENSE)
[![CI](https://github.com/pacphi/claude-code-agent-manager/actions/workflows/ci.yml/badge.svg)](https://github.com/pacphi/claude-code-agent-manager/actions/workflows/ci.yml)

A powerful YAML-driven system for managing Claude Code [subagents](https://docs.anthropic.com/en/docs/claude-code/sub-agents) from multiple sources.

## Quick Start

```bash
# Clone and build
gh repo clone pacphi/claude-code-agent-manager
cd claude-code-agent-manager
make build

# Install agents
./bin/agent-manager install

# List installed agents
./bin/agent-manager list
```

## Key Features

- **YAML Configuration** - Define sources and behaviors in `agents-config.yaml`
- **Multiple Sources** - GitHub, Git, Local filesystem support
- **Conflict Resolution** - Intelligent handling with backup strategies
- **Installation Tracking** - Complete file-level tracking
- **Transformations** - Remove prefixes, extract docs, custom scripts

## Documentation

### 🚀 Getting Started

- [**Installation**](docs/getting-started/INSTALLATION.md) - Setup and first install
- [**First Steps**](docs/getting-started/FIRST-STEPS.md) - Basic usage tutorial
- [**Prerequisites**](docs/getting-started/PREREQUISITES.md) - System requirements

### 📚 Command Reference

- [**Install**](docs/commands/INSTALL.md) - Installation commands
- [**Marketplace**](docs/commands/MARKETPLACE.md) - Browse subagents.sh
- [**Advanced**](docs/commands/ADVANCED.md) - Update, list, validate

### 📖 Guides

- [**Configuration**](docs/guides/CONFIGURATION.md) - Complete YAML guide
- [**Workflows**](docs/guides/WORKFLOWS.md) - Common usage patterns
- [**Troubleshooting**](docs/guides/TROUBLESHOOTING.md) - Problem resolution
- [**Conflict Resolution**](docs/guides/CONFLICT-RESOLUTION.md) - Handling conflicts

### 🔧 Reference

- [**CLI Reference**](docs/reference/CLI-REFERENCE.md) - All commands
- [**Config Schema**](docs/reference/CONFIG-SCHEMA.md) - YAML schema
- [**Exit Codes**](docs/reference/EXIT-CODES.md) - Error codes

### 👨‍💻 Development

- [**Build**](docs/development/BUILD.md) - Build instructions
- [**Docker**](docs/development/DOCKER.md) - Container deployment
- [**Architecture**](docs/development/ARCHITECTURE.md) - Technical details
- [**Contributing**](docs/development/CONTRIBUTING.md) - Contribution guide
- [**Release**](docs/development/RELEASE.md) - Release process

## Requirements

- **Go 1.24.6+** - Building Agent Manager
- **Git** - Repository operations
- **Chrome/Chromium/Brave** (optional) - Marketplace features

See [Prerequisites Guide](docs/getting-started/PREREQUISITES.md) for detailed setup.

## Support

- Run `agent-manager --help` for command help
- Use `agent-manager validate` to check configuration
- Report issues at [GitHub Issues](https://github.com/pacphi/claude-code-agent-manager/issues)

## License

MIT License - See [LICENSE](LICENSE) for details.
