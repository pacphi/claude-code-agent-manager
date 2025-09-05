# First Steps with Agent Manager

A hands-on tutorial to get you started with Agent Manager.

## What You'll Learn

- Installing your first agents
- Understanding sources and transformations
- Managing conflicts
- Updating and maintaining agents

## Before You Begin

Ensure you've completed:

1. [Prerequisites](PREREQUISITES.md) - System requirements
2. [Installation](INSTALLATION.md) - Agent Manager setup

Verify everything is ready:

```bash
# Check Agent Manager is installed
agent-manager version

# Validate configuration
agent-manager validate
```

## Tutorial: Your First Agent Installation

### Step 1: Understand the Configuration

View your configuration:

```bash
cat agents-config.yaml
```

Key sections:

- **settings**: Global configuration
- **sources**: Where agents come from
- **transformations**: How files are processed

### Step 2: Preview What Will Be Installed

Use dry-run to see what would happen:

```bash
agent-manager install --dry-run
```

This shows:

- Which files would be installed
- Where they would go
- Any conflicts that would occur

### Step 3: Install Agents

Now install for real:

```bash
agent-manager install
```

Watch the output:

```
Installing from source: awesome-claude-code-subagents
→ Cloning repository...
→ Applying transformations...
→ Installing files...
✓ Installed 25 agents to .claude/agents
```

### Step 4: Verify Installation

Check what was installed:

```bash
# List all installed agents
agent-manager list

# Search for specific agents
agent-manager query "code review"

# Get detailed information about an agent
agent-manager show code-reviewer

# View statistics about your installation
agent-manager stats

# Check the actual files
ls -la ~/.claude/agents/
```

## Understanding Key Concepts

### Sources

Sources define where agents come from:

```yaml
sources:
  - name: awesome-claude-code-subagents
    type: github  # GitHub repository
    repository: VoltAgent/awesome-claude-code-subagents
```

Types:

- **github**: GitHub repositories
- **git**: Any Git repository
- **local**: Local filesystem
- **subagents**: Marketplace integration

### Transformations

Transformations modify files during installation:

```yaml
transformations:
  - type: remove_numeric_prefix  # Remove "01-" prefixes
  - type: extract_docs           # Move docs to separate folder
```

### Conflict Resolution

When files already exist:

```yaml
conflict_strategy: backup  # Default: create backups
# Options: backup, overwrite, skip, merge
```

## Discovering and Querying Agents

### Search Your Installed Agents

Once you have agents installed, you can search and analyze them:

```bash
# Find agents related to code review
agent-manager query "code review"

# Search for agents that work with TypeScript
agent-manager query --description "typescript"

# Find agents using specific tools
agent-manager query --tools "Read,Write"

# Get detailed info about a specific agent
agent-manager show code-reviewer

# See what tools an agent uses
agent-manager show code-reviewer --extract-tools
```

### Analyze Your Installation

```bash
# Get overall statistics
agent-manager stats

# Show validation report
agent-manager stats --validation

# Show detailed statistics by source
agent-manager stats --detailed

# Show tool usage across agents
agent-manager stats --tools
```

### Validate Your Agents

```bash
# Validate all agents against Claude Code spec
agent-manager validate --agents

# Check for common issues
agent-manager validate --agents --check-names --check-duplicates
```

## Common Tasks

### Install from Specific Source

```bash
# Install only from one source
agent-manager install --source awesome-claude-code-subagents
```

### Update Existing Agents

```bash
# Check for updates
agent-manager update --check-only

# Apply updates
agent-manager update
```

### Remove Agents

```bash
# Remove specific source
agent-manager uninstall --source old-agents

# Remove all
agent-manager uninstall --all
```

### Handle Conflicts

When you see a conflict, configure the strategy in agents-config.yaml:

```yaml
settings:
  conflict_strategy: skip     # Or: overwrite, backup (default)
```

Then install:

```bash
agent-manager install
```

## Working with Multiple Sources

Add a new source to your config:

```yaml
sources:
  - name: my-custom-agents
    type: github
    repository: myusername/my-agents
    enabled: true
```

Then install:

```bash
# Install all enabled sources
agent-manager install

# Or just the new one
agent-manager install --source my-custom-agents
```

## Marketplace Discovery

Browse available agents:

```bash
# List all categories
agent-manager marketplace list

# Browse specific category
agent-manager marketplace list --category "Development"

# View agent details
agent-manager marketplace show "code-reviewer"
```

## Best Practices

### 1. Always Validate First

```bash
agent-manager validate --verbose
```

### 2. Use Dry Run for Testing

```bash
agent-manager install --dry-run --verbose
```

### 3. Keep Backups

```bash
# Default backup strategy is configured in agents-config.yaml
agent-manager install
```

### 4. Track Your Changes

```bash
# See what's installed
agent-manager list --verbose

# Check installation state
cat ~/.agent-manager/state.json
```

## Exercise: Complete Workflow

Try this complete workflow:

```bash
# 1. Validate configuration
agent-manager validate

# 2. Preview installation
agent-manager install --dry-run

# 3. Install agents
agent-manager install

# 4. Verify installation
agent-manager list

# 5. Check for updates (later)
agent-manager update --check-only

# 6. Apply updates
agent-manager update
```

## Troubleshooting Common Issues

### "Configuration not found"

```bash
# Ensure config exists
ls agents-config.yaml

# Or specify path
agent-manager install --config /path/to/config.yaml
```

### "Permission denied"

```bash
# Check file permissions
chmod +x bin/agent-manager

# For system directories, use sudo
sudo agent-manager install
```

### "Source not found"

```bash
# List available sources
grep "name:" agents-config.yaml

# Check source is enabled
grep -A2 "name: source-name" agents-config.yaml
```

## What's Next?

You've learned the basics! Now explore:

1. [Configuration Guide](../guides/CONFIGURATION.md) - Advanced configuration
2. [Workflows](../guides/WORKFLOWS.md) - Complex usage patterns
3. [Marketplace](../commands/MARKETPLACE.md) - Discover more agents
4. [Docker Usage](../development/DOCKER.md) - Container deployment

## Getting Help

- Command help: `agent-manager [command] --help`
- [Troubleshooting Guide](../guides/TROUBLESHOOTING.md)
- [GitHub Issues](https://github.com/pacphi/claude-code-agent-manager/issues)
