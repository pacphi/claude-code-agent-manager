# Release Guide

This guide provides comprehensive documentation for creating releases of Agent Manager, including both automated and manual processes.

## Quick Release Process

For maintainers who need to create a release quickly:

```bash
# 1. Prepare and validate
make test && make lint && make vet
git status  # Ensure clean working directory

# 2. Create and push tag (triggers automated release)
git tag v1.2.3
git push origin v1.2.3

# 3. Monitor GitHub Actions and verify release
gh workflow list
gh release list
```

The GitHub Actions workflow will automatically build cross-platform binaries and create the GitHub release.

## Release Types

### Semantic Versioning

Agent Manager follows [semantic versioning](https://semver.org/):

- **Major version** (`v2.0.0`): Breaking changes
- **Minor version** (`v1.1.0`): New features, backward compatible
- **Patch version** (`v1.0.1`): Bug fixes, backward compatible
- **Pre-release** (`v1.0.0-alpha.1`, `v1.0.0-beta.2`, `v1.0.0-rc.1`): Testing versions

### Release Types

- **Stable Release**: Production-ready (e.g., `v1.0.0`)
- **Pre-release**: Testing version (e.g., `v1.0.0-beta.1`)
- **Hotfix**: Critical bug fix (e.g., `v1.0.1`)

## Automated Release Process

### GitHub Actions Workflow

The project uses a tag-based release workflow (`.github/workflows/release.yml`):

**Trigger**: Push of tags matching `v*` pattern
**Platforms**: Linux, macOS, Windows (AMD64 + ARM64)
**Artifacts**: Cross-compiled binaries, checksums, release notes

### Workflow Steps

1. **Checkout**: Fetch source code with full history
2. **Setup**: Configure Go 1.24.11 with module caching
3. **Test**: Run full test suite
4. **Build**: Cross-compile for all platforms using `make cross-compile`
5. **Checksum**: Generate SHA256SUMS for all binaries
6. **Release Notes**: Auto-generate from CHANGELOG.md or git commits
7. **Publish**: Create GitHub release with artifacts

### Artifacts Produced

- `agent-manager-darwin-amd64` - macOS Intel
- `agent-manager-darwin-arm64` - macOS Apple Silicon
- `agent-manager-linux-amd64` - Linux Intel/AMD
- `agent-manager-linux-arm64` - Linux ARM
- `agent-manager-windows-amd64.exe` - Windows Intel/AMD
- `SHA256SUMS` - Checksums for verification

## Manual Release Process

### Prerequisites

Ensure you have the required tools:

```bash
# Required
go version          # Go 1.24.11+
git --version       # Git for tagging
gh --version        # GitHub CLI for monitoring

# Optional but recommended
make --version      # For convenience targets
golangci-lint --version  # For pre-release linting
```

### Step-by-Step Process

#### 1. Pre-Release Preparation

**Verify clean state:**

```bash
git status
git pull origin main
```

**Run full validation:**

```bash
# Run tests
make test

# Check code quality
make fmt
make vet
make lint  # Optional, runs in CI

# Verify build works
make clean && make build
```

**Test cross-compilation:**

```bash
make cross-compile
ls -la bin/
```

#### 2. Version Planning

**Check current version:**

```bash
git tag --list --sort=-version:refname | head -5
```

**Determine next version** based on changes:

- Review recent commits: `git log --oneline $(git describe --tags --abbrev=0)..HEAD`
- Check for breaking changes, new features, or bug fixes
- Follow semantic versioning principles

#### 3. Create Release

**Create annotated tag:**

```bash
# For stable release
git tag -a v1.2.3 -m "Release v1.2.3"

# For pre-release
git tag -a v1.2.3-beta.1 -m "Release v1.2.3-beta.1"
```

**Push tag to trigger workflow:**

```bash
git push origin v1.2.3
```

#### 4. Monitor Release Process

**Watch GitHub Actions:**

```bash
# List recent workflow runs
gh run list --limit 5

# Watch specific run (get ID from above)
gh run watch <run-id>

# View workflow logs if needed
gh run view <run-id> --log
```

**Verify release creation:**

```bash
# List releases
gh release list

# View specific release
gh release view v1.2.3
```

#### 5. Post-Release Verification

**Test release artifacts:**

```bash
# Download and test a binary
gh release download v1.2.3 --pattern "*linux-amd64*"
chmod +x agent-manager-linux-amd64
./agent-manager-linux-amd64 version
```

**Verify checksums:**

```bash
gh release download v1.2.3 --pattern "SHA256SUMS"
sha256sum -c SHA256SUMS
```

**Update documentation** if needed:

- README.md version references
- Installation instructions
- CHANGELOG.md (if maintained manually)

## Emergency Procedures

### Hotfix Release

For critical bugs in production:

```bash
# Create hotfix branch from tag
git checkout -b hotfix/v1.0.1 v1.0.0

# Make minimal fix
# ... edit files ...

# Test thoroughly
make test

# Commit and tag
git commit -m "fix: critical security issue"
git tag -a v1.0.1 -m "Hotfix v1.0.1: Security fix"
git push origin hotfix/v1.0.1
git push origin v1.0.1

# Merge back to main
git checkout main
git merge hotfix/v1.0.1
git push origin main
```

### Release Rollback

If a release has critical issues:

```bash
# Mark release as pre-release (reduces visibility)
gh release edit v1.2.3 --prerelease

# Or create replacement release
git tag -a v1.2.4 -m "Release v1.2.4: Fixes issues in v1.2.3"
git push origin v1.2.4
```

### Failed Release Recovery

If GitHub Actions fails:

```bash
# Check workflow status
gh run list --workflow=release.yml

# Re-run failed workflow
gh run rerun <run-id>

# Or delete tag and recreate
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3
# Fix issues, then recreate tag
```

## Release Checklist

### Pre-Release

- [ ] Clean working directory (`git status`)
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] Vet checks pass (`make vet`)
- [ ] Cross-compilation works (`make cross-compile`)
- [ ] Version number follows semantic versioning
- [ ] Release notes prepared (if using manual CHANGELOG.md)

### Release

- [ ] Annotated tag created with descriptive message
- [ ] Tag pushed to origin
- [ ] GitHub Actions workflow triggered successfully
- [ ] All platform builds completed
- [ ] Release artifacts uploaded
- [ ] Release notes generated/attached

### Post-Release

- [ ] Release artifacts tested on at least one platform
- [ ] Checksums verified
- [ ] Release marked as stable (not pre-release) if appropriate
- [ ] Documentation updated if needed
- [ ] Team notified of new release

## Troubleshooting

### Common Issues

**Build fails in CI:**

- Check Go version compatibility
- Verify all dependencies are available
- Test cross-compilation locally first

**Tag already exists:**

```bash
# Delete local and remote tag
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3
```

**Release artifacts missing:**

- Check GitHub Actions workflow logs
- Verify workflow permissions (needs `contents: write`)
- Ensure all steps completed successfully

**Wrong version in binary:**

- Version is set via git tags in LDFLAGS
- Verify tag format matches `v*` pattern
- Check build logs for version injection

### Getting Help

- View workflow logs: `gh run view <run-id> --log`
- Check release status: `gh release view v1.2.3`
- Review build artifacts: `gh run download <run-id>`
- Validate configuration: `./bin/agent-manager validate`

## Integration with Build System

### Makefile Targets

Agent Manager's `Makefile` provides several release-related targets:

```bash
make cross-compile  # Build for all platforms
make release        # Create release artifacts (tarballs)
make clean          # Remove build artifacts
make test           # Run test suite
make validate       # Validate configuration
```

### Build Configuration

The build process uses these key variables:

- `VERSION`: Derived from `git describe --tags --always --dirty`
- `BUILD_TIME`: Current timestamp
- `LDFLAGS`: Injects version info into binary

### Documentation Links

- **[Build Guide](BUILD.md)**: Detailed build instructions
- **[CLI Reference](../reference/CLI-REFERENCE.md)**: Command examples and workflows
- **[Configuration Reference](../guides/CONFIGURATION.md)**: YAML configuration docs
- **[Makefile](../../Makefile)**: Build targets and variables

## Best Practices

### Version Management

- Use annotated tags for all releases
- Include meaningful commit messages in tag annotations
- Follow semantic versioning consistently
- Test release candidates before stable releases

### Quality Assurance

- Never skip tests for releases
- Validate cross-platform builds locally when possible
- Use pre-releases for significant changes
- Maintain backward compatibility in minor/patch releases

### Communication

- Coordinate releases with team members
- Document breaking changes clearly
- Provide migration guides for major versions
- Announce releases through appropriate channels

### Security

- Verify checksums of release artifacts
- Use signed tags for security-critical releases
- Review dependencies for vulnerabilities before release
- Follow responsible disclosure for security fixes