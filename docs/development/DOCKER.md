# Docker Deployment

Deploy and run Agent Manager in Docker containers for isolated, reproducible environments. This deployment method is ideal for CI/CD pipelines, cloud deployments, and development environments.

## Prerequisites

- Docker installed and running
- Docker Compose (optional, for easier management)

## Building the Image

```bash
# Using Make
make docker-build

# Or using Docker directly
docker build -f docker/Dockerfile -t agent-manager:latest .

# Build with specific version tag
docker build -f docker/Dockerfile -t agent-manager:v1.0.0 --build-arg VERSION=v1.0.0 .
```

## Basic Usage

### Show Help

```bash
docker run --rm agent-manager:latest --help
```

### Validate Configuration

```bash
docker run --rm \
  -v $(pwd)/agents-config.yaml:/app/agents-config.yaml:ro \
  agent-manager:latest validate
```

### Install Agents

```bash
docker run --rm \
  -v $(pwd)/agents-config.yaml:/app/agents-config.yaml:ro \
  -v agent-data:/app/.claude/agents \
  -v tracker-data:/app/.agent-manager \
  agent-manager:latest install
```

### List Installed Agents

```bash
docker run --rm \
  -v tracker-data:/app/.agent-manager:ro \
  agent-manager:latest list
```

## Environment Variables

Pass environment variables for authentication:

```bash
docker run --rm \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  -v $(pwd)/agents-config.yaml:/app/agents-config.yaml:ro \
  -v agent-data:/app/.claude/agents \
  agent-manager:latest install
```

## Docker Compose

### Basic Commands

```bash
# Build the image
docker-compose -f docker/docker-compose.yml build

# Run install command
docker-compose -f docker/docker-compose.yml run agent-manager install

# Run with specific source
docker-compose -f docker/docker-compose.yml run agent-manager install --source awesome-claude-code-subagents

# List installed agents
docker-compose -f docker/docker-compose.yml run agent-manager list

# Update agents
docker-compose -f docker/docker-compose.yml run agent-manager update
```

### Custom Configuration

```bash
# Set version and log level
VERSION=v1.0.0 LOG_LEVEL=debug docker-compose -f docker/docker-compose.yml run agent-manager install

# Use different config file
docker-compose -f docker/docker-compose.yml run \
  -v $(pwd)/custom-config.yaml:/app/agents-config.yaml:ro \
  agent-manager install
```

## Volume Mounts

| Volume/Mount | Purpose | Mode |
|--------------|---------|------|
| `/app/agents-config.yaml` | Configuration file | Read-only |
| `/app/.claude/agents` | Installed agents directory | Read-write |
| `/app/.agent-manager` | Installation tracking data | Read-write |

### Custom Paths

```bash
docker run --rm \
  -v /path/to/config.yaml:/app/agents-config.yaml:ro \
  -v /path/to/agents:/app/.claude/agents \
  -v /path/to/tracker:/app/.agent-manager \
  agent-manager:latest install
```

## Security Features

- **Minimal base image**: Uses `scratch` (empty) image
- **Non-root user**: Runs as `appuser`, not root
- **Read-only filesystem**: Container filesystem is read-only
- **No shell**: No shell or package manager in final image
- **Static binary**: Single static binary with no dependencies
- **Resource limits**: CPU and memory limits in docker-compose.yml

## CI/CD Integration

### GitHub Actions

```yaml
name: Install Agents
on:
  push:
    branches: [main]

jobs:
  install:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: docker build -f docker/Dockerfile -t agent-manager:${{ github.sha }} .

      - name: Install agents
        run: |
          docker run --rm \
            -v ${{ github.workspace }}/agents-config.yaml:/app/agents-config.yaml:ro \
            -v agents:/app/.claude/agents \
            agent-manager:${{ github.sha }} install
```

### GitLab CI

```yaml
install-agents:
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -f docker/Dockerfile -t agent-manager:$CI_COMMIT_SHA .
    - docker run --rm
        -v $CI_PROJECT_DIR/agents-config.yaml:/app/agents-config.yaml:ro
        -v agents:/app/.claude/agents
        agent-manager:$CI_COMMIT_SHA install
```

## Troubleshooting

### Permission Denied

Create volumes with proper permissions:

```bash
docker volume create agent-data
docker volume create tracker-data
```

### Configuration Not Found

Mount configuration file with absolute path:

```bash
docker run -v $(realpath agents-config.yaml):/app/agents-config.yaml:ro ...
```

### Network Issues

For private repositories, ensure authentication:

```bash
# Pass GitHub token
docker run -e GITHUB_TOKEN=$GITHUB_TOKEN ...

# For SSH authentication (not recommended for production)
docker run -v ~/.ssh:/home/appuser/.ssh:ro ...
```

### Debugging

Build debug image with shell:

```bash
# Build debug image
docker build -f docker/Dockerfile --target builder -t agent-manager:debug .
docker run --rm -it agent-manager:debug /bin/sh
```

## Related Documentation

- [`install`](../commands/INSTALL.md) - Installation commands
- [`list`](../commands/ADVANCED.md#list-command) - List agents
- [`update`](../commands/ADVANCED.md#update-command) - Update agents

## See Also

- [Docker Compose file](../../docker/docker-compose.yml)
- [Dockerfile](../../docker/Dockerfile)
- [Configuration Guide](../guides/CONFIGURATION.md)
