# Contributing to Agent Manager

Welcome! We're excited that you want to contribute to Agent Manager. This guide will help you get started.

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.24.6+** installed
- **Git** for version control
- **Make** for build commands
- **GitHub CLI** (optional but recommended)

### Development Setup

1. **Fork and Clone**

   ```bash
   # Fork the repository on GitHub, then:
   gh repo clone your-username/claude-code-agent-manager
   cd claude-code-agent-manager

   # Add upstream remote
   git remote add upstream https://github.com/pacphi/claude-code-agent-manager.git
   ```

2. **Install Dependencies**

   ```bash
   make deps
   ```

3. **Build and Test**

   ```bash
   make build
   make test
   ```

4. **Verify Installation**

   ```bash
   ./bin/agent-manager --help
   ```

## Development Workflow

### Branch Strategy

We use a simple branching model:

- `main` - stable releases
- `feature/description` - new features
- `bugfix/description` - bug fixes
- `docs/description` - documentation updates

### Creating a Contribution

1. **Create a Branch**

   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make Changes**

   Follow our coding standards (see below)

3. **Test Thoroughly**

   ```bash
   make test
   make test-coverage
   make lint
   ```

4. **Commit Changes**

   Use conventional commits:

   ```bash
   git commit -m "feat: add marketplace filtering support"
   git commit -m "fix: resolve auth token validation issue"
   git commit -m "docs: update installation guide"
   ```

5. **Push and Create PR**

   ```bash
   git push origin feature/my-new-feature
   gh pr create --title "Add marketplace filtering support" --body "Description of changes"
   ```

## Code Standards

### Go Style

We follow standard Go conventions:

- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable and function names

### Code Organization

```
internal/
├── component/          # Each major component in its own package
│   ├── component.go    # Main implementation
│   ├── component_test.go # Unit tests
│   └── types.go        # Type definitions
```

### Error Handling

Use structured errors with context:

```go
// Good
return fmt.Errorf("failed to install source %s: %w", source.Name, err)

// Better
return &InstallationError{
    Source: source.Name,
    Cause:  err,
}
```

### Testing

#### Unit Tests

Every public function should have tests:

```go
func TestSourceHandler_Install(t *testing.T) {
    handler := NewSourceHandler()

    tests := []struct {
        name    string
        source  Source
        wantErr bool
    }{
        {
            name:   "valid source",
            source: Source{Name: "test", Type: "github"},
            wantErr: false,
        },
        {
            name:   "invalid source",
            source: Source{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := handler.Install(context.Background(), tt.source)
            if (err != nil) != tt.wantErr {
                t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### Integration Tests

Test complete workflows:

```go
func TestInstallWorkflow(t *testing.T) {
    // Set up test environment
    tmpDir := t.TempDir()
    config := createTestConfig(tmpDir)

    // Run the workflow
    installer := NewInstaller(config)
    err := installer.Install()

    // Verify results
    assert.NoError(t, err)
    assert.FileExists(t, filepath.Join(tmpDir, "expected-file.md"))
}
```

### Mocking

Use interfaces for testability:

```go
type GitHubClient interface {
    GetRepository(owner, repo string) (*Repository, error)
    DownloadArchive(owner, repo, ref string) (io.ReadCloser, error)
}

type MockGitHubClient struct {
    repositories map[string]*Repository
}
```

## Documentation

### Code Documentation

- Document all public functions and types
- Use Go doc conventions
- Include examples for complex functions

```go
// Install downloads and installs agents from the specified source.
// It returns an error if the source is invalid or installation fails.
//
// Example:
//   source := Source{Name: "example", Type: "github", Repository: "owner/repo"}
//   err := handler.Install(context.Background(), source)
func (h *Handler) Install(ctx context.Context, source Source) error {
    // implementation
}
```

### User Documentation

- Update relevant documentation for user-facing changes
- Add examples for new features
- Update CLI help text when adding commands/options

## Types of Contributions

### Bug Fixes

1. **Create an Issue** (if one doesn't exist)
2. **Reproduce the Bug** in tests
3. **Fix the Issue**
4. **Verify the Fix** with tests
5. **Update Documentation** if needed

### New Features

1. **Discuss the Feature** in an issue first
2. **Design API** if adding new interfaces
3. **Implement with Tests**
4. **Document the Feature**
5. **Update Examples**

### Source Handlers

To add a new source type:

1. **Implement SourceHandler Interface**

   ```go
   type MySourceHandler struct {
       config MySourceConfig
   }

   func (h *MySourceHandler) Install(ctx context.Context, source Source) error {
       // implementation
   }

   func (h *MySourceHandler) Update(ctx context.Context, source Source) error {
       // implementation
   }

   // ... other interface methods
   ```

2. **Register Handler**

   ```go
   func init() {
       RegisterSourceHandler("mysource", NewMySourceHandler)
   }
   ```

3. **Add Configuration Schema**

   Update configuration validation to support new fields.

4. **Add Tests**

   Comprehensive unit and integration tests.

5. **Document Usage**

   Add examples to configuration documentation.

### Transformations

To add a new transformation type:

1. **Implement Transformer Interface**

   ```go
   type MyTransformer struct{}

   func (t *MyTransformer) Transform(files []File, options map[string]interface{}) ([]File, error) {
       // transformation logic
   }
   ```

2. **Register Transformer**

   ```go
   func init() {
       RegisterTransformer("mytransform", NewMyTransformer)
   }
   ```

3. **Add Tests and Documentation**

### Documentation Improvements

- Fix typos and improve clarity
- Add missing examples
- Update outdated information
- Improve organization and navigation

## Testing Guidelines

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/installer/...

# Run with verbose output
go test -v ./internal/installer/

# Run specific test
go test -run TestInstaller_Install ./internal/installer/
```

### Test Categories

1. **Unit Tests**: Test individual functions/methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete user workflows
4. **Performance Tests**: Test scalability and performance

### Test Data

- Use `testdata/` directories for test fixtures
- Create realistic test scenarios
- Clean up test resources properly

```go
func TestWithTestData(t *testing.T) {
    testFile := filepath.Join("testdata", "config.yaml")
    config, err := LoadConfig(testFile)
    assert.NoError(t, err)
    // test using config
}
```

## Performance Considerations

### Benchmarking

Add benchmarks for performance-critical code:

```go
func BenchmarkInstaller_Install(b *testing.B) {
    installer := setupBenchmarkInstaller()
    source := createBenchmarkSource()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        installer.Install(context.Background(), source)
    }
}
```

### Memory Management

- Avoid loading large files entirely into memory
- Use streaming for large downloads
- Clean up resources properly

```go
func processLargeFile(r io.Reader) error {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        // process line by line
    }
    return scanner.Err()
}
```

## Security Guidelines

### Input Validation

Always validate user input:

```go
func ValidateSource(source Source) error {
    if source.Name == "" {
        return errors.New("source name is required")
    }

    if !isValidSourceType(source.Type) {
        return fmt.Errorf("invalid source type: %s", source.Type)
    }

    return nil
}
```

### Path Safety

Prevent directory traversal attacks:

```go
func SafePath(base, path string) (string, error) {
    full := filepath.Join(base, path)
    if !strings.HasPrefix(full, base) {
        return "", fmt.Errorf("unsafe path: %s", path)
    }
    return full, nil
}
```

### Credential Handling

- Never log credentials
- Load from environment variables
- Use secure storage when possible

```go
func getToken(envVar string) (string, error) {
    token := os.Getenv(envVar)
    if token == "" {
        return "", fmt.Errorf("token not found in %s", envVar)
    }
    return token, nil
}
```

## Release Process

### Version Numbering

We use semantic versioning (semver):

- `MAJOR.MINOR.PATCH`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes

### Creating Releases

1. **Update Version**

   Update version in relevant files.

2. **Update Changelog**

   Document all changes since last release.

3. **Create Tag**

   ```bash
   git tag -a v1.2.0 -m "Release version 1.2.0"
   git push origin v1.2.0
   ```

4. **Build Release Assets**

   ```bash
   make cross-compile
   ```

## Communication

### Issues

- Use issue templates when available
- Provide clear reproduction steps for bugs
- Include environment information

### Pull Requests

- Use PR template
- Link related issues
- Request review from maintainers
- Respond to feedback promptly

### Discussions

- Use GitHub Discussions for questions
- Be respectful and constructive
- Help others when possible

## Code Review

### As a Contributor

- Keep PRs focused and small
- Write clear commit messages
- Respond to feedback constructively
- Be patient during the review process

### As a Reviewer

- Be constructive and helpful
- Focus on code quality and maintainability
- Consider security and performance implications
- Approve when satisfied with changes

## Community Guidelines

- Be respectful and inclusive
- Follow the code of conduct
- Help newcomers get started
- Share knowledge and best practices

## Getting Help

- **Documentation**: Check existing docs first
- **Issues**: Search existing issues
- **Discussions**: Ask questions in GitHub Discussions
- **Code**: Use GitHub's code search

## Recognition

Contributors are recognized through:

- Commit attribution in Git history
- Changelog acknowledgments
- Contributor listings
- Special recognition for significant contributions

Thank you for contributing to Agent Manager! 🎉