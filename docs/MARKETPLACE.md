# Marketplace Integration

This document describes the subagents.sh marketplace integration for the Claude Code Agent Manager.

## Overview

The marketplace integration allows users to browse, search, and install agents directly from the [subagents.sh](https://subagents.sh) marketplace. The integration uses browser automation to scrape the marketplace since it's built with modern JavaScript/React and requires client-side rendering.

## Requirements

### System Requirements

- **Chrome or Chromium browser** must be installed and accessible in your PATH
- Supported browsers:
  - Google Chrome
  - Chromium
  - Chrome Canary
  - Microsoft Edge (Chromium-based)
  - Brave Browser

### Installation Options

#### macOS

```bash
# Via Homebrew
brew install --cask google-chrome
# or
brew install chromium
# or
brew install --cask brave-browser

# Via direct download
# https://www.google.com/chrome/
# https://brave.com/
```

#### Linux

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install google-chrome-stable
# or
sudo apt-get install chromium-browser
# or
sudo snap install brave

# CentOS/RHEL/Fedora
sudo yum install google-chrome-stable
# or
sudo dnf install chromium
# or
sudo dnf install brave-browser
```

#### Windows

```cmd
# Via Chocolatey
choco install googlechrome
# or
choco install chromium
# or
choco install brave

# Via Scoop
scoop install googlechrome
# or
scoop install chromium
# or
scoop install brave

# Alternative: Download directly from vendors
```

## Usage

### Configuration

Add a marketplace source to your `agents-config.yaml`:

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

### CLI Commands

#### List Categories

```bash
agent-manager marketplace list
```

#### List Agents in a Category

```bash
agent-manager marketplace list --category "Development"
```


#### Show Agent Details

```bash
agent-manager marketplace show "code-reviewer"
```

#### Refresh Cache

```bash
agent-manager marketplace refresh
```

### Installation

Once you find an agent you want to install, add it to your `agents-config.yaml`:

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

## Architecture

### Browser Automation

The marketplace integration uses [chromedp](https://github.com/chromedp/chromedp) for browser automation to handle the JavaScript-rendered content on subagents.sh. This approach:

- Executes JavaScript to render the full page content
- Extracts agent and category data using DOM queries
- Handles dynamic content loading
- Provides robust scraping of modern web applications

### Caching

The scraper includes intelligent caching to minimize API calls:

- Configurable TTL (time-to-live) for cached data
- Memory-efficient storage using Ristretto cache
- Cache invalidation and refresh capabilities

### Error Handling

- Graceful degradation when Chrome/Chromium is not available
- Fallback content generation for agents without accessible content
- Comprehensive logging for debugging scraping issues

## Troubleshooting

### "executable file not found in $PATH"

This error indicates Chrome/Chromium is not installed or not in your PATH. Install a supported browser using the installation options above.

### "context deadline exceeded"

This typically indicates:

- Slow network connection to subagents.sh
- JavaScript taking longer than expected to load
- Browser startup issues

Try increasing the timeout or checking your network connection.

### "No categories found" or "No agents found"

This may indicate:

- Changes to the subagents.sh website structure
- JavaScript failing to execute properly
- Browser automation issues

Check the logs for detailed error messages and consider refreshing the cache.

## Development

### Testing

The marketplace integration includes comprehensive unit tests, but integration tests require a Chrome/Chromium installation. To run tests:

```bash
make test
```

### Debugging

Enable verbose logging to see detailed browser automation logs:

```bash
agent-manager marketplace list --verbose
```

### Contributing

When contributing to the marketplace integration:

1. Ensure changes work with headless Chrome
2. Add appropriate error handling for browser failures
3. Update tests for any new functionality
4. Test against the live subagents.sh website