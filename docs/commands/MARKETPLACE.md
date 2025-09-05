# Marketplace Command

Browse and discover Claude Code subagents from the [subagents.sh](https://subagents.sh) marketplace.

## Prerequisites

**Browser Requirement**: Chrome, Chromium, or a Chromium-based browser must be installed and accessible in your PATH.

### Supported Browsers

- Google Chrome
- Chromium
- Chrome Canary
- Microsoft Edge (Chromium-based)
- Brave Browser

### Installation

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

## Commands

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

## Installing Marketplace Agents

Once you find an agent you want to install, add it to your `agents-config.yaml`:

### Install All Marketplace Agents

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

### Install Specific Agent

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

## Cache Management

The marketplace integration caches data to improve performance:

- **TTL (Time To Live)**: Configure how long to cache data
- **Max Size**: Limit cache size to prevent excessive disk usage
- **Manual Refresh**: Force update with `marketplace refresh`

```yaml
cache:
  enabled: true
  ttl_hours: 1      # Cache for 1 hour
  max_size_mb: 50   # Maximum 50MB cache
```

## Troubleshooting

### Common Issues

**"executable file not found in $PATH"**
Chrome/Chromium is not installed or not in your PATH. Install a supported browser using the installation options above.

**"context deadline exceeded"**
Indicates slow network connection or browser startup issues. Try increasing the timeout or checking your network connection.

**"No categories found" or "No agents found"**
May indicate changes to the subagents.sh website structure or JavaScript execution issues. Check logs and consider refreshing the cache.

### Debugging

Enable verbose logging for detailed browser automation logs:

```bash
agent-manager marketplace list --verbose
```

Check browser availability:

```bash
which chrome || which chromium || which brave
```

## Advanced Usage

### Custom Browser Path

If your browser is not in PATH, specify it in the configuration:

```yaml
marketplace:
  browser_path: /Applications/Google Chrome.app/Contents/MacOS/Google Chrome
```

### Headless Mode

Run browser in headless mode (no UI):

```yaml
marketplace:
  headless: true
```

### Timeout Configuration

Adjust timeouts for slower connections:

```yaml
marketplace:
  timeout_seconds: 30
```

## Related Commands

- [`install`](INSTALL.md) - Install agents from sources
- [`list`](ADVANCED.md#list-command) - List installed agents
- [`validate`](ADVANCED.md#validate-command) - Validate configuration

## See Also

- [Configuration Guide](../guides/CONFIGURATION.md)
- [Workflows](../guides/WORKFLOWS.md)

