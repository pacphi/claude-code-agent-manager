package installer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace"
	"github.com/pacphi/claude-code-agent-manager/internal/progress"
	"github.com/pacphi/claude-code-agent-manager/internal/util"
)

// SourceHandler interface for different source types
type SourceHandler interface {
	Fetch(source config.Source, destDir string) (string, string, error)
	CheckUpdate(source config.Source, currentCommit string) (bool, string, error)
}

// GitHubHandler handles GitHub repositories
type GitHubHandler struct{}

// Fetch clones a GitHub repository
func (g *GitHubHandler) Fetch(source config.Source, destDir string) (string, string, error) {
	// Try using gh CLI first
	if commandExists("gh") {
		return g.fetchWithGH(source, destDir)
	}

	// Fall back to git
	gitURL := fmt.Sprintf("https://github.com/%s.git", source.Repository)
	gitSource := source
	gitSource.URL = gitURL

	handler := &GitHandler{}
	return handler.Fetch(gitSource, destDir)
}

func (g *GitHubHandler) fetchWithGH(source config.Source, destDir string) (string, string, error) {
	// Validate inputs
	if err := util.ValidateRepository(source.Repository); err != nil {
		return "", "", fmt.Errorf("invalid repository: %w", err)
	}
	if err := util.ValidateBranch(source.Branch); err != nil {
		return "", "", fmt.Errorf("invalid branch: %w", err)
	}
	if err := util.ValidatePath(destDir); err != nil {
		return "", "", fmt.Errorf("invalid destination directory: %w", err)
	}

	clonePath, err := util.SecureJoin(destDir, "repo")
	if err != nil {
		return "", "", fmt.Errorf("failed to create secure clone path: %w", err)
	}

	// Build gh command with validated arguments
	args := []string{"repo", "clone", source.Repository, clonePath}

	if source.Branch != "" && source.Branch != "main" {
		args = append(args, "--", "-b", source.Branch)
	}

	// Create secure command
	cmd, err := util.SecureCommand("gh", args...)
	if err != nil {
		return "", "", fmt.Errorf("failed to create secure command: %w", err)
	}

	// Set auth token if provided
	if source.Auth.TokenEnv != "" {
		token := os.Getenv(source.Auth.TokenEnv)
		if token != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GH_TOKEN=%s", token))
		}
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("gh clone failed: %s", output)
	}

	// Get commit hash
	commit, err := g.getCommitHash(clonePath)
	if err != nil {
		return "", "", err
	}

	// Return the source path within the clone
	sourcePath, err := util.SecureJoin(clonePath, source.Paths.Source)
	if err != nil {
		return "", "", fmt.Errorf("failed to create secure source path: %w", err)
	}
	return sourcePath, commit, nil
}

func (g *GitHubHandler) getCommitHash(repoPath string) (string, error) {
	// Validate repository path
	if err := util.ValidatePath(repoPath); err != nil {
		return "", fmt.Errorf("invalid repository path: %w", err)
	}

	// Create secure command
	cmd, err := util.SecureCommand("git", "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to create secure command: %w", err)
	}
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CheckUpdate checks if updates are available
func (g *GitHubHandler) CheckUpdate(source config.Source, currentCommit string) (bool, string, error) {
	// Create temp directory for checking
	tempDir, err := os.MkdirTemp("", "agent-update-check-*")
	if err != nil {
		return false, "", err
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			// Log error but don't fail the entire operation
			fmt.Printf("Warning: failed to remove temp directory %s: %v\n", tempDir, err)
		}
	}()

	// Fetch latest
	_, latestCommit, err := g.Fetch(source, tempDir)
	if err != nil {
		return false, "", err
	}

	hasUpdate := latestCommit != currentCommit
	return hasUpdate, latestCommit, nil
}

// GitHandler handles generic git repositories
type GitHandler struct{}

// Fetch clones a git repository
func (g *GitHandler) Fetch(source config.Source, destDir string) (string, string, error) {
	clonePath := filepath.Join(destDir, "repo")

	// Clone options
	cloneOpts := &git.CloneOptions{
		URL:      source.URL,
		Progress: nil, // Could set to os.Stdout for git progress, but we're using our own progress
	}

	// Set branch
	if source.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(source.Branch)
	}

	// Handle authentication securely
	if source.Auth.Method == "token" {
		token := os.Getenv(source.Auth.TokenEnv)
		if token != "" {
			// For HTTPS URLs with token auth, use proper auth methods instead of embedding in URL
			if strings.HasPrefix(source.URL, "https://") {
				// Use go-git's auth mechanisms instead of embedding token in URL
				// This prevents token exposure in logs and error messages
				cloneOpts.Auth = &http.BasicAuth{
					Username: "token", // GitHub uses "token" as username for token auth
					Password: token,
				}
			}
		}
	}

	// Clone repository
	repo, err := git.PlainClone(clonePath, false, cloneOpts)
	if err != nil {
		return "", "", fmt.Errorf("git clone failed: %w", err)
	}

	// Get HEAD commit
	ref, err := repo.Head()
	if err != nil {
		return "", "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit := ref.Hash().String()

	// Return the source path within the clone
	sourcePath := filepath.Join(clonePath, source.Paths.Source)
	return sourcePath, commit, nil
}

// CheckUpdate checks if updates are available
func (g *GitHandler) CheckUpdate(source config.Source, currentCommit string) (bool, string, error) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "agent-update-check-*")
	if err != nil {
		return false, "", err
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			// Log error but don't fail the entire operation
			fmt.Printf("Warning: failed to remove temp directory %s: %v\n", tempDir, err)
		}
	}()

	// Fetch latest
	_, latestCommit, err := g.Fetch(source, tempDir)
	if err != nil {
		return false, "", err
	}

	hasUpdate := latestCommit != currentCommit
	return hasUpdate, latestCommit, nil
}

// LocalHandler handles local file system sources
type LocalHandler struct{}

// Fetch copies from local file system
func (l *LocalHandler) Fetch(source config.Source, destDir string) (string, string, error) {
	sourcePath, err := expandPath(source.Paths.Source)
	if err != nil {
		return "", "", fmt.Errorf("failed to expand source path: %w", err)
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	// For local sources, we don't copy to temp, just return the source path
	// Generate a "commit" based on directory modification time
	info, err := os.Stat(sourcePath)
	if err != nil {
		return "", "", err
	}

	// Use modification time as version identifier
	commit := fmt.Sprintf("local-%d", info.ModTime().Unix())

	return sourcePath, commit, nil
}

// CheckUpdate checks if local source has been modified
func (l *LocalHandler) CheckUpdate(source config.Source, currentCommit string) (bool, string, error) {
	sourcePath, err := expandPath(source.Paths.Source)
	if err != nil {
		return false, "", fmt.Errorf("failed to expand source path: %w", err)
	}

	info, err := os.Stat(sourcePath)
	if err != nil {
		return false, "", err
	}

	newCommit := fmt.Sprintf("local-%d", info.ModTime().Unix())
	hasUpdate := newCommit != currentCommit

	return hasUpdate, newCommit, nil
}

// applyFilters filters files based on configuration
func (i *Installer) applyFilters(basePath string, filters config.FilterConfig) ([]string, error) {
	var result []string

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}

		// Check if file should be included
		if shouldInclude(relPath, info.Name(), filters) {
			result = append(result, relPath)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", basePath, err)
	}

	return result, nil
}

func shouldInclude(relPath, fileName string, filters config.FilterConfig) bool {
	// Check exclude patterns first
	if isExcluded(relPath, fileName, filters.Exclude.Patterns) {
		return false
	}

	// If no include filters, include everything not excluded
	if hasNoIncludeFilters(filters) {
		return true
	}

	// Check if file matches any include criteria
	return matchesIncludeCriteria(relPath, fileName, filters)
}

func isExcluded(relPath, fileName string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if matched, err := filepath.Match(pattern, fileName); err == nil && matched {
			return true
		}
		if matched, err := filepath.Match(pattern, relPath); err == nil && matched {
			return true
		}
	}
	return false
}

func hasNoIncludeFilters(filters config.FilterConfig) bool {
	return len(filters.Include.Extensions) == 0 &&
		len(filters.Include.Patterns) == 0 &&
		len(filters.Include.Regex) == 0
}

func matchesIncludeCriteria(relPath, fileName string, filters config.FilterConfig) bool {
	// Check extensions
	if matchesIncludeExtensions(fileName, filters.Include.Extensions) {
		return true
	}

	// Check patterns
	if matchesIncludePatterns(relPath, fileName, filters.Include.Patterns) {
		return true
	}

	// Check regex
	if matchesIncludeRegex(relPath, filters.Include.Regex) {
		return true
	}

	return false
}

func matchesIncludeExtensions(fileName string, extensions []string) bool {
	if len(extensions) == 0 {
		return false
	}

	ext := filepath.Ext(fileName)
	for _, allowedExt := range extensions {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

func matchesIncludePatterns(relPath, fileName string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, err := filepath.Match(pattern, fileName); err == nil && matched {
			return true
		}
		if matched, err := filepath.Match(pattern, relPath); err == nil && matched {
			return true
		}
	}
	return false
}

func matchesIncludeRegex(relPath string, regexPatterns []string) bool {
	for _, regexStr := range regexPatterns {
		re, err := regexp.Compile(regexStr)
		if err != nil {
			continue
		}
		if re.MatchString(relPath) {
			return true
		}
	}
	return false
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func expandPath(path string) (string, error) {
	return util.ExpandPath(path)
}

// SubagentsHandler handles subagents.sh marketplace
type SubagentsHandler struct {
	container *marketplace.Container
	config    *config.Config
}

func NewSubagentsHandler(cfg *config.Config) (*SubagentsHandler, error) {
	containerConfig := marketplace.ContainerConfig{
		BaseURL:         "https://subagents.sh",
		CacheEnabled:    true,
		CacheTTLHours:   1,
		CacheMaxSizeMB:  50,
		BrowserHeadless: true,
		BrowserTimeout:  30,
		UserAgent:       "agent-manager/1.0",
	}

	container, err := marketplace.NewContainer(containerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create marketplace scraper: %w", err)
	}

	return &SubagentsHandler{
		container: container,
		config:    cfg,
	}, nil
}

// Fetch implements SourceHandler interface
func (s *SubagentsHandler) Fetch(source config.Source, destDir string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Override container config if source has custom settings
	if source.Cache.Enabled || source.Cache.TTLHours > 0 || source.Cache.MaxSizeMB > 0 || source.MarketplaceURL != "" {
		containerConfig := marketplace.ContainerConfig{
			BaseURL:         source.MarketplaceURL,
			CacheEnabled:    source.Cache.Enabled,
			CacheTTLHours:   source.Cache.TTLHours,
			CacheMaxSizeMB:  int64(source.Cache.MaxSizeMB),
			BrowserHeadless: true,
			BrowserTimeout:  30,
			UserAgent:       "agent-manager/1.0",
		}

		if containerConfig.BaseURL == "" {
			containerConfig.BaseURL = "https://subagents.sh"
		}
		if containerConfig.CacheTTLHours == 0 {
			containerConfig.CacheTTLHours = 1 // Default
		}
		if containerConfig.CacheMaxSizeMB == 0 {
			containerConfig.CacheMaxSizeMB = 50 // Default
		}

		// Close existing container and create new one
		if s.container != nil {
			_ = s.container.Close()
		}

		var err error
		s.container, err = marketplace.NewContainer(containerConfig)
		if err != nil {
			return "", "", fmt.Errorf("failed to create custom container: %w", err)
		}
	}

	// Get marketplace data
	categories, err := s.container.Service.GetCategories(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch marketplace categories: %w", err)
	}

	// Get agents by category
	var agents []marketplace.Agent
	if category := source.Category; category != "" {
		// Get agents for specific category
		categoryAgents, err := s.container.Service.GetAgents(ctx, category)
		if err != nil {
			return "", "", fmt.Errorf("failed to fetch agents for category %s: %w", category, err)
		}
		agents = categoryAgents
	} else {
		// Get agents from all categories
		for _, cat := range categories {
			categoryAgents, err := s.container.Service.GetAgents(ctx, cat.Slug)
			if err != nil {
				continue // Skip categories that fail
			}
			agents = append(agents, categoryAgents...)
		}
	}

	if len(agents) == 0 {
		return "", "", fmt.Errorf("no agents found for the specified criteria")
	}

	// Create temporary directory structure
	sourcePath := filepath.Join(destDir, "agents")
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create source directory: %w", err)
	}

	// Download all agent content
	pm := progress.Default()
	progressID := fmt.Sprintf("subagents-%s", source.Name)
	if len(agents) > 1 {
		pm.StartProgress(progressID, "Downloading marketplace agents", len(agents))
		defer pm.FinishProgress(progressID, true, "")
	}

	for _, agent := range agents {
		content, err := s.container.Service.GetAgentContent(ctx, agent.ID)
		if err != nil {
			fmt.Printf("Warning: failed to get content for %s: %v\n", agent.Name, err)
			if len(agents) > 1 {
				pm.UpdateProgress(progressID, 1)
			}
			continue
		}

		// Write agent file
		filename := fmt.Sprintf("%s.md", agent.Slug)
		agentPath := filepath.Join(sourcePath, filename)

		// Format content with proper frontmatter
		formattedContent := s.formatAgentContent(agent, content)

		if err := os.WriteFile(agentPath, []byte(formattedContent), 0644); err != nil {
			return "", "", fmt.Errorf("failed to write agent %s: %w", agent.Name, err)
		}

		if len(agents) > 1 {
			pm.UpdateProgress(progressID, 1)
		}
	}

	// Generate version hash based on agents and timestamp
	versionHash := s.generateVersionHash(agents)

	return sourcePath, versionHash, nil
}

// CheckUpdate implements SourceHandler interface
func (s *SubagentsHandler) CheckUpdate(source config.Source, currentCommit string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	categories, err := s.container.Service.GetCategories(ctx)
	if err != nil {
		return false, "", fmt.Errorf("failed to check marketplace updates: %w", err)
	}

	// Get all agents for hash generation
	var agents []marketplace.Agent
	for _, cat := range categories {
		categoryAgents, err := s.container.Service.GetAgents(ctx, cat.Slug)
		if err != nil {
			continue // Skip categories that fail
		}
		agents = append(agents, categoryAgents...)
	}

	newHash := s.generateVersionHash(agents)
	hasUpdate := newHash != currentCommit

	return hasUpdate, newHash, nil
}

// Helper methods
func (s *SubagentsHandler) formatAgentContent(agent marketplace.Agent, content string) string {
	frontmatter := fmt.Sprintf(`---
name: %s
description: %s
category: %s
author: %s
rating: %.1f
downloads: %d
tags: %s
created_at: %s
updated_at: %s
source: subagents.sh
source_url: %s
---

%s`,
		agent.Name,
		agent.Description,
		agent.Category,
		agent.Author,
		agent.Rating,
		agent.Downloads,
		strings.Join(agent.Tags, ", "),
		agent.CreatedAt.Format("2006-01-02"),
		agent.UpdatedAt.Format("2006-01-02"),
		agent.ContentURL,
		content)

	return frontmatter
}

func (s *SubagentsHandler) generateVersionHash(agents []marketplace.Agent) string {
	// Generate a hash based on current timestamp and agent count
	now := time.Now()
	hashInput := fmt.Sprintf("%s-%d", now.Format("2006-01-02-15-04-05"), len(agents))

	// Use SHA256 for proper hashing
	hasher := sha256.New()
	hasher.Write([]byte(hashInput))
	hashBytes := hasher.Sum(nil)

	// Use first 12 characters of hex hash for readability
	hash := fmt.Sprintf("%x", hashBytes)
	if len(hash) > 12 {
		hash = hash[:12]
	}
	return fmt.Sprintf("subagents-%s", hash)
}
