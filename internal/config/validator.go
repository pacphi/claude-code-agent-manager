package config

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Validate checks if the configuration is valid
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate version
	if cfg.Version != "1.0" {
		return fmt.Errorf("unsupported configuration version: %s", cfg.Version)
	}

	// Validate settings
	if err := validateSettings(&cfg.Settings); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	// Validate sources
	if len(cfg.Sources) == 0 {
		return fmt.Errorf("no sources defined")
	}

	sourceNames := make(map[string]bool)
	for i, source := range cfg.Sources {
		if err := validateSource(&source); err != nil {
			return fmt.Errorf("invalid source[%d] '%s': %w", i, source.Name, err)
		}

		// Check for duplicate names
		if sourceNames[source.Name] {
			return fmt.Errorf("duplicate source name: %s", source.Name)
		}
		sourceNames[source.Name] = true
	}

	// Validate metadata
	if err := validateMetadata(&cfg.Metadata); err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}

	return nil
}

func validateSettings(settings *Settings) error {
	// Validate conflict strategy
	validStrategies := []string{"backup", "overwrite", "skip", "merge"}
	if !contains(validStrategies, settings.ConflictStrategy) {
		return fmt.Errorf("invalid conflict strategy: %s (must be one of: %s)",
			settings.ConflictStrategy, strings.Join(validStrategies, ", "))
	}

	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, settings.LogLevel) {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)",
			settings.LogLevel, strings.Join(validLevels, ", "))
	}

	// Validate concurrent downloads
	if settings.ConcurrentDownloads < 1 || settings.ConcurrentDownloads > 10 {
		return fmt.Errorf("concurrent_downloads must be between 1 and 10")
	}

	// Validate timeout
	if settings.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	return nil
}

func validateSource(source *Source) error {
	// Validate basic source properties
	if err := validateSourceBasics(source); err != nil {
		return err
	}

	// Type-specific validation
	if err := validateSourceType(source); err != nil {
		return err
	}

	// Validate authentication
	if err := validateSourceAuth(source); err != nil {
		return err
	}

	// Validate other source components
	return validateSourceComponents(source)
}

func validateSourceBasics(source *Source) error {
	// Validate name
	if source.Name == "" {
		return fmt.Errorf("source name is required")
	}

	// Validate source type
	validTypes := []string{"github", "git", "local", "subagents"}
	if !contains(validTypes, source.Type) {
		return fmt.Errorf("invalid source type: %s (must be one of: %s)",
			source.Type, strings.Join(validTypes, ", "))
	}

	// Validate paths
	if source.Paths.Target == "" {
		return fmt.Errorf("target path is required")
	}

	return nil
}

func validateSourceType(source *Source) error {
	switch source.Type {
	case "github":
		if source.Repository == "" {
			return fmt.Errorf("repository is required for github source")
		}
		// Validate repository format (owner/repo)
		if !regexp.MustCompile(`^[^/]+/[^/]+$`).MatchString(source.Repository) {
			return fmt.Errorf("invalid github repository format (expected: owner/repo)")
		}

	case "git":
		if source.URL == "" {
			return fmt.Errorf("url is required for git source")
		}
		// Validate URL
		if _, err := url.Parse(source.URL); err != nil {
			return fmt.Errorf("invalid git URL: %w", err)
		}

	case "local":
		if source.Paths.Source == "" {
			return fmt.Errorf("source path is required for local source")
		}
	}

	return nil
}

func validateSourceAuth(source *Source) error {
	if source.Auth.Method == "" {
		return nil
	}

	validMethods := []string{"token", "ssh"}
	if !contains(validMethods, source.Auth.Method) {
		return fmt.Errorf("invalid auth method: %s", source.Auth.Method)
	}

	if source.Auth.Method == "token" && source.Auth.TokenEnv == "" {
		return fmt.Errorf("token_env is required for token auth")
	}

	if source.Auth.Method == "ssh" && source.Auth.SSHKey == "" {
		return fmt.Errorf("ssh_key is required for ssh auth")
	}

	return nil
}

func validateSourceComponents(source *Source) error {
	// Validate filters
	if err := validateFilters(&source.Filters); err != nil {
		return fmt.Errorf("invalid filters: %w", err)
	}

	// Validate transformations
	for i, transform := range source.Transformations {
		if err := validateTransformation(&transform); err != nil {
			return fmt.Errorf("invalid transformation[%d]: %w", i, err)
		}
	}

	// Validate post-install actions
	for i, action := range source.PostInstall {
		if err := validatePostInstall(&action); err != nil {
			return fmt.Errorf("invalid post_install[%d]: %w", i, err)
		}
	}

	// Validate conflict strategy override
	if source.ConflictStrategy != "" {
		validStrategies := []string{"backup", "overwrite", "skip", "merge"}
		if !contains(validStrategies, source.ConflictStrategy) {
			return fmt.Errorf("invalid conflict strategy override: %s", source.ConflictStrategy)
		}
	}

	return nil
}

func validateFilters(filters *FilterConfig) error {
	// Validate regex patterns
	for _, pattern := range filters.Include.Regex {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
		}
	}

	// Validate glob patterns
	for _, pattern := range filters.Include.Patterns {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
		}
	}

	for _, pattern := range filters.Exclude.Patterns {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("invalid exclude pattern '%s': %w", pattern, err)
		}
	}

	return nil
}

func validateTransformation(transform *Transformation) error {
	if transform.Type == "" {
		return fmt.Errorf("transformation type is required")
	}

	validTypes := []string{
		"remove_numeric_prefix",
		"extract_docs",
		"rename_files",
		"replace_content",
		"custom_script",
	}

	if !contains(validTypes, transform.Type) {
		return fmt.Errorf("invalid transformation type: %s", transform.Type)
	}

	// Type-specific validation
	switch transform.Type {
	case "remove_numeric_prefix":
		if transform.Pattern != "" {
			if _, err := regexp.Compile(transform.Pattern); err != nil {
				return fmt.Errorf("invalid pattern: %w", err)
			}
		}

	case "extract_docs":
		if transform.TargetDir == "" {
			return fmt.Errorf("target_dir is required for extract_docs")
		}
		if transform.Naming != "" {
			validNaming := []string{"UPPERCASE_UNDERSCORE", "lowercase_dash", "CamelCase"}
			if !contains(validNaming, transform.Naming) {
				return fmt.Errorf("invalid naming strategy: %s", transform.Naming)
			}
		}

	case "custom_script":
		if transform.Script == "" {
			return fmt.Errorf("script path is required for custom_script")
		}
	}

	return nil
}

func validatePostInstall(action *PostInstall) error {
	if action.Type == "" {
		return fmt.Errorf("post-install action type is required")
	}

	validTypes := []string{"script", "command"}
	if !contains(validTypes, action.Type) {
		return fmt.Errorf("invalid post-install type: %s", action.Type)
	}

	if action.Path == "" {
		return fmt.Errorf("path is required for post-install action")
	}

	return nil
}

func validateMetadata(metadata *Metadata) error {
	if metadata.TrackingFile == "" {
		return fmt.Errorf("tracking_file is required")
	}

	if metadata.LogFile == "" {
		return fmt.Errorf("log_file is required")
	}

	// Check if parent directories can be created
	trackingDir := filepath.Dir(metadata.TrackingFile)
	if trackingDir != "." && trackingDir != "/" {
		// Just check if it's a valid path format
		if filepath.Clean(trackingDir) != trackingDir {
			return fmt.Errorf("invalid tracking file path: %s", metadata.TrackingFile)
		}
	}

	return nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			// Return original path if home directory can't be determined
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
