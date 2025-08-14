# Configuration Reference

This document provides a complete reference for the Agent Manager YAML configuration file.

## Configuration File Location

Default locations (in order of preference):
1. `--config` command-line flag
2. `AGENT_MANAGER_CONFIG` environment variable
3. `agents-config.yaml` in current directory

## Schema Overview

```yaml
version: "1.0"
settings:
  # Global settings
sources:
  # Array of agent sources
metadata:
  # Tracking and logging configuration
```

## Version

**Type**: `string`
**Required**: Yes
**Default**: `"1.0"`

Configuration schema version. Currently only `"1.0"` is supported.

```yaml
version: "1.0"
```

## Settings

Global settings that apply to all sources.

### base_dir

**Type**: `string`
**Default**: `.claude/agents`

Base directory where agents will be installed.

```yaml
settings:
  base_dir: .claude/agents
```

### docs_dir

**Type**: `string`
**Default**: `docs`

Directory where extracted documentation will be placed.

```yaml
settings:
  docs_dir: docs
```

### conflict_strategy

**Type**: `string`
**Default**: `backup`
**Values**: `backup`, `overwrite`, `skip`, `merge`

How to handle file conflicts when installing agents over existing files.

| Strategy | Installation Behavior | Uninstall Behavior | Backup Created | Best For |
|----------|----------------------|-------------------|----------------|----------|
| `backup` | Creates backup, then overwrites | Restores original from backup | ✅ Yes | Most scenarios - provides safety while allowing updates |
| `overwrite` | Immediately overwrites existing file | Preserves file with new content | ❌ No | When you don't need original content |
| `skip` | Keeps existing file unchanged | File remains completely unchanged | ❌ Not needed | Preserving local customizations |
| `merge` | Creates backup, attempts intelligent merge | Restores original from backup | ✅ Yes | When you want to combine changes |

**Detailed Behavior:**

- **`backup`**: Original file is copied to `.claude/backups/` with timestamp, then overwritten. During uninstall, original content is restored.
- **`overwrite`**: No backup is created. Original content is permanently lost. During uninstall, file remains but contains new content.
- **`skip`**: Conflicting files are not installed at all. Existing files remain exactly as they were.
- **`merge`**: Performs three-way merge between original, existing, and incoming content. Creates conflict markers when automatic merge fails.

```yaml
settings:
  conflict_strategy: backup
```

### backup_dir

**Type**: `string`
**Default**: `.claude/backups`

Directory where backup files are stored.

```yaml
settings:
  backup_dir: .claude/backups
```

### log_level

**Type**: `string`
**Default**: `info`
**Values**: `debug`, `info`, `warn`, `error`

Logging verbosity level.

```yaml
settings:
  log_level: info
```

### concurrent_downloads

**Type**: `integer`
**Default**: `3`
**Range**: `1-10`

Number of sources to process concurrently.

```yaml
settings:
  concurrent_downloads: 3
```

### timeout

**Type**: `duration`
**Default**: `300s`

Timeout for source operations (clone, download, etc.).

```yaml
settings:
  timeout: 300s  # 5 minutes
```

### continue_on_error

**Type**: `boolean`
**Default**: `false`

Whether to continue processing other sources if one fails.

```yaml
settings:
  continue_on_error: false
```

## Sources

Array of agent sources to install from.

### Common Source Fields

#### name

**Type**: `string`
**Required**: Yes

Unique identifier for the source.

```yaml
sources:
  - name: awesome-claude-code-subagents
```

#### enabled

**Type**: `boolean`
**Default**: `true`

Whether this source should be processed.

```yaml
sources:
  - name: my-source
    enabled: true
```

#### type

**Type**: `string`
**Required**: Yes
**Values**: `github`, `git`, `local`

Type of source.

```yaml
sources:
  - name: my-source
    type: github
```

#### paths

**Type**: `object`
**Required**: Yes

Source and target path configuration.

```yaml
paths:
  source: categories    # Path within source
  target: .claude/agents  # Local target path
```

##### source

**Type**: `string`

Path within the source repository/directory to copy from.

##### target

**Type**: `string`

Local path where files will be installed. Supports variable substitution.

#### conflict_strategy

**Type**: `string`
**Values**: `backup`, `overwrite`, `skip`, `merge`

Override global conflict strategy for this source.

```yaml
sources:
  - name: experimental-agents
    conflict_strategy: skip
```

### GitHub Sources

For GitHub repositories using `type: github`.

#### repository

**Type**: `string`
**Required**: Yes
**Format**: `owner/repo-name`

GitHub repository in owner/repo format.

```yaml
sources:
  - name: my-github-source
    type: github
    repository: VoltAgent/awesome-claude-code-subagents
```

#### branch

**Type**: `string`
**Default**: `main`

Git branch to use.

```yaml
sources:
  - name: my-source
    type: github
    repository: owner/repo
    branch: develop
```

### Git Sources

For generic Git repositories using `type: git`.

#### url

**Type**: `string`
**Required**: Yes

Git repository URL (HTTPS or SSH).

```yaml
sources:
  - name: my-git-source
    type: git
    url: https://gitlab.com/user/agents.git
```

#### branch

**Type**: `string`
**Default**: Repository default

Git branch to use.

```yaml
sources:
  - name: my-source
    type: git
    url: https://example.com/repo.git
    branch: feature-branch
```

### Local Sources

For local file system sources using `type: local`.

#### watch

**Type**: `boolean`
**Default**: `false`

Watch for changes in development mode (future feature).

```yaml
sources:
  - name: local-dev
    type: local
    paths:
      source: ~/my-agents
      target: .claude/agents/local
    watch: true
```

### Authentication

Authentication configuration for sources that require it.

#### auth.method

**Type**: `string`
**Values**: `token`, `ssh`

Authentication method.

```yaml
auth:
  method: token
```

#### auth.token_env

**Type**: `string`

Environment variable containing the authentication token.

```yaml
auth:
  method: token
  token_env: GITHUB_TOKEN
```

#### auth.ssh_key

**Type**: `string`

Path to SSH private key file.

```yaml
auth:
  method: ssh
  ssh_key: ~/.ssh/id_rsa
```

### Filters

Control which files are included or excluded.

#### include

Inclusion filters. If specified, only matching files are included.

##### extensions

**Type**: `array of strings`

File extensions to include.

```yaml
filters:
  include:
    extensions: [".md", ".yaml", ".json"]
```

##### patterns

**Type**: `array of strings`

Glob patterns to include.

```yaml
filters:
  include:
    patterns: ["*-agent/*", "docs/*"]
```

##### regex

**Type**: `array of strings`

Regular expressions to match file paths.

```yaml
filters:
  include:
    regex: ["^(core|data).*\\.md$"]
```

#### exclude

Exclusion filters. Matching files are excluded.

##### patterns

**Type**: `array of strings`

Glob patterns to exclude.

```yaml
filters:
  exclude:
    patterns: ["test-*", "*.tmp", ".*"]
```

### Transformations

File transformations applied during installation.

#### type

**Type**: `string`
**Required**: Yes
**Values**: `remove_numeric_prefix`, `extract_docs`, `rename_files`, `replace_content`, `custom_script`

Type of transformation.

#### remove_numeric_prefix

Remove numeric prefixes from directory names.

```yaml
transformations:
  - type: remove_numeric_prefix
    pattern: "^[0-9]{2}-"  # Optional custom pattern
```

##### pattern

**Type**: `string`
**Default**: `"^[0-9]{2}-"`

Regular expression pattern to remove.

#### extract_docs

Extract documentation files to a separate directory.

```yaml
transformations:
  - type: extract_docs
    source_pattern: "*/README.md"
    target_dir: docs
    naming: UPPERCASE_UNDERSCORE
```

##### source_pattern

**Type**: `string`
**Default**: `"*/README.md"`

Pattern matching files to extract.

##### target_dir

**Type**: `string`
**Required**: Yes

Directory where extracted docs are placed.

##### naming

**Type**: `string`
**Default**: `UPPERCASE_UNDERSCORE`
**Values**: `UPPERCASE_UNDERSCORE`, `lowercase_dash`, `CamelCase`

How to transform extracted file names.

#### custom_script

Run a custom script for transformation.

```yaml
transformations:
  - type: custom_script
    script: ./transform-agents.sh
    args: ["--format", "claude"]
```

##### script

**Type**: `string`
**Required**: Yes

Path to the script to execute.

##### args

**Type**: `array of strings`

Arguments to pass to the script.

### Post-Install Actions

Actions to run after installation.

#### type

**Type**: `string`
**Required**: Yes
**Values**: `script`, `command`

Type of post-install action.

#### script

Run a script after installation.

```yaml
post_install:
  - type: script
    path: scripts/fix-doc-links.sh
    args: ["--config", "${source.name}"]
```

##### path

**Type**: `string`
**Required**: Yes

Path to the script to execute.

##### args

**Type**: `array of strings`

Arguments to pass to the script.

## Metadata

Configuration for tracking and logging.

### tracking_file

**Type**: `string`
**Default**: `.claude/.installed-agents.json`

File where installation tracking data is stored.

```yaml
metadata:
  tracking_file: .claude/.installed-agents.json
```

### log_file

**Type**: `string`
**Default**: `.claude/installation.log`

File where detailed logs are written.

```yaml
metadata:
  log_file: .claude/installation.log
```

### lock_file

**Type**: `string`
**Default**: `.claude/.lock`

Lock file to prevent concurrent operations.

```yaml
metadata:
  lock_file: .claude/.lock
```

## Variable Substitution

The configuration supports variable substitution using `${variable}` syntax.

### Settings Variables

Reference settings values:

```yaml
settings:
  base_dir: .claude/agents

sources:
  - name: my-source
    paths:
      target: ${settings.base_dir}/custom
```

### Environment Variables

Reference environment variables:

```yaml
sources:
  - name: ${env.COMPANY}-agents
    repository: ${env.GITHUB_ORG}/agents
    auth:
      token_env: ${env.TOKEN_VAR}
```

### Source Variables

Reference source properties (in post-install scripts):

```yaml
post_install:
  - type: script
    path: scripts/process.sh
    args: ["--source", "${source.name}"]
```

## Complete Example

```yaml
version: "1.0"

settings:
  base_dir: .claude/agents
  docs_dir: docs
  conflict_strategy: backup
  backup_dir: .claude/backups
  log_level: info
  concurrent_downloads: 3
  timeout: 300s
  continue_on_error: false

sources:
  # GitHub source with full configuration
  - name: awesome-claude-code-subagents
    enabled: true
    type: github
    repository: VoltAgent/awesome-claude-code-subagents
    branch: main
    paths:
      source: categories
      target: ${settings.base_dir}
    filters:
      include:
        extensions: [".md"]
      exclude:
        patterns: [".*", "*.tmp", "test-*"]
    transformations:
      - type: remove_numeric_prefix
        pattern: "^[0-9]{2}-"
      - type: extract_docs
        source_pattern: "*/README.md"
        target_dir: ${settings.docs_dir}
        naming: UPPERCASE_UNDERSCORE
    post_install:
      - type: script
        path: scripts/fix-doc-links.sh

  # Private Git repository with authentication
  - name: company-agents
    enabled: true
    type: git
    url: https://gitlab.company.com/ai/claude-agents.git
    branch: develop
    auth:
      method: token
      token_env: GITLAB_TOKEN
    paths:
      source: agents
      target: ${settings.base_dir}/company
    filters:
      include:
        regex: ["^(production|staging).*\\.md$"]
    conflict_strategy: skip

  # Local development source
  - name: local-dev-agents
    enabled: false
    type: local
    paths:
      source: ~/development/my-agents
      target: ${settings.base_dir}/local
    filters:
      include:
        extensions: [".md", ".yaml"]
    watch: true

metadata:
  tracking_file: .claude/.installed-agents.json
  log_file: .claude/installation.log
  lock_file: .claude/.lock
```

## Validation

Use `agent-manager validate` to check your configuration:

```bash
# Validate default config
agent-manager validate

# Validate specific config
agent-manager validate --config my-config.yaml

# Verbose validation with warnings
agent-manager validate --verbose
```

Common validation errors:

- Invalid YAML syntax
- Missing required fields
- Invalid regex patterns
- Duplicate source names
- Invalid conflict strategies
- Invalid log levels