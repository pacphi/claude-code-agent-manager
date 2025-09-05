# Troubleshooting Guide

Solutions for common issues and debugging techniques for Agent Manager.

## Common Issues

### Installation Problems

#### "Command not found"

```bash
# Check if binary exists
ls -la bin/agent-manager

# If not, rebuild
make clean && make build

# For system-wide access
sudo make install
```

#### "Permission denied"

```bash
# Make binary executable
chmod +x bin/agent-manager

# For system directories
sudo agent-manager install

# Fix directory permissions
chmod 755 ~/.claude
chmod 755 ~/.claude/agents
```

#### Build Failures

```bash
# Clean build artifacts
make clean

# Update dependencies
make deps

# Verbose build
go build -v -o bin/agent-manager cmd/agent-manager/main.go
```

### Configuration Issues

#### "Configuration file not found"

```bash
# Check file exists
ls agents-config.yaml

# Use explicit path
agent-manager install --config /path/to/config.yaml

# Create from example
cp agents-config.yaml.example agents-config.yaml
```

#### "Invalid YAML"

```bash
# Validate syntax
agent-manager validate --verbose

# Common YAML issues:
# - Use spaces, not tabs
# - Consistent indentation (2 or 4 spaces)
# - Quote special characters
# - Escape $ in strings or use single quotes
```

#### Variable Substitution Errors

```yaml
# Wrong
sources:
  - repository: ${GITHUB_ORG}/agents  # Missing env prefix

# Correct
sources:
  - repository: ${env.GITHUB_ORG}/agents
  - target: ${settings.base_dir}/agents
```

### Source Access Issues

#### GitHub Authentication Failed

```bash
# Check token is set
echo $GITHUB_TOKEN

# Test token
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/user

# Set token
export GITHUB_TOKEN=your_personal_access_token

# Or use gh CLI
gh auth login
```

#### "Repository not found"

```bash
# Verify repository exists
gh repo view owner/repo-name

# Check access permissions
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/owner/repo-name
```

#### SSH Key Issues

```bash
# Test SSH connection
ssh -T git@github.com

# Add SSH key to agent
ssh-add ~/.ssh/id_rsa

# Check SSH config
cat ~/.ssh/config
```

### Marketplace Issues

#### "Chrome/Chromium not found"

```bash
# Check browser installation
which chrome || which chromium || which brave

# Install browser (macOS)
brew install --cask google-chrome

# Install browser (Linux)
sudo apt-get install chromium-browser

# Set custom path in config
```

```yaml
marketplace:
  browser_path: /Applications/Google Chrome.app/Contents/MacOS/Google Chrome
```

#### "Context deadline exceeded"

```bash
# Increase timeout
agent-manager marketplace list --timeout 60

# Use headless mode
agent-manager marketplace list --headless

# Check network
ping subagents.sh
```

#### "No agents found"

```bash
# Refresh cache
agent-manager marketplace refresh

# Enable verbose logging
agent-manager marketplace list --verbose

# Check browser console for errors
# May need to update browser or clear cache
```

### Conflict Resolution Issues

#### Unexpected File Overwrites

```bash
# Check current strategy
grep conflict_strategy agents-config.yaml

# Configure backup strategy (default) in agents-config.yaml
agent-manager install

# Preview changes
agent-manager install --dry-run
```

#### Lost Custom Changes

```bash
# Check backups
ls -la ~/.claude/backups/

# Restore from backup
cp ~/.claude/backups/file.md.timestamp ~/.claude/agents/file.md

# Configure skip strategy to preserve changes
# Set in agents-config.yaml: conflict_strategy: skip
agent-manager update
```

### Query and Index Issues

#### "No results found" for known agents

```bash
# Test if the query system is working
agent-manager validate --query

# Rebuild the search index
agent-manager index rebuild

# Check index statistics
agent-manager index stats

# Verify agents are actually installed
agent-manager list
```

#### Slow query performance

```bash
# Check index statistics
agent-manager index stats

# Build index if not present
agent-manager index build

# Clear cache if stale
agent-manager index cache-clear
```

#### "Agent not found" with show command

```bash
# Try fuzzy search
agent-manager query "partial-name"

# List all agents to see exact names
agent-manager list

# Show specific agent (fuzzy matching is built-in)
agent-manager show exact-name
```

### State and Tracking Issues

#### "Inconsistent state"

```bash
# Check state file
cat ~/.agent-manager/state.json | jq .

# Reset state
rm ~/.agent-manager/state.json
agent-manager list  # Rebuilds state

# Full reset
agent-manager uninstall --all
rm -rf ~/.agent-manager
agent-manager install
```

#### "File not tracked"

```bash
# Rebuild tracking
agent-manager index rebuild

# Manual state verification
find ~/.claude/agents -type f | wc -l
jq '.installations | length' ~/.agent-manager/state.json
```

## Debugging Techniques

### Enable Verbose Logging

```bash
# Verbose output for any command
agent-manager install --verbose
agent-manager validate --verbose
agent-manager list --verbose

# Debug environment
export DEBUG=true
agent-manager install
```

### Dry Run Mode

```bash
# Preview without changes
agent-manager install --dry-run
agent-manager update --dry-run
agent-manager uninstall --source test --dry-run
```

### Check Specific Components

```bash
# Validate configuration only
agent-manager validate

# Test specific source
agent-manager install --source single-source --dry-run

# Check network connectivity
curl -I https://github.com
curl -I https://subagents.sh
```

### Log Analysis

```bash
# Capture full output
agent-manager install --verbose 2>&1 | tee install.log

# Search for errors
grep -i error install.log
grep -i "failed\|fail" install.log

# Check recent operations
ls -lt ~/.agent-manager/logs/
```

## Platform-Specific Issues

### macOS

#### "xcrun: error"

```bash
# Install Xcode Command Line Tools
xcode-select --install

# Reset if corrupted
sudo xcode-select --reset
```

#### Quarantine Issues

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine bin/agent-manager

# Allow in System Preferences > Security & Privacy
```

### Linux

#### Missing Dependencies

```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install build-essential git

# RHEL/CentOS
sudo yum groupinstall "Development Tools"
sudo yum install git
```

#### SELinux Issues

```bash
# Check SELinux status
sestatus

# Temporary disable (testing only)
sudo setenforce 0

# Add exception for Agent Manager
sudo chcon -t bin_t bin/agent-manager
```

### Windows (WSL2)

#### Line Ending Issues

```bash
# Configure Git for Unix endings
git config --global core.autocrlf false

# Convert existing files
dos2unix agents-config.yaml
```

#### Path Issues

```bash
# Use Unix-style paths
agent-manager install --config /mnt/c/Users/name/config.yaml

# Not Windows paths
# agent-manager install --config C:\Users\name\config.yaml
```

## Performance Issues

### Slow Installation

```bash
# Check network speed
speedtest-cli

# Use specific sources
agent-manager install --source fast-source

# Increase parallelism
agent-manager install --parallel 4
```

### High Memory Usage

```bash
# Monitor memory
top -p $(pgrep agent-manager)

# Limit concurrent operations
agent-manager install --max-concurrent 1

# Clear cache
rm -rf ~/.agent-manager/cache
```

### Disk Space Issues

```bash
# Check available space
df -h ~/.claude

# Clean old backups
find ~/.claude/backups -mtime +30 -delete

# Remove cache
rm -rf ~/.agent-manager/cache
```

## Recovery Procedures

### Full Reset

```bash
# Backup current state
tar -czf agent-backup.tar.gz ~/.claude ~/.agent-manager

# Complete removal
agent-manager uninstall --all
rm -rf ~/.claude/agents
rm -rf ~/.agent-manager

# Fresh start
agent-manager install
```

### Restore from Backup

```bash
# Restore from tarball
tar -xzf agent-backup.tar.gz -C ~/

# Verify restoration
agent-manager list
agent-manager validate
```

### Partial Recovery

```bash
# Fix specific source
agent-manager uninstall --source problematic-source
agent-manager install --source problematic-source --verbose

# Repair permissions
find ~/.claude -type d -exec chmod 755 {} \;
find ~/.claude -type f -exec chmod 644 {} \;
```

## Getting Help

### Self-Diagnosis

```bash
# Run built-in diagnostics
agent-manager diagnose

# System information
agent-manager version --verbose
go version
git --version
```

### Collecting Debug Information

When reporting issues, include:

```bash
# Version information
agent-manager version --verbose > debug.txt

# Configuration (sanitized)
cat agents-config.yaml | sed 's/token.*/token: REDACTED/' >> debug.txt

# Recent operations
agent-manager list --verbose >> debug.txt

# System info
uname -a >> debug.txt
go version >> debug.txt
```

### Support Channels

1. **Documentation**: Check all guides in `/docs`
2. **GitHub Issues**: [Report bugs](https://github.com/pacphi/claude-code-agent-manager/issues)
3. **Discussions**: Community help and Q&A
4. **Wiki**: Additional tips and tricks

## See Also

- [Installation Guide](../getting-started/INSTALLATION.md)
- [Configuration Guide](CONFIGURATION.md)
- [Conflict Resolution](CONFLICT-RESOLUTION.md)
- [Command Reference](../reference/CLI-REFERENCE.md)
