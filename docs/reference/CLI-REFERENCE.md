# CLI Reference

Complete command-line interface reference for Agent Manager.

## Synopsis

```bash
agent-manager [global-options] <command> [command-options]
```

## Global Options

These options work with all commands:

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--config` | `-c` | Configuration file path | `agents-config.yaml` |
| `--verbose` | `-v` | Enable verbose output | `false` |
| `--dry-run` | | Preview changes without applying | `false` |
| `--no-color` | | Disable colored output | `false` |
| `--no-progress` | | Disable progress indicators | `false` |
| `--help` | `-h` | Show help for command | |

## Commands

### install

Install agents from configured sources.

```bash
agent-manager install [options]
```

**Options:**

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--source` | `-s` | Install from specific source only | All enabled |

*Note: Advanced options like --force, --conflict-strategy, --parallel, and --timeout are configured via the YAML configuration file rather than command-line flags.*

**Examples:**

```bash
# Install all sources
agent-manager install

# Install specific source
agent-manager install --source github-agents

# Force reinstall with overwrite
agent-manager install --force --conflict-strategy overwrite
```

### uninstall

Remove installed agents.

```bash
agent-manager uninstall [options]
```

**Options:**

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--source` | `-s` | Uninstall specific source | Required unless --all |
| `--all` | `-a` | Uninstall all sources | `false` |
| `--keep-backups` | | Preserve backup files | `false` |

**Examples:**

```bash
# Uninstall specific source
agent-manager uninstall --source old-agents

# Uninstall everything
agent-manager uninstall --all
```

### update

Update installed agents to latest versions.

```bash
agent-manager update [options]
```

**Options:**

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--source` | `-s` | Update specific source | All installed |
| `--check-only` | | Check for updates without applying | `false` |

**Examples:**

```bash
# Update all sources
agent-manager update

# Check for updates only
agent-manager update --check-only

# Update specific source
agent-manager update --source github-agents
```

### list

List installed agents and sources.

```bash
agent-manager list [options]
```

**Options:**

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--source` | `-s` | List specific source only | All sources |
| `--search` | | Search in agent names, descriptions, or content | |
| `--name` | | Search by agent name | |
| `--description` | | Search in description | |
| `--tools` | | Filter by tools (comma-separated) | |
| `--no-tools` | | Show agents with inherited tools only | `false` |
| `--custom-tools` | | Show agents with explicit tools only | `false` |
| `--limit` | | Limit number of results | `50` |

*Note: Format options (json/tree/csv) are not implemented in the list command. Use the query command with --output option for different formats.*

**Examples:**

```bash
# List all agents
agent-manager list

# Detailed listing of specific source
agent-manager list --source github-agents --verbose
```

### marketplace

Browse and discover agents from subagents.sh marketplace.

```bash
agent-manager marketplace <subcommand> [options]
```

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `list` | List categories or agents |
| `show` | Show agent details |
| `refresh` | Update marketplace cache |

**List Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `--category` | Filter by category | All categories |
| `--limit` | Maximum results | `50` |

**Show Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `--content` | Include full agent content | `false` |

**Examples:**

```bash
# List all categories
agent-manager marketplace list

# List agents in category
agent-manager marketplace list --category "Development"

# Show agent details
agent-manager marketplace show "code-reviewer"
```

### query

Search installed agents with complex queries, regex patterns, and fuzzy matching.

```bash
agent-manager query [QUERY] [options]
```

**Options:**

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--field` | `-f` | Search specific field (name, description, content, tools, source) | |
| `--limit` | `-l` | Limit number of results | unlimited |
| `--no-tools` | | Find agents with inherited tools only | `false` |
| `--custom-tools` | | Find agents with explicit tools only | `false` |
| `--source` | `-s` | Filter by source | |
| `--output` | `-o` | Output format (table, json, yaml) | `table` |
| `--regex` | | Use regex pattern matching | `false` |
| `--fuzzy-score` | | Fuzzy matching threshold (0.0-1.0) | `0.7` |
| `--timeout` | | Query timeout | `30s` |

**Examples:**

```bash
# Basic field searches
agent-manager query "name:go"
agent-manager query "tools:bash,git"
agent-manager query "description:automation"

# Regex pattern matching
agent-manager query "name:^data.*processor$" --regex

# Multi-field fuzzy search
agent-manager query "database management" --fuzzy-score 0.6

# Output formats
agent-manager query "go" --output json
agent-manager query "go" --output yaml
```

### show

Display detailed information about specific agents.

```bash
agent-manager show <agent-name> [options]
```

*Note: The show command has no additional options. It displays detailed information including name, description, file path, tools, and a prompt preview. Fuzzy matching is supported by default.*

**Examples:**

```bash
# Show agent details (fuzzy match supported)
agent-manager show code-reviewer
agent-manager show reviewer  # Finds "code-reviewer.md"
```

### stats

Aggregate statistics about installed agents.

```bash
agent-manager stats [options]
```

**Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `--detailed` | Show detailed statistics by source | `false` |
| `--validation` | Show validation report | `false` |
| `--tools` | Show top tools usage | `false` |
| `--tools-limit` | Limit number of tools shown | `10` |

**Examples:**

```bash
# Basic statistics
agent-manager stats

# Detailed analysis
agent-manager stats --coverage --duplicates
agent-manager stats --tools --by-source
```

### index

Manage search index and cache with subcommands.

```bash
agent-manager index <subcommand>
```

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `build` | Build/update index |
| `rebuild` | Force rebuild index |
| `stats` | Show index statistics |
| `cache-clear` | Clear query cache |
| `cache-stats` | Show cache statistics |

**Examples:**

```bash
# Index management
agent-manager index build
agent-manager index rebuild
agent-manager index stats
agent-manager index cache-clear
```

### validate

Validate configuration file syntax and semantics, plus agent-specific validation.

```bash
agent-manager validate [options]
```

**Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `--strict` | Fail on warnings | `false` |
| `--verbose` | Show detailed validation | `false` |
| `--agents` | Validate all agent files | `false` |
| `--agent` | Validate specific agent | |
| `--check-names` | Verify name format (lowercase, hyphens) | `false` |
| `--check-tools` | Verify tool declarations exist | `false` |
| `--check-duplicates` | Find duplicate agent names | `false` |
| `--check-required` | Ensure name & description present | `false` |

**Examples:**

```bash
# Basic validation
agent-manager validate

# Agent-specific validation
agent-manager validate --agents
agent-manager validate --agent code-reviewer
agent-manager validate --agents --check-names --check-tools
```

### version

Display version information.

```bash
agent-manager version [options]
```

**Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `--verbose` | Show detailed build info | `false` |
| `--check-update` | Check for newer versions | `false` |

**Output includes:**

- Version number
- Build date
- Go version
- Git commit
- Platform/Architecture

### help

Display help information.

```bash
agent-manager help [command]
```

**Examples:**

```bash
# General help
agent-manager help

# Command-specific help
agent-manager help install
```

## Configuration File

Agent Manager uses a YAML configuration file (default: `agents-config.yaml`).

### Structure

```yaml
settings:
  base_dir: .claude/agents
  conflict_strategy: backup
  timeout_seconds: 300

sources:
  - name: source-name
    type: github|git|local|subagents
    enabled: true
    # ... source-specific options
```

### Override with CLI

Command-line options override configuration:

```bash
# Override config file
agent-manager install --config production.yaml

# Override conflict strategy
agent-manager install --conflict-strategy overwrite

# Override timeout
agent-manager install --timeout 600
```

## Environment Variables

Agent Manager supports environment variables:

| Variable | Description | Usage |
|----------|-------------|-------|
| `AGENT_MANAGER_CONFIG` | Default config file | Alternative to --config |
| `AGENT_MANAGER_HOME` | Base directory | Overrides settings.base_dir |
| `GITHUB_TOKEN` | GitHub authentication | For private repos |
| `GITLAB_TOKEN` | GitLab authentication | For private repos |
| `NO_COLOR` | Disable colors | Set to any value |
| `DEBUG` | Debug mode | Set to "true" for verbose |

## Exit Codes

| Code | Meaning | Common Causes |
|------|---------|---------------|
| 0 | Success | Operation completed |
| 1 | General error | Unexpected failure |
| 2 | Configuration error | Invalid YAML, missing file |
| 3 | Installation error | Permission denied, conflicts |
| 4 | Network error | Connection failed, timeout |
| 5 | Authentication error | Invalid token, access denied |
| 127 | Command not found | Binary not in PATH |

## Output Formats

### Text (Default)

Human-readable output with colors and formatting.

### JSON

Machine-readable JSON for scripting:

```bash
agent-manager list --format json | jq '.sources[].name'
```

### Tree

Hierarchical view of installations:

```bash
agent-manager list --format tree
```

### CSV

Comma-separated values for spreadsheets:

```bash
agent-manager list --format csv > agents.csv
```

## Scripting Examples

### Check if source is installed

```bash
if agent-manager list --source my-source --format json | jq -e '.sources | length > 0' > /dev/null; then
  echo "Source is installed"
fi
```

### Update only if changes available

```bash
if agent-manager update --check-only | grep -q "Updates available"; then
  agent-manager update
fi
```

### Backup before update

```bash
agent-manager list --format json > pre-update.json
agent-manager update
agent-manager list --format json > post-update.json
diff pre-update.json post-update.json
```

## Advanced Usage

### Parallel Operations

```bash
# Increase parallelism for faster operations
agent-manager install --parallel 4
```

### Custom Timeout

```bash
# Increase timeout for slow networks
agent-manager install --timeout 600
```

### Debug Mode

```bash
# Maximum verbosity
DEBUG=true agent-manager install --verbose
```

### Dry Run Everything

```bash
# Test complex operations
agent-manager update --dry-run --verbose
```

## See Also

- [Configuration Guide](../guides/CONFIGURATION.md)
- [Command Documentation](../commands/)
- [Docker Deployment](../development/DOCKER.md)
- [Troubleshooting](../guides/TROUBLESHOOTING.md)
