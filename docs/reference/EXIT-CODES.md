# Exit Codes Reference

Agent Manager uses standard exit codes to indicate the result of operations.

## Standard Exit Codes

| Code | Name | Description | Common Causes |
|------|------|-------------|---------------|
| **0** | SUCCESS | Operation completed successfully | Normal completion |
| **1** | ERROR | General or unspecified error | Unexpected failures, unhandled errors |
| **2** | CONFIG_ERROR | Configuration file error | Invalid YAML, missing required fields, file not found |
| **3** | INSTALL_ERROR | Installation failed | Permission denied, file conflicts, write failures |
| **4** | NETWORK_ERROR | Network operation failed | Connection timeout, DNS failure, unreachable host |
| **5** | AUTH_ERROR | Authentication failed | Invalid token, expired credentials, access denied |
| **127** | COMMAND_NOT_FOUND | Command not found | Binary not in PATH, typo in command name |

## Detailed Error Scenarios

### Exit Code 0: Success

The operation completed without errors.

```bash
$ agent-manager install
✓ Installation complete
$ echo $?
0
```

### Exit Code 1: General Error

Unspecified or unexpected errors.

**Common causes:**

- Unhandled exceptions
- Internal errors
- Unknown failures

**Examples:**

```bash
$ agent-manager install
Error: unexpected error during installation
$ echo $?
1
```

**Resolution:**

- Check logs with `--verbose`
- Report issue if persistent

### Exit Code 2: Configuration Error

Problems with the configuration file or settings.

**Common causes:**

- YAML syntax errors
- Missing required fields
- Invalid field values
- File not found
- Permission to read config denied

**Examples:**

```bash
$ agent-manager validate
Error: agents-config.yaml: line 10: found undefined anchor
$ echo $?
2

$ agent-manager install --config missing.yaml
Error: configuration file not found: missing.yaml
$ echo $?
2
```

**Resolution:**

- Validate YAML syntax
- Check required fields
- Verify file permissions
- Use `validate --verbose` for details

### Exit Code 3: Installation Error

Failures during agent installation or file operations.

**Common causes:**

- Permission denied writing files
- Disk full
- File conflicts without resolution
- Invalid source paths
- Corrupted downloads

**Examples:**

```bash
$ agent-manager install
Error: permission denied: cannot write to /usr/local/agents
$ echo $?
3

$ agent-manager install
Error: disk full: cannot create .claude/agents/new-file.md
$ echo $?
3
```

**Resolution:**

- Check file permissions
- Verify disk space
- Use appropriate conflict strategy
- Run with sudo if needed

### Exit Code 4: Network Error

Network-related failures.

**Common causes:**

- No internet connection
- DNS resolution failure
- Connection timeout
- Host unreachable
- Proxy configuration issues

**Examples:**

```bash
$ agent-manager install --source github-agents
Error: failed to connect to github.com: connection timeout
$ echo $?
4

$ agent-manager marketplace list
Error: cannot reach subagents.sh: no route to host
$ echo $?
4
```

**Resolution:**

- Check internet connection
- Verify DNS settings
- Check proxy configuration
- Increase timeout with `--timeout`
- Test with `curl` or `ping`

### Exit Code 5: Authentication Error

Authentication or authorization failures.

**Common causes:**

- Invalid or expired token
- Missing credentials
- Incorrect username/password
- No repository access
- Rate limiting

**Examples:**

```bash
$ agent-manager install --source private-repo
Error: authentication failed: invalid GitHub token
$ echo $?
5

$ GITHUB_TOKEN=invalid agent-manager install
Error: 401 Unauthorized: bad credentials
$ echo $?
5
```

**Resolution:**

- Verify token is valid
- Check token permissions/scopes
- Ensure environment variables are set
- Test authentication separately
- Wait if rate-limited

### Exit Code 127: Command Not Found

The agent-manager command cannot be found.

**Common causes:**

- Binary not installed
- Not in system PATH
- Typo in command name

**Examples:**

```bash
$ agent-manager install
bash: agent-manager: command not found
$ echo $?
127
```

**Resolution:**

- Install Agent Manager
- Add to PATH
- Use full path to binary
- Check command spelling

## Using Exit Codes in Scripts

### Basic Error Handling

```bash
#!/bin/bash

if agent-manager install; then
    echo "Installation successful"
else
    exit_code=$?
    echo "Installation failed with code: $exit_code"

    case $exit_code in
        2)
            echo "Check your configuration file"
            ;;
        3)
            echo "Check file permissions"
            ;;
        4)
            echo "Check network connection"
            ;;
        5)
            echo "Check authentication credentials"
            ;;
        *)
            echo "Unknown error"
            ;;
    esac

    exit $exit_code
fi
```

### Retry Logic

```bash
#!/bin/bash

MAX_RETRIES=3
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    agent-manager update
    EXIT_CODE=$?

    if [ $EXIT_CODE -eq 0 ]; then
        echo "Update successful"
        break
    elif [ $EXIT_CODE -eq 4 ]; then
        # Network error - retry
        RETRY_COUNT=$((RETRY_COUNT + 1))
        echo "Network error, retry $RETRY_COUNT of $MAX_RETRIES"
        sleep 5
    else
        # Other error - don't retry
        echo "Update failed with code $EXIT_CODE"
        exit $EXIT_CODE
    fi
done
```

### CI/CD Integration

```yaml
# GitHub Actions
- name: Install Agents
  run: |
    agent-manager install
  continue-on-error: false

- name: Check specific errors
  if: failure()
  run: |
    if [ $? -eq 5 ]; then
      echo "::error::Authentication failed. Check GITHUB_TOKEN secret"
    fi
```

### Conditional Execution

```bash
#!/bin/bash

# Only proceed if validation succeeds
agent-manager validate || exit $?

# Install and check specific errors
agent-manager install
case $? in
    0)
        echo "Success - proceeding with tests"
        make test
        ;;
    3)
        echo "Installation error - check permissions"
        exit 3
        ;;
    *)
        echo "Unexpected error"
        exit 1
        ;;
esac
```

## Best Practices

1. **Always Check Exit Codes** in scripts
2. **Handle Specific Errors** appropriately
3. **Provide Clear Error Messages** to users
4. **Log Exit Codes** for debugging
5. **Use Appropriate Exit Codes** in custom scripts
6. **Document Expected Exit Codes** in automation

## Debugging Exit Codes

### Verbose Output

Get more information about failures:

```bash
agent-manager install --verbose
echo "Exit code: $?"
```

### Debug Mode

Enable debug logging:

```bash
DEBUG=true agent-manager install
```

### Check Logs

Review log files for details:

```bash
# Check recent logs
ls -lt ~/.agent-manager/logs/
tail -100 ~/.agent-manager/logs/latest.log
```

### Test Commands

Verify specific operations:

```bash
# Test configuration
agent-manager validate; echo "Validate: $?"

# Test network
curl -I https://github.com; echo "Network: $?"

# Test authentication
gh auth status; echo "Auth: $?"
```

## Custom Exit Codes

When writing scripts that use Agent Manager:

```bash
#!/bin/bash

# Use consistent exit codes
SUCCESS=0
CONFIG_ERROR=2
INSTALL_ERROR=3
NETWORK_ERROR=4
AUTH_ERROR=5

# Your script logic
if ! agent-manager validate; then
    exit $CONFIG_ERROR
fi

if ! agent-manager install; then
    # Preserve Agent Manager's exit code
    exit $?
fi

# Custom success
exit $SUCCESS
```

## See Also

- [CLI Reference](CLI-REFERENCE.md)
- [Troubleshooting Guide](../guides/TROUBLESHOOTING.md)
- [Configuration Guide](../guides/CONFIGURATION.md)
