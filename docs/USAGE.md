# Agent Manager Usage Guide

This guide provides detailed instructions on how to use Agent Manager for managing Claude Code agents.

## Command Overview

Agent Manager provides several commands for managing your agents:

```bash
agent-manager [command] [options]
```

### Available Commands

| Command | Description |
|---------|-------------|
| `install` | Install agents from configured sources |
| `uninstall` | Remove installed agents |
| `update` | Update existing installations |
| `list` | List installed agents |
| `marketplace` | Browse and discover agents from subagents.sh |
| `validate` | Validate configuration file |
| `version` | Show version information |

### Global Options

| Option | Description |
|--------|-------------|
| `--config, -c` | Configuration file (default: agents-config.yaml) |
| `--verbose, -v` | Verbose output |
| `--dry-run` | Simulate actions without making changes |
| `--no-color` | Disable colored output |

## Installation Commands

### Install All Enabled Sources

Install agents from all enabled sources in your configuration:

```bash
agent-manager install
```

### Install Specific Source

Install agents from a specific source only:

```bash
agent-manager install --source awesome-claude-code-subagents
```

### Dry Run Installation

Preview what would be installed without making changes:

```bash
agent-manager install --dry-run
```

### Verbose Installation

See detailed output during installation:

```bash
agent-manager install --verbose
```

## Uninstall Commands

### Uninstall Specific Source

Remove agents from a specific source:

```bash
agent-manager uninstall --source awesome-claude-code-subagents
```

### Uninstall All Sources

Remove all installed agents:

```bash
agent-manager uninstall --all
```

### Keep Backups During Uninstall

Preserve backup files when uninstalling:

```bash
agent-manager uninstall --source my-agents --keep-backups
```

## Update Commands

### Update All Sources

Check for and apply updates to all installed sources:

```bash
agent-manager update
```

### Update Specific Source

Update a specific source only:

```bash
agent-manager update --source awesome-claude-code-subagents
```

### Check for Updates Only

See what updates are available without applying them:

```bash
agent-manager update --check-only
```

## List Commands

### List All Installed Agents

Show all installed agents:

```bash
agent-manager list
```

### List Agents from Specific Source

Show agents from a specific source:

```bash
agent-manager list --source awesome-claude-code-subagents
```

### Verbose Listing

Show detailed information about installations:

```bash
agent-manager list --verbose
```

## Marketplace Commands

The marketplace integration allows you to browse and discover agents from the [subagents.sh](https://subagents.sh) marketplace.

### Prerequisites

**Browser Requirement**: Chrome, Chromium, or a Chromium-based browser must be installed and accessible in your PATH.

**Supported browsers:**
- Google Chrome
- Chromium
- Chrome Canary
- Microsoft Edge (Chromium-based)
- Brave Browser

**Installation options:**

**macOS:**
```bash
# Via Homebrew
brew install --cask google-chrome
# or
brew install chromium
# or
brew install --cask brave-browser
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install google-chrome-stable
# or
sudo apt-get install chromium-browser

# CentOS/RHEL/Fedora
sudo yum install google-chrome-stable
# or
sudo dnf install chromium
```

**Windows:**
```cmd
# Via Chocolatey
choco install googlechrome
# or
choco install chromium
# or
choco install brave
```

### Browse Categories

List all available agent categories:

```bash
agent-manager marketplace list
```

### Browse Agents by Category

List agents in a specific category:

```bash
agent-manager marketplace list --category "Development"
agent-manager marketplace list --category "AI & ML"
agent-manager marketplace list --category "Writing"
```

### Show Agent Details

View detailed information about a specific agent:

```bash
agent-manager marketplace show "code-reviewer"
```

Show agent with full definition (when available):

```bash
agent-manager marketplace show "documentation-expert" --content
```

### Refresh Cache

Update cached marketplace data:

```bash
agent-manager marketplace refresh
```

### Installing Marketplace Agents

Once you find an agent you want to install, add it to your `agents-config.yaml`:

```yaml
sources:
  - name: subagents-marketplace-all
    enabled: true
    type: subagents
    paths:
      target: ${settings.base_dir}/marketplace
    cache:
      enabled: true
      ttl_hours: 1
      max_size_mb: 50
```

Or target a specific agent:

```yaml
sources:
  - name: specific-agent
    enabled: true
    type: subagents
    category: "Development"  # Optional: filter by category
    paths:
      target: .claude/agents
    filters:
      include:
        regex: ["^agent-name.*\\.md$"]  # Filter for specific agent
```

Then run:

```bash
agent-manager install
```

### Troubleshooting Marketplace

**"executable file not found in $PATH"**: Chrome/Chromium is not installed or not in your PATH. Install a supported browser using the installation options above.

**"context deadline exceeded"**: Indicates slow network connection or browser startup issues. Try increasing the timeout or checking your network connection.

**"No categories found" or "No agents found"**: May indicate changes to the subagents.sh website structure or JavaScript execution issues. Check logs and consider refreshing the cache.

Enable verbose logging for detailed browser automation logs:

```bash
agent-manager marketplace list --verbose
```

## Docker Usage

Agent Manager can be run in a Docker container for isolated, reproducible deployments. This is particularly useful for CI/CD pipelines, cloud deployments, or when you want to avoid installing dependencies on your host system.

### Prerequisites

- Docker installed and running
- Docker Compose (optional, for easier management)

### Building the Docker Image

Build the Docker image using the provided Dockerfile:

```bash
# Using Make
make docker-build

# Or using Docker directly
docker build -t agent-manager:latest .

# Build with specific version tag
docker build -t agent-manager:v1.0.0 --build-arg VERSION=v1.0.0 .
```

### Running with Docker

#### Basic Usage

Run Agent Manager commands in a container:

```bash
# Show help
docker run --rm agent-manager:latest --help

# Validate configuration
docker run --rm \
  -v $(pwd)/agents-config.yaml:/app/agents-config.yaml:ro \
  agent-manager:latest validate

# Install agents
docker run --rm \
  -v $(pwd)/agents-config.yaml:/app/agents-config.yaml:ro \
  -v agent-data:/app/.claude/agents \
  -v tracker-data:/app/.agent-manager \
  agent-manager:latest install

# List installed agents
docker run --rm \
  -v tracker-data:/app/.agent-manager:ro \
  agent-manager:latest list
```

#### With Environment Variables

Pass environment variables for authentication:

```bash
docker run --rm \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  -v $(pwd)/agents-config.yaml:/app/agents-config.yaml:ro \
  -v agent-data:/app/.claude/agents \
  agent-manager:latest install
```

### Using Docker Compose

Docker Compose provides easier volume and configuration management:

#### Starting the Service

```bash
# Build the image
docker-compose build

# Run install command
docker-compose run agent-manager install

# Run with specific source
docker-compose run agent-manager install --source awesome-claude-code-subagents

# List installed agents
docker-compose run agent-manager list

# Update agents
docker-compose run agent-manager update

# Validate configuration
docker-compose run agent-manager validate
```

#### With Custom Configuration

Override default settings using environment variables:

```bash
# Set version and log level
VERSION=v1.0.0 LOG_LEVEL=debug docker-compose run agent-manager install

# Use different config file
docker-compose run \
  -v $(pwd)/custom-config.yaml:/app/agents-config.yaml:ro \
  agent-manager install
```

### Volume Mounts

The Docker setup uses volumes for persistent data:

| Volume/Mount | Purpose | Mode |
|--------------|---------|------|
| `/app/agents-config.yaml` | Configuration file | Read-only |
| `/app/.claude/agents` | Installed agents directory | Read-write |
| `/app/.agent-manager` | Installation tracking data | Read-write |

Example with custom paths:

```bash
docker run --rm \
  -v /path/to/config.yaml:/app/agents-config.yaml:ro \
  -v /path/to/agents:/app/.claude/agents \
  -v /path/to/tracker:/app/.agent-manager \
  agent-manager:latest install
```

### Security Features

The Docker image includes multiple security features:

- **Minimal base image**: Uses `scratch` (empty) image, reducing attack surface
- **Non-root user**: Runs as `appuser`, not root
- **Read-only filesystem**: Container filesystem is read-only by default
- **No shell**: No shell or package manager in the final image
- **Static binary**: Single static binary with no dependencies
- **Resource limits**: CPU and memory limits configured in docker-compose.yml

### CI/CD Integration

#### GitHub Actions Example

```yaml
name: Install Agents
on:
  push:
    branches: [main]

jobs:
  install:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: docker build -t agent-manager:${{ github.sha }} .

      - name: Install agents
        run: |
          docker run --rm \
            -v ${{ github.workspace }}/agents-config.yaml:/app/agents-config.yaml:ro \
            -v agents:/app/.claude/agents \
            agent-manager:${{ github.sha }} install
```

#### GitLab CI Example

```yaml
install-agents:
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -t agent-manager:$CI_COMMIT_SHA .
    - docker run --rm
        -v $CI_PROJECT_DIR/agents-config.yaml:/app/agents-config.yaml:ro
        -v agents:/app/.claude/agents
        agent-manager:$CI_COMMIT_SHA install
```

### Troubleshooting Docker

**Permission denied errors**: Ensure volumes have correct permissions:

```bash
# Create volumes with proper permissions
docker volume create agent-data
docker volume create tracker-data
```

**Configuration not found**: Mount configuration file correctly:

```bash
# Ensure absolute path for config
docker run -v $(realpath agents-config.yaml):/app/agents-config.yaml:ro ...
```

**Network issues in container**: For private repositories, ensure authentication:

```bash
# Pass GitHub token
docker run -e GITHUB_TOKEN=$GITHUB_TOKEN ...

# For SSH authentication, mount SSH keys (not recommended for production)
docker run -v ~/.ssh:/home/appuser/.ssh:ro ...
```

**Debugging container issues**: Run with shell for debugging (requires debug image):

```bash
# Build debug image with shell
docker build --target builder -t agent-manager:debug .
docker run --rm -it agent-manager:debug /bin/sh
```

## Configuration Validation

### Validate Current Configuration

Check if your configuration file is valid:

```bash
agent-manager validate
```

### Validate Custom Configuration

Validate a different configuration file:

```bash
agent-manager validate --config my-config.yaml
```

### Verbose Validation

Show detailed validation information and warnings:

```bash
agent-manager validate --verbose
```

## Common Workflows

### Initial Setup

1. **Create your configuration**:

   ```bash
   # Start with the default configuration
   cp agents-config.yaml my-config.yaml
   # Edit my-config.yaml to suit your needs
   ```

2. **Validate your configuration**:

   ```bash
   agent-manager validate --config my-config.yaml
   ```

3. **Perform a dry run**:

   ```bash
   agent-manager install --config my-config.yaml --dry-run
   ```

4. **Install agents**:

   ```bash
   agent-manager install --config my-config.yaml
   ```

### Daily Operations

**Check what's installed**:

```bash
agent-manager list
```

**Update everything**:

```bash
agent-manager update
```

**Add a new source**:

1. Edit your configuration file
2. Validate: `agent-manager validate`
3. Install: `agent-manager install --source new-source-name`

### Troubleshooting

**Check for configuration issues**:

```bash
agent-manager validate --verbose
```

**See what would happen**:

```bash
agent-manager install --dry-run --verbose
```

**Reinstall problematic source**:

```bash
agent-manager uninstall --source problem-source
agent-manager install --source problem-source --verbose
```

## Working with Different Sources

### GitHub Repositories

Basic GitHub source:

```yaml
- name: my-github-agents
  type: github
  repository: owner/repo-name
  branch: main
```

With authentication:

```yaml
- name: private-agents
  type: github
  repository: company/private-agents
  auth:
    method: token
    token_env: GITHUB_TOKEN
```

### Git Repositories

HTTPS repository:

```yaml
- name: git-agents
  type: git
  url: https://gitlab.com/user/agents.git
  branch: develop
```

SSH repository:

```yaml
- name: ssh-agents
  type: git
  url: git@gitlab.com:user/agents.git
  auth:
    method: ssh
    ssh_key: ~/.ssh/id_rsa
```

### Local Sources

Local development:

```yaml
- name: local-agents
  type: local
  paths:
    source: ~/my-local-agents
    target: .claude/agents/local
  watch: true
```

### Marketplace Sources

Subagents.sh marketplace integration:

```yaml
- name: subagents-marketplace-all
  type: subagents
  paths:
    target: ${settings.base_dir}/marketplace
  cache:
    enabled: true
    ttl_hours: 1
    max_size_mb: 50
```

With category filtering:

```yaml
- name: development-agents
  type: subagents
  category: "Development"
  paths:
    target: .claude/agents/dev
  filters:
    include:
      regex: ["^(code-reviewer|documentation-expert).*\\.md$"]
```

## Conflict Resolution

When files already exist, Agent Manager uses conflict resolution strategies:

### Backup Strategy (Default)

Creates timestamped backups before overwriting:

```bash
# Files are backed up to .claude/backups/
original-file.md -> original-file.md.20240113-150405
```

### Override Strategy

Force overwrite existing files:

```yaml
settings:
  conflict_strategy: overwrite
```

### Skip Strategy

Keep existing files, skip new ones:

```yaml
sources:
  - name: my-source
    conflict_strategy: skip
```

### Merge Strategy

Attempt to merge content (currently same as backup):

```yaml
settings:
  conflict_strategy: merge
```

## File Filtering

Control which files are installed using filters:

### Include by Extension

```yaml
filters:
  include:
    extensions: [".md", ".yaml"]
```

### Include by Pattern

```yaml
filters:
  include:
    patterns: ["*-agent/*", "docs/*"]
```

### Include by Regex

```yaml
filters:
  include:
    regex: ["^(core|data).*\\.md$"]
```

### Exclude Patterns

```yaml
filters:
  exclude:
    patterns: ["test-*", "*.tmp", ".*"]
```

## Transformations

Transform files during installation:

### Remove Numeric Prefixes

```yaml
transformations:
  - type: remove_numeric_prefix
    pattern: "^[0-9]{2}-"
```

### Extract Documentation

```yaml
transformations:
  - type: extract_docs
    source_pattern: "*/README.md"
    target_dir: docs
    naming: UPPERCASE_UNDERSCORE
```

### Custom Scripts

```yaml
transformations:
  - type: custom_script
    script: ./transform-agents.sh
    args: ["--format", "claude"]
```

## Environment Variables

Customize behavior with environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `AGENT_MANAGER_CONFIG` | Default config file | agents-config.yaml |
| `AGENT_MANAGER_LOG_LEVEL` | Log level | info |
| `NO_COLOR` | Disable colored output | - |
| `GITHUB_TOKEN` | GitHub API token | - |

Example:

```bash
export AGENT_MANAGER_LOG_LEVEL=debug
export GITHUB_TOKEN=ghp_your_token_here
agent-manager install
```

## Exit Codes

Agent Manager uses standard exit codes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Installation error |
| 4 | Network error |

## Advanced Usage

### Batch Operations

Install multiple specific sources:

```bash
for source in core-agents data-agents; do
    agent-manager install --source $source
done
```

### Configuration Templates

Use environment variable substitution:

```yaml
sources:
  - name: ${COMPANY}-agents
    repository: ${GITHUB_ORG}/claude-agents
    auth:
      token_env: ${TOKEN_VAR}
```

### Monitoring Installations

Track installation progress:

```bash
# Terminal 1: Run installation
agent-manager install --verbose

# Terminal 2: Monitor tracking file
watch -n 1 cat .claude/.installed-agents.json
```

## Best Practices

1. **Always validate before installing**:

   ```bash
   agent-manager validate && agent-manager install
   ```

2. **Use dry-run for new configurations**:

   ```bash
   agent-manager install --dry-run --verbose
   ```

3. **Keep backups enabled** (default behavior)

4. **Use version control for your configuration files**

5. **Test with local sources first** when developing new agents

6. **Use meaningful source names** for easier management

7. **Document your custom transformations** and post-install scripts

8. **Regularly update your agents**:

   ```bash
   agent-manager update
   ```

9. **Monitor installation logs** for issues:

   ```bash
   tail -f .claude/installation.log
   ```

10. **Use source-specific conflict strategies** when needed:

    ```yaml
    sources:
      - name: experimental-agents
        conflict_strategy: skip  # Don't overwrite existing
    ```