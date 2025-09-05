# Conflict Resolution Guide

Understanding and managing file conflicts during agent installation.

## What Are Conflicts?

Conflicts occur when Agent Manager tries to install a file that already exists in the target location. This commonly happens when:

- Updating existing agents
- Installing from multiple sources with overlapping files
- Reinstalling after manual modifications
- Switching between different agent versions

## Conflict Strategies

Agent Manager provides four strategies for handling conflicts:

### 1. Backup Strategy (Default)

Creates timestamped backups before overwriting files.

```yaml
settings:
  conflict_strategy: backup
```

Behavior:
- Original file is moved to `.claude/backups/`
- Backup filename includes timestamp
- New file is installed in place
- Safe for preserving custom modifications

Example:
```bash
# Original file
.claude/agents/code-reviewer.md

# Backup created
.claude/backups/code-reviewer.md.20240113-150405

# New file installed
.claude/agents/code-reviewer.md
```

### 2. Overwrite Strategy

Replaces existing files without creating backups.

```yaml
settings:
  conflict_strategy: overwrite
```

Behavior:
- Existing file is deleted
- New file is installed
- No backup created
- Fastest but destructive

Use when:
- You don't need to preserve local changes
- Files are read-only references
- Speed is more important than safety

### 3. Skip Strategy

Keeps existing files and skips installation of conflicting files.

```yaml
settings:
  conflict_strategy: skip
```

Behavior:
- Existing file is preserved
- New file is not installed
- Installation continues with other files
- Safe for preserving local customizations

Use when:
- Local modifications should be preserved
- Updating only non-conflicting files
- Testing new sources without affecting existing setup

### 4. Merge Strategy

Attempts to merge content (currently same as backup).

```yaml
settings:
  conflict_strategy: merge
```

Behavior:
- Currently implements backup strategy
- Future: intelligent content merging
- Preserves both versions for manual merge

## Configuration Options

### Global Strategy

Set default strategy for all sources:

```yaml
settings:
  conflict_strategy: backup  # Applied to all sources
```

### Per-Source Strategy

Override global strategy for specific sources:

```yaml
sources:
  - name: stable-agents
    type: github
    repository: company/stable-agents
    conflict_strategy: skip  # Keep existing files

  - name: experimental-agents
    type: github
    repository: company/experimental-agents
    conflict_strategy: overwrite  # Always use latest
```

### Command-Line Override

Override configuration temporarily:

```bash
# Use specific strategy for this run
agent-manager install --conflict-strategy overwrite

# Applies to all sources in this execution
agent-manager update --conflict-strategy skip
```

## Conflict Detection

### Preview Conflicts

Use dry-run to see potential conflicts:

```bash
agent-manager install --dry-run
```

Output shows:
```
Would install: code-reviewer.md
  Conflict: File exists at .claude/agents/code-reviewer.md
  Resolution: Would create backup at .claude/backups/code-reviewer.md.20240113-150405
```

### List Existing Files

Check what's already installed:

```bash
# List all installed files
agent-manager list --verbose

# Check specific directory
ls -la ~/.claude/agents/
```

## Handling Specific Scenarios

### Scenario 1: Updating Modified Agents

You've customized an agent and want to update without losing changes:

```bash
# Option 1: Skip strategy preserves your version
agent-manager update --conflict-strategy skip

# Option 2: Backup strategy keeps both versions
agent-manager update --conflict-strategy backup
# Then manually merge changes from backup
```

### Scenario 2: Fresh Installation Over Existing

Replace all existing agents with fresh versions:

```bash
# Remove all existing agents
agent-manager uninstall --all

# Fresh install
agent-manager install
```

Or use overwrite:

```bash
agent-manager install --conflict-strategy overwrite
```

### Scenario 3: Testing New Sources

Try new sources without affecting existing setup:

```bash
# Use skip to preserve existing files
agent-manager install --source new-source --conflict-strategy skip

# Review what was installed
agent-manager list --source new-source
```

### Scenario 4: Recovering from Bad Update

Restore from backups:

```bash
# List available backups
ls -la ~/.claude/backups/

# Restore specific file
cp ~/.claude/backups/agent.md.20240113-150405 ~/.claude/agents/agent.md

# Or restore all from a specific time
for backup in ~/.claude/backups/*.20240113-*; do
  original=$(basename "$backup" | sed 's/\.[0-9-]*$//')
  cp "$backup" "~/.claude/agents/$original"
done
```

## Backup Management

### Backup Location

Default backup directory:
```bash
~/.claude/backups/
```

Configure custom location:
```yaml
settings:
  backup_dir: /path/to/backups
```

### Backup Naming

Format: `original-filename.YYYYMMDD-HHMMSS`

Example: `code-reviewer.md.20240113-150405`

### Cleaning Old Backups

Remove backups older than 30 days:

```bash
find ~/.claude/backups -type f -mtime +30 -delete
```

Keep only latest 5 backups per file:

```bash
for file in $(ls ~/.claude/backups | cut -d. -f1-2 | sort -u); do
  ls -t ~/.claude/backups/${file}.* | tail -n +6 | xargs rm -f
done
```

## Advanced Patterns

### Conditional Strategy

Use environment variables for dynamic strategy:

```yaml
settings:
  conflict_strategy: ${env.CONFLICT_STRATEGY:-backup}
```

Usage:
```bash
# Development: always use latest
CONFLICT_STRATEGY=overwrite agent-manager install

# Production: preserve existing
CONFLICT_STRATEGY=skip agent-manager install
```

### Source Priority

Install in order of importance:

```bash
# Critical agents - skip conflicts
agent-manager install --source critical-agents --conflict-strategy skip

# Optional agents - overwrite if needed
agent-manager install --source optional-agents --conflict-strategy overwrite
```

### Selective Restoration

Restore only specific agents from backup:

```bash
# Find all backups for specific agent
ls ~/.claude/backups/code-reviewer.md.*

# Restore latest backup
latest=$(ls -t ~/.claude/backups/code-reviewer.md.* | head -1)
cp "$latest" ~/.claude/agents/code-reviewer.md
```

## Best Practices

1. **Use Backup by Default**: Safest for most scenarios
2. **Preview with Dry Run**: Always check conflicts before installing
3. **Document Custom Changes**: Keep notes on modified agents
4. **Regular Backup Cleanup**: Prevent backup directory from growing too large
5. **Test Strategy Changes**: Use dry-run when changing strategies
6. **Version Control Config**: Track configuration changes in Git

## Troubleshooting

### "Permission Denied" During Backup

```bash
# Check backup directory permissions
ls -la ~/.claude/

# Fix permissions
chmod 755 ~/.claude/backups
```

### Backup Directory Full

```bash
# Check disk space
df -h ~/.claude/backups

# Clean old backups
find ~/.claude/backups -type f -mtime +7 -delete
```

### Wrong Strategy Applied

```bash
# Verify current configuration
agent-manager validate --verbose

# Check for environment overrides
echo $CONFLICT_STRATEGY
```

## See Also

- [Installation Guide](../getting-started/INSTALLATION.md)
- [Configuration Guide](CONFIGURATION.md)
- [Install Command](../commands/INSTALL.md)