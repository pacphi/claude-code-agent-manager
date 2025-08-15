# Claude Code Agent Manager

A powerful, YAML-driven system for managing Claude Code agents from multiple sources with intelligent conflict resolution, transformations, and tracking.

## Overview

This repository provides **Agent Manager**, a Go-based command-line tool that revolutionizes how you install and manage Claude Code agents. Unlike the previous shell script approach, Agent Manager offers:

- **YAML-driven configuration** for complete control
- **Multiple source types** (GitHub, Git, Local)
- **Intelligent conflict resolution** with backup strategies
- **File transformations** for consistent organization
- **Complete installation tracking** for easy management
- **Atomic operations** with rollback capabilities

> [!Note]
> The default configuration includes VoltAgent's [Awesome Claude Code Subagents](https://github.com/VoltAgent/awesome-claude-code-subagents)

## Quick Start

### 1. Clone and Build

```bash
gh repo clone pacphi/claude-code-agent-manager
cd claude-code-agent-manager

# Build the agent manager
make build
# or use the convenience script
./scripts/build.sh
```

### 2. Install Agents

```bash
# Install from all configured sources
./bin/agent-manager install

# Or install specific source
./bin/agent-manager install --source awesome-claude-code-subagents
```

### 3. Manage Your Installation

```bash
# List installed agents
./bin/agent-manager list

# Update agents
./bin/agent-manager update

# Validate configuration
./bin/agent-manager validate
```

## Key Features

### 🔧 YAML Configuration

All agent sources and behaviors are defined in `agents-config.yaml`:

```yaml
sources:
  - name: awesome-claude-code-subagents
    type: github
    repository: VoltAgent/awesome-claude-code-subagents
    transformations:
      - type: remove_numeric_prefix
      - type: extract_docs
```

### 🔄 Multiple Source Types

- **GitHub**: Direct GitHub repository access with CLI integration
- **Git**: Any Git repository with authentication support
- **Local**: Local file system sources for development

### ⚡ Conflict Resolution

Intelligent handling of existing files:
- **Backup**: Create timestamped backups (default)
- **Overwrite**: Replace existing files
- **Skip**: Keep existing, ignore new
- **Merge**: Attempt content merging

### 🔍 Installation Tracking

Complete tracking of what was installed from where:

- File-level tracking with hashes and timestamps
- Easy uninstallation of specific sources
- Update detection and management

### 🛠 File Transformations

Built-in transformations for common patterns:

- Remove numeric prefixes from directories
- Extract documentation to separate directory
- Custom script support for specialized needs

## Architecture

Agent Manager is built with a modular Go architecture:

```bash
cmd/agent-manager/     # CLI interface
internal/
├── config/           # YAML parsing and validation
├── installer/        # Core installation logic
├── transformer/      # File transformation engine
├── tracker/          # Installation state management
└── conflict/         # Conflict resolution strategies
```

## Documentation

- **[Build Guide](docs/BUILD.md)**: Comprehensive build and installation instructions
- **[Usage Guide](docs/USAGE.md)**: Detailed command examples and workflows
- **[Configuration Reference](docs/CONFIG.md)**: Complete YAML configuration documentation
- **[Release Guide](docs/RELEASE.md)**: Release process and version management
- **[Architecture Overview](docs/README.md)**: Technical details and design decisions

## Configuration Examples

### Basic GitHub Source

```yaml
sources:
  - name: my-agents
    type: github
    repository: owner/repo-name
    paths:
      source: agents
      target: .claude/agents
```

### Authenticated Git Repository

```yaml
sources:
  - name: private-agents
    type: git
    url: https://gitlab.company.com/agents.git
    auth:
      method: token
      token_env: GITLAB_TOKEN
```

### Local Development

```yaml
sources:
  - name: local-dev
    type: local
    paths:
      source: ~/my-local-agents
      target: .claude/agents/local
    watch: true
```

## Requirements

- **Go 1.24.6+**: For building Agent Manager
- **Git**: For cloning repositories
- **GitHub CLI** (optional): Enhanced GitHub integration

## Advanced Usage

### Custom Transformations

```yaml
transformations:
  - type: custom_script
    script: ./my-transform.sh
    args: ["--format", "claude"]
```

### Environment Variables

```yaml
sources:
  - name: ${env.COMPANY}-agents
    repository: ${env.GITHUB_ORG}/claude-agents
    auth:
      token_env: ${env.TOKEN_VAR}
```

### Filtering

```yaml
filters:
  include:
    extensions: [".md"]
    regex: ["^(core|data).*\\.md$"]
  exclude:
    patterns: ["test-*", "*.tmp"]
```

## Contributing

We welcome contributions! Agent Manager is designed to be extensible:

- Add new source handlers in `internal/installer/`
- Create custom transformations in `internal/transformer/`
- Implement new conflict resolution strategies
- Extend the configuration schema

## License

This project is licensed under the MIT License. See LICENSE file for details.

## Support

- Run `agent-manager --help` for command-line help
- Use `agent-manager validate` to check your configuration
- See the [Usage Guide](docs/USAGE.md) for detailed examples
- Check the [Build Guide](docs/BUILD.md) for installation help
