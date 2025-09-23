# Common Workflows

Practical workflows and patterns for using Agent Manager effectively.

## Initial Setup Workflow

### 1. Create Your Configuration

```bash
# Start with the default configuration
cp agents-config.yaml my-config.yaml

# Edit to suit your needs
nano my-config.yaml
```

### 2. Validate Configuration

```bash
agent-manager validate --config my-config.yaml
```

### 3. Preview Installation

```bash
agent-manager install --config my-config.yaml --dry-run
```

### 4. Install Agents

```bash
agent-manager install --config my-config.yaml
```

## Daily Operations

### Check Installation Status

```bash
# List all installed agents
agent-manager list

# Search for specific agents
agent-manager query "code review"

# Get detailed info about an agent
agent-manager show code-reviewer

# View installation statistics
agent-manager stats --detailed

# Check specific source
agent-manager list --source awesome-claude-code-subagents
```

### Keep Agents Updated

```bash
# Check for updates
agent-manager update --check-only

# Update everything
agent-manager update

# Update specific source
agent-manager update --source my-agents
```

### Add New Sources

1. Edit your configuration file
2. Add the new source:

   ```yaml
   sources:
     - name: new-source
       type: github
       repository: owner/repo
   ```

3. Validate: `agent-manager validate`
4. Install: `agent-manager install --source new-source`

## Source-Specific Workflows

### GitHub Repository

#### Public Repository

```yaml
sources:
  - name: public-agents
    type: github
    repository: owner/repo-name
    branch: main
    paths:
      source: agents
      target: .claude/agents
```

#### Private Repository

```yaml
sources:
  - name: private-agents
    type: github
    repository: company/private-repo
    auth:
      method: token
      token_env: GITHUB_TOKEN
```

Set up authentication:

```bash
export GITHUB_TOKEN=your_personal_access_token
agent-manager install --source private-agents
```

### Git Repository

#### HTTPS with Authentication

```yaml
sources:
  - name: gitlab-agents
    type: git
    url: https://gitlab.com/user/agents.git
    auth:
      method: token
      token_env: GITLAB_TOKEN
    branch: develop
```

#### SSH Repository

```yaml
sources:
  - name: ssh-agents
    type: git
    url: git@gitlab.com:user/agents.git
    auth:
      method: ssh
      ssh_key: ~/.ssh/id_rsa
```

### Local Development

```yaml
sources:
  - name: local-dev
    type: local
    paths:
      source: ~/my-local-agents
      target: .claude/agents/local
    watch: true  # Auto-sync changes
```

Workflow:

```bash
# Develop locally
cd ~/my-local-agents
# Make changes...

# Sync to Agent Manager
agent-manager install --source local-dev

# Test changes
# When ready, push to repository
```

### Marketplace Integration

```yaml
sources:
  - name: marketplace-all
    type: subagents
    paths:
      target: .claude/agents/marketplace
    cache:
      enabled: true
      ttl_hours: 1
```

Browse and install:

```bash
# Discover agents
agent-manager marketplace list

# View categories
agent-manager marketplace list --category "Development"

# Install from marketplace
agent-manager install --source marketplace-all
```

## Team Collaboration Workflows

### Shared Configuration

1. Create team configuration repository:

   ```bash
   git init team-agent-config
   cd team-agent-config
   ```

2. Add team configuration:

   ```yaml
   # team-agents.yaml
   sources:
     - name: team-shared
       type: github
       repository: team/shared-agents
     - name: team-tools
       type: github
       repository: team/tool-agents
   ```

3. Team members clone and use:

   ```bash
   git clone team-config-repo
   agent-manager install --config team-config-repo/team-agents.yaml
   ```

### Department-Specific Setup

```yaml
# engineering-agents.yaml
sources:
  - name: engineering-base
    type: github
    repository: company/eng-agents

  - name: frontend-specific
    type: github
    repository: company/frontend-agents
    enabled: ${env.TEAM_FRONTEND}

  - name: backend-specific
    type: github
    repository: company/backend-agents
    enabled: ${env.TEAM_BACKEND}
```

Usage:

```bash
# Frontend team
export TEAM_FRONTEND=true
agent-manager install --config engineering-agents.yaml

# Backend team
export TEAM_BACKEND=true
agent-manager install --config engineering-agents.yaml
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Update Agents
on:
  schedule:
    - cron: '0 0 * * *'  # Daily
  workflow_dispatch:

jobs:
  update-agents:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.24'

      - name: Build Agent Manager
        run: make build

      - name: Update Agents
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          ./bin/agent-manager update
          ./bin/agent-manager list > agents-inventory.txt

      - name: Commit Updates
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add agents-inventory.json
          git commit -m "Update agents inventory" || true
          git push
```

### GitLab CI

```yaml
update-agents:
  image: golang:1.24
  script:
    - make build
    - ./bin/agent-manager update
    - ./bin/agent-manager list > agents.txt
  artifacts:
    paths:
      - agents.txt
    expire_in: 1 week
  only:
    - schedules
```

## Troubleshooting Workflows

### Debug Installation Issues

```bash
# Verbose validation
agent-manager validate --verbose

# Dry run with details
agent-manager install --dry-run --verbose

# Check specific source
agent-manager install --source problematic-source --verbose
```

### Reset and Reinstall

```bash
# Complete reset
agent-manager uninstall --all
rm -rf ~/.agent-manager/state.json

# Fresh install
agent-manager validate
agent-manager install --verbose
```

### Conflict Resolution

When conflicts occur:

```bash
# Option 1: Preview conflicts first
agent-manager install --dry-run

# Option 2: Configure strategy in agents-config.yaml and install
# Available strategies: skip, backup, overwrite
agent-manager install

# Option 3: Manual resolution - review and handle files manually
# then install with appropriate strategy configured
```

## Agent Discovery and Analysis Workflows

### Finding the Right Agent

```bash
# Search for agents by functionality
agent-manager query "typescript"
agent-manager query --description "code review"

# Find agents using specific tools
agent-manager query --tools "Read,Write"

# Interactive search and selection
agent-manager query --interactive "documentation"
```

### Analyzing Your Agent Collection

```bash
# Get overview statistics
agent-manager stats

# Analyze completeness and quality
agent-manager stats --validation

# Understand tool usage patterns
agent-manager stats --tools

# Get detailed statistics by source
agent-manager stats --detailed
```

### Agent Quality Assurance

```bash
# Validate configuration file
agent-manager validate

# Validate all installed agents
agent-manager validate --agents

# Test query functionality
agent-manager validate --query
```

### Performance Optimization

```bash
# Build search index for fast queries
agent-manager index build

# Check index statistics
agent-manager index stats

# Check cache performance
agent-manager index cache-stats

# Update index after installations
agent-manager index build
```

## Maintenance Workflows

### Regular Maintenance

Weekly routine:

```bash
# Monday: Check for updates
agent-manager update --check-only

# Wednesday: Apply updates
agent-manager update

# Friday: Verify and clean
agent-manager list --verbose
agent-manager validate
```

### Backup Strategy

```bash
# Backup configuration
cp agents-config.yaml agents-config.yaml.backup

# Backup state
cp -r ~/.agent-manager ~/.agent-manager.backup

# Backup agents
tar -czf agents-backup.tar.gz ~/.claude/agents
```

### Migration to New System

```bash
# On old system
agent-manager list > agents-manifest.txt
tar -czf agent-data.tar.gz ~/.claude/agents ~/.agent-manager

# On new system
tar -xzf agent-data.tar.gz
agent-manager validate
agent-manager list
```

## Best Practices

1. **Version Control Your Config**: Keep `agents-config.yaml` in Git
2. **Use Environment Variables**: For sensitive data like tokens
3. **Regular Updates**: Schedule weekly or monthly updates
4. **Test in Dry Run**: Always preview changes before applying
5. **Document Custom Sources**: Add comments to your configuration
6. **Backup Before Major Changes**: Save state and configuration
7. **Use Specific Sources**: Install/update individual sources when possible

## See Also

- [Configuration Guide](CONFIGURATION.md)
- [Install Command](../commands/INSTALL.md)
- [Troubleshooting](TROUBLESHOOTING.md)
