# Configuration Schema Reference

Complete YAML configuration schema for Agent Manager.

## Top-Level Structure

```yaml
settings:         # Global settings
sources:          # Array of agent sources
marketplace:      # Marketplace configuration (optional)
query:            # Query and indexing configuration (optional)
```

## Settings Section

Global configuration that applies to all sources.

```yaml
settings:
  base_dir: string                    # Default: .claude/agents
  backup_dir: string                  # Default: .claude/backups
  state_dir: string                   # Default: .agent-manager
  conflict_strategy: enum             # backup|overwrite|skip|merge
  timeout_seconds: integer            # Default: 300
  parallel_operations: integer        # Default: 2
  cache_enabled: boolean              # Default: true
  cache_dir: string                   # Default: .agent-manager/cache
  log_level: enum                     # debug|info|warn|error
  color_output: boolean               # Default: true
```

### Field Descriptions

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `base_dir` | string | `.claude/agents` | Base directory for agent installation |
| `backup_dir` | string | `.claude/backups` | Directory for file backups |
| `state_dir` | string | `.agent-manager` | Directory for state tracking |
| `conflict_strategy` | enum | `backup` | Global conflict resolution strategy |
| `timeout_seconds` | integer | `300` | Operation timeout in seconds |
| `parallel_operations` | integer | `2` | Number of concurrent operations |
| `cache_enabled` | boolean | `true` | Enable caching |
| `cache_dir` | string | `.agent-manager/cache` | Cache directory |
| `log_level` | string | `info` | Logging verbosity |
| `color_output` | boolean | `true` | Enable colored terminal output |

## Sources Section

Array of source configurations.

```yaml
sources:
  - name: string                      # Required: Unique identifier
    type: enum                        # Required: github|git|local|subagents
    enabled: boolean                  # Default: true
    description: string               # Optional: Human-readable description

    # Type-specific fields
    repository: string                # GitHub type only
    url: string                       # Git type only
    branch: string                    # GitHub/Git types
    tag: string                       # GitHub/Git types
    commit: string                    # GitHub/Git types

    # Paths
    paths:
      source: string                  # Source directory/path
      target: string                  # Target installation path

    # Authentication
    auth:
      method: enum                    # token|ssh|basic
      token_env: string               # Environment variable name
      token: string                   # Direct token (not recommended)
      username: string                # Basic auth username
      password_env: string            # Basic auth password env var
      ssh_key: string                 # Path to SSH key

    # Filtering
    filters:
      include:
        patterns: array<string>       # Glob patterns to include
        extensions: array<string>     # File extensions to include
        regex: array<string>          # Regular expressions
      exclude:
        patterns: array<string>       # Glob patterns to exclude
        extensions: array<string>     # File extensions to exclude
        regex: array<string>          # Regular expressions

    # Transformations
    transformations:
      - type: enum                    # Transformation type
        options: object               # Type-specific options

    # Conflict resolution
    conflict_strategy: enum           # Override global strategy

    # Caching
    cache:
      enabled: boolean                # Enable source caching
      ttl_hours: integer              # Cache time-to-live
      max_size_mb: integer            # Maximum cache size

    # Advanced
    watch: boolean                    # Watch for changes (local type)
    auto_update: boolean              # Auto-update on changes
    validate_ssl: boolean             # SSL certificate validation
    follow_redirects: boolean         # Follow HTTP redirects
    max_redirects: integer            # Maximum redirects to follow
```

## Source Types

### GitHub Source

```yaml
sources:
  - name: github-example
    type: github
    repository: owner/repo-name       # Required
    branch: main                      # Optional: default branch
    tag: v1.0.0                       # Optional: specific tag
    commit: abc123                    # Optional: specific commit
    paths:
      source: agents                  # Subdirectory in repo
      target: .claude/agents          # Local installation path
    auth:
      method: token
      token_env: GITHUB_TOKEN         # GitHub personal access token
```

### Git Source

```yaml
sources:
  - name: git-example
    type: git
    url: https://gitlab.com/user/repo.git  # Required
    branch: develop
    auth:
      method: ssh
      ssh_key: ~/.ssh/id_rsa
    paths:
      source: /
      target: .claude/agents/gitlab
```

### Local Source

```yaml
sources:
  - name: local-example
    type: local
    paths:
      source: ~/my-local-agents       # Required: absolute or ~ path
      target: .claude/agents/local
    watch: true                       # Auto-sync changes
    filters:
      include:
        extensions: [.md]
      exclude:
        patterns: ["test-*", "*.draft"]
```

### Subagents Marketplace

```yaml
sources:
  - name: marketplace-example
    type: subagents
    category: Development             # Optional: filter by category
    paths:
      target: .claude/agents/marketplace
    cache:
      enabled: true
      ttl_hours: 24
      max_size_mb: 100
    filters:
      include:
        regex: ["^(code-reviewer|docs).*\\.md$"]
```

## Transformations

### Available Transformation Types

#### remove_numeric_prefix

Removes numeric prefixes from directory names.

```yaml
transformations:
  - type: remove_numeric_prefix
    options:
      pattern: "^[0-9]+-"             # Regex pattern to match
      recursive: true                 # Apply recursively
```

#### extract_docs

Extracts documentation to separate directory.

```yaml
transformations:
  - type: extract_docs
    options:
      target_dir: .claude/docs        # Documentation target
      patterns: ["*.md", "*.txt"]     # Files to extract
      keep_structure: true            # Preserve directory structure
```

#### custom_script

Run custom transformation script.

```yaml
transformations:
  - type: custom_script
    options:
      script: ./transform.sh          # Script path
      args: ["--format", "claude"]    # Script arguments
      timeout: 60                     # Script timeout seconds
```

#### rename

Rename files based on patterns.

```yaml
transformations:
  - type: rename
    options:
      pattern: "(.*)\.txt$"           # Match pattern
      replacement: "$1.md"            # Replacement pattern
```

## Variable Substitution

### Syntax

Variables use `${scope.name}` syntax.

### Available Scopes

#### settings

Reference settings values:

```yaml
sources:
  - paths:
      target: ${settings.base_dir}/custom
```

#### env

Reference environment variables:

```yaml
sources:
  - repository: ${env.GITHUB_ORG}/agents
    auth:
      token_env: ${env.TOKEN_NAME}
```

#### source

Reference current source fields:

```yaml
sources:
  - name: my-source
    paths:
      target: .claude/${source.name}
```

### Examples

```yaml
settings:
  base_dir: ${env.HOME}/.claude/agents

sources:
  - name: ${env.TEAM_NAME}-agents
    type: github
    repository: ${env.GITHUB_ORG}/${env.REPO_NAME}
    enabled: ${env.ENABLE_AGENTS}
    paths:
      target: ${settings.base_dir}/${source.name}
```

## Marketplace Configuration

```yaml
marketplace:
  browser_path: string                # Custom browser executable
  headless: boolean                   # Run browser headless
  timeout: integer                    # Browser timeout seconds
  viewport:
    width: integer                    # Browser width
    height: integer                   # Browser height
  user_agent: string                  # Custom user agent
```

## Query Configuration

Configuration for agent search, indexing, and analytics capabilities.

```yaml
query:
  index:
    enabled: boolean                  # Enable search indexing
    path: string                      # Index storage path
    update_on_install: boolean        # Auto-update index on install
    rebuild_interval: string          # Auto-rebuild interval (e.g., "24h")

  cache:
    enabled: boolean                  # Enable query result caching
    ttl: string                       # Cache time-to-live (e.g., "1h")
    max_size: string                  # Maximum cache size (e.g., "100MB")

  defaults:
    format: string                    # Default output format (table/json/yaml)
    limit: integer                    # Default result limit
    fuzzy: boolean                    # Enable fuzzy matching by default

  validation:
    check_name_format: boolean        # Enforce lowercase-hyphen naming
    check_required_fields: boolean    # Ensure name & description exist
    check_tool_validity: boolean      # Verify tools are valid Claude Code tools
```

### Query Field Descriptions

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `query.index.enabled` | boolean | `true` | Enable search index for fast queries |
| `query.index.path` | string | `${settings.base_dir}/.agent-index` | Index storage location |
| `query.index.update_on_install` | boolean | `true` | Automatically update index after installs |
| `query.index.rebuild_interval` | string | `24h` | How often to rebuild the index |
| `query.cache.enabled` | boolean | `true` | Enable query result caching |
| `query.cache.ttl` | string | `1h` | How long to cache query results |
| `query.cache.max_size` | string | `100MB` | Maximum cache storage |
| `query.defaults.format` | string | `table` | Default output format |
| `query.defaults.limit` | integer | `20` | Default number of results |
| `query.defaults.fuzzy` | boolean | `true` | Enable fuzzy matching |
| `query.validation.check_name_format` | boolean | `true` | Enforce name format rules |
| `query.validation.check_required_fields` | boolean | `true` | Check for required fields |
| `query.validation.check_tool_validity` | boolean | `true` | Validate tool names |

## Complete Example

```yaml
# Global settings
settings:
  base_dir: ~/.claude/agents
  backup_dir: ~/.claude/backups
  conflict_strategy: backup
  timeout_seconds: 300
  log_level: info

# Marketplace configuration
marketplace:
  headless: true
  timeout: 60

# Query and indexing configuration
query:
  index:
    enabled: true
    path: ~/.claude/.agent-index
    update_on_install: true
    rebuild_interval: 24h
  cache:
    enabled: true
    ttl: 1h
    max_size: 100MB
  defaults:
    format: table
    limit: 20
    fuzzy: true
  validation:
    check_name_format: true
    check_required_fields: true
    check_tool_validity: true

# Agent sources
sources:
  # Public GitHub repository
  - name: awesome-agents
    type: github
    repository: VoltAgent/awesome-claude-code-subagents
    description: "Community-curated agents"
    transformations:
      - type: remove_numeric_prefix
      - type: extract_docs
        options:
          target_dir: ~/.claude/docs

  # Private GitHub repository
  - name: company-agents
    type: github
    repository: ${env.COMPANY_ORG}/private-agents
    branch: production
    enabled: ${env.ENABLE_PRIVATE}
    auth:
      method: token
      token_env: GITHUB_TOKEN
    conflict_strategy: skip

  # GitLab repository with SSH
  - name: gitlab-agents
    type: git
    url: git@gitlab.com:team/agents.git
    auth:
      method: ssh
      ssh_key: ~/.ssh/gitlab_key
    filters:
      include:
        extensions: [.md]
      exclude:
        patterns: ["draft-*", "*.tmp"]

  # Local development
  - name: local-dev
    type: local
    paths:
      source: ~/projects/my-agents
      target: ~/.claude/agents/dev
    watch: true
    enabled: ${env.LOCAL_DEV}

  # Marketplace with filtering
  - name: marketplace-filtered
    type: subagents
    category: "Development"
    filters:
      include:
        regex: ["^(code-reviewer|test).*"]
    cache:
      enabled: true
      ttl_hours: 6
```

## Validation Rules

1. **Required Fields**:
   - Each source must have `name` and `type`
   - GitHub sources require `repository`
   - Git sources require `url`
   - Local sources require `paths.source`

2. **Unique Names**:
   - Source names must be unique within configuration

3. **Valid Enums**:
   - `type`: github, git, local, subagents
   - `conflict_strategy`: backup, overwrite, skip, merge
   - `auth.method`: token, ssh, basic

4. **Path Requirements**:
   - Absolute paths or paths starting with `~`
   - No trailing slashes

5. **Environment Variables**:
   - Must exist when referenced in enabled fields
   - Token environment variables checked at runtime

## Best Practices

1. **Use Environment Variables** for sensitive data
2. **Set Appropriate Timeouts** based on network speed
3. **Enable Caching** for frequently accessed sources
4. **Use Filters** to minimize unnecessary files
5. **Document Sources** with description field
6. **Test with Dry Run** before applying changes
7. **Use Specific Branches/Tags** for stability

## See Also

- [Configuration Guide](../guides/CONFIGURATION.md)
- [CLI Reference](CLI-REFERENCE.md)
- [Workflows](../guides/WORKFLOWS.md)
