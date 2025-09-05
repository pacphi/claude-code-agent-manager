# Install Command

Install Claude Code subagents from configured sources to your system.

## Basic Usage

```bash
agent-manager install [options]
```

## Installation Options

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

*Note: Force reinstall functionality is not implemented. The installer automatically detects changes and reinstalls when source content has been updated.*

## Conflict Resolution

When files already exist, Agent Manager uses conflict resolution strategies:

### Backup Strategy (Default)

Creates timestamped backups before overwriting:

```bash
# Files are backed up to .claude/backups/
original-file.md -> original-file.md.20240113-150405
```

### Override Strategy

Force overwrite existing files without backup:

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

## Source-Specific Installation

You can control which sources to install from:

```bash
# Install from single source
agent-manager install --source github-source

# Note: Multiple sources in a single command not supported
# Use separate commands for multiple sources
```

## Installation Tracking

All installations are tracked in `.claude/.agent-manager/state.json`:

- File paths and checksums
- Installation timestamps
- Source information
- Transformation history

## Error Handling

The installation process handles errors gracefully:

- **Network failures**: Automatic retry with backoff
- **Permission errors**: Clear error messages with resolution steps
- **Conflict issues**: Uses configured resolution strategy
- **Rollback**: Automatic rollback on critical failures

## Examples

### Example: Install with Custom Config

```bash
agent-manager install --config custom-config.yaml
```

*Note: Custom target directory option is not implemented. Installation directory is configured via the YAML configuration file.*

### Example: Install with Authentication

```bash
# Using environment variable
export GITHUB_TOKEN=your_token
agent-manager install --source private-repo

# Using config file
agent-manager install --config config-with-auth.yaml
```

## Related Commands

- [`list`](ADVANCED.md#list-command) - View installed agents
- [`update`](ADVANCED.md#update-command) - Update installed agents
- [`uninstall`](ADVANCED.md#uninstall-command) - Remove installed agents
- [`validate`](ADVANCED.md#validate-command) - Validate configuration

## See Also

- [Configuration Guide](../guides/CONFIGURATION.md)
- [Conflict Resolution Guide](../guides/CONFLICT-RESOLUTION.md)
- [Workflows](../guides/WORKFLOWS.md)
