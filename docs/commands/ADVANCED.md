# Advanced Commands

Additional Agent Manager commands for maintaining, managing, and querying your installations.

## Update Command

Update existing agent installations with the latest versions from their sources.

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

## Uninstall Command

Remove installed agents and optionally clean up related files.

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

### Keep Backups

Preserve backup files when uninstalling:

```bash
agent-manager uninstall --source my-agents --keep-backups
```

## List Command

View information about installed agents and their sources.

### List All Installed Agents

Show all installed agents:

```bash
agent-manager list
```

### List by Source

Show agents from a specific source:

```bash
agent-manager list --source awesome-claude-code-subagents
```

### Verbose Listing

Show detailed information about installations:

```bash
agent-manager list --verbose
```

*Note: Format options are not implemented in the list command. Use the query command with --output option for JSON/YAML formats.*

## Validate Command

Check configuration files for errors and validate settings.

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

*Note: Advanced validation options for agent files are not currently implemented. The validate command only validates the YAML configuration file.*

## Query Command

Search installed agents based on Claude Code subagent specification fields.

### Basic Search

Search across all agent fields:

```bash
# Search across name, description, and content
agent-manager query "code review"

# Search for TypeScript-related agents
agent-manager query "typescript"
```

### Field-Specific Search

Search within specific agent fields using field:value syntax:

```bash
# Search by agent name
agent-manager query "name:reviewer"

# Search in descriptions
agent-manager query "description:typescript"

# Search within agent prompts
agent-manager query "content:javascript"

# Find agents using specific tools
agent-manager query "tools:Read,Write"

# Or use the --field flag
agent-manager query --field name "reviewer"
```

### Advanced Search Options

```bash
# Use regex patterns
agent-manager query "code.*review" --regex

# Find agents without explicit tools (inherited)
agent-manager query --no-tools

# Find agents with explicit tool definitions
agent-manager query --custom-tools

# Filter by installation source
agent-manager query --source "awesome-agents"

# Output in different formats
agent-manager query "go" --output json
agent-manager query "go" --output yaml

# Set fuzzy matching threshold
agent-manager query "database management" --fuzzy-score 0.6

# Set query timeout
agent-manager query "complex search" --timeout 60s
```

*Note: Interactive mode and temporal filtering are not implemented.*

## Show Command

Display detailed information about specific agents with fuzzy matching support.

### Basic Usage

```bash
# Exact match
agent-manager show code-reviewer

# Fuzzy matching finds "code-reviewer.md"
agent-manager show reviewer
agent-manager show "code rev"
```

*Note: The show command has no additional options. It displays detailed information including name, description, file path, tools, and a prompt preview. Fuzzy matching is supported by default.*

## Stats Command

Get aggregate statistics and analytics about your installed agents.

### Basic Statistics

```bash
# Overall agent statistics
agent-manager stats

# Detailed statistics by source
agent-manager stats --detailed
```

### Analysis Options

```bash
# Show validation report
agent-manager stats --validation

# Show top tools usage
agent-manager stats --tools

# Limit number of tools shown
agent-manager stats --tools --tools-limit 20
```

## Index Command

Manage search index and cache for optimal query performance.

### Index Management

```bash
# Build or update the search index
agent-manager index build

# Force rebuild the search index
agent-manager index rebuild

# Show index statistics
agent-manager index stats
```

### Cache Management

```bash
# Clear query cache
agent-manager index cache-clear

# Show cache performance stats
agent-manager index cache-stats
```

## Version Command

Display version information and check for updates.

### Show Version

```bash
agent-manager version
```

### Detailed Version Info

```bash
agent-manager version --verbose
```

Output includes:

- Version number
- Build date
- Go version
- Git commit hash
- Platform information

## Global Options

These options work with all commands:

### Configuration File

Specify an alternative configuration file:

```bash
agent-manager [command] --config /path/to/config.yaml
```

### Verbose Output

Enable detailed logging:

```bash
agent-manager [command] --verbose
```

### Dry Run

Preview actions without making changes:

```bash
agent-manager [command] --dry-run
```

### No Color

Disable colored output:

```bash
agent-manager [command] --no-color
```

### No Progress

Disable progress indicators:

```bash
agent-manager [command] --no-progress
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
| 5 | Permission error |

## Examples

### Update Workflow

```bash
# Check for updates
agent-manager update --check-only

# If updates available, review and apply
agent-manager update --verbose

# Verify updates
agent-manager list
```

### Maintenance Workflow

```bash
# Validate configuration
agent-manager validate

# List current state
agent-manager list --verbose

# Clean old backups
agent-manager uninstall --source old-source --keep-backups=false
```

### Debugging Issues

```bash
# Verbose validation
agent-manager validate --verbose

# Dry run with verbose output
agent-manager install --dry-run --verbose

# Check specific source
agent-manager list --source problematic-source --verbose
```

## Related Commands

- [`install`](INSTALL.md) - Install agents
- [`marketplace`](MARKETPLACE.md) - Browse marketplace

**Note**: The `query`, `show`, `stats`, `index`, and enhanced `validate` commands work with installed agents to help you discover, analyze, and manage your Claude Code subagents.

## See Also

- [Configuration Guide](../guides/CONFIGURATION.md)
- [Troubleshooting](../guides/TROUBLESHOOTING.md)
- [CLI Reference](../reference/CLI-REFERENCE.md)
