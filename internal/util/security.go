package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidatePath checks for path traversal attacks and validates path format
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal detected: %s", path)
	}

	// Check for null bytes which can bypass security checks
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("null byte detected in path: %s", path)
	}

	// Normalize path for cross-platform comparison
	cleanPath := filepath.Clean(path)
	normalizedPath := strings.ToLower(strings.ReplaceAll(cleanPath, "\\", "/"))

	// Comprehensive dangerous patterns (normalized to forward slashes)
	dangerousPatterns := []string{
		"/etc/",
		"/proc/",
		"/sys/",
		"/dev/",
		"/boot/",
		"/root/",
		"/var/log/",
		"c:/windows/",
		"c:/program files/",
		"c:/program files (x86)/",
		"c:/users/all users/",
		"c:/documents and settings/",
	}

	for _, pattern := range dangerousPatterns {
		if strings.HasPrefix(normalizedPath, pattern) {
			return fmt.Errorf("access to system path denied: %s", path)
		}
	}

	return nil
}

// ValidateRepository validates repository names to prevent injection
func ValidateRepository(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository cannot be empty")
	}

	// Repository should only contain alphanumeric, hyphens, underscores, forward slashes, and dots
	validRepo := regexp.MustCompile(`^[a-zA-Z0-9\-_./]+$`)
	if !validRepo.MatchString(repo) {
		return fmt.Errorf("invalid repository format: %s", repo)
	}

	// Check for common injection patterns
	injectionPatterns := []string{
		";", "|", "&", "$(", "`", "&&", "||", ">>", "<<",
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(repo, pattern) {
			return fmt.Errorf("potential command injection in repository: %s", repo)
		}
	}

	return nil
}

// ValidateScriptPath validates that script paths are in allowed directories
func ValidateScriptPath(scriptPath string) error {
	if scriptPath == "" {
		return fmt.Errorf("script path cannot be empty")
	}

	// Clean and resolve the path
	cleanPath := filepath.Clean(scriptPath)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected in script path: %s", scriptPath)
	}

	// Define allowed script directories (configurable)
	allowedDirs := []string{
		"./scripts/",
		"scripts/",
		"/usr/local/bin/agent-manager-scripts/",
	}

	// Check if script is in an allowed directory
	isAllowed := false
	for _, dir := range allowedDirs {
		// Resolve absolute paths for comparison
		absDir, _ := filepath.Abs(dir)
		absScript, _ := filepath.Abs(cleanPath)

		if strings.HasPrefix(absScript, absDir) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		// Check if script exists in working directory's scripts folder
		cwd, _ := os.Getwd()
		localScriptsDir := filepath.Join(cwd, "scripts")
		absScript, _ := filepath.Abs(cleanPath)
		if strings.HasPrefix(absScript, localScriptsDir) {
			isAllowed = true
		}
	}

	if !isAllowed {
		return fmt.Errorf("script path not in allowed directories: %s", scriptPath)
	}

	// Verify the script exists and is executable
	info, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("script not found: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a script: %s", scriptPath)
	}

	return nil
}

// ValidateBranch validates git branch names
func ValidateBranch(branch string) error {
	if branch == "" {
		return nil // Empty branch is allowed (uses default)
	}

	// Git branch names have specific rules
	validBranch := regexp.MustCompile(`^[a-zA-Z0-9\-_./]+$`)
	if !validBranch.MatchString(branch) {
		return fmt.Errorf("invalid branch format: %s", branch)
	}

	// Check for injection attempts
	if strings.Contains(branch, "..") ||
		strings.Contains(branch, ";") ||
		strings.Contains(branch, "|") ||
		strings.Contains(branch, "&") {
		return fmt.Errorf("potential command injection in branch: %s", branch)
	}

	return nil
}

// SecureCommand creates a secure exec.Cmd with validated arguments
func SecureCommand(name string, args ...string) (*exec.Cmd, error) {
	// Validate command name
	if name == "" {
		return nil, fmt.Errorf("command name cannot be empty")
	}

	// Only allow specific known safe commands
	allowedCommands := map[string]bool{
		"git":  true,
		"gh":   true,
		"bash": true,
		"sh":   true,
	}

	if !allowedCommands[name] {
		return nil, fmt.Errorf("command not allowed: %s", name)
	}

	// Additional validation for script executors
	if name == "bash" || name == "sh" {
		// First argument should be the script path
		if len(args) > 0 {
			scriptPath := args[0]
			// Validate script path is within allowed directories
			if err := ValidateScriptPath(scriptPath); err != nil {
				return nil, fmt.Errorf("invalid script path: %w", err)
			}
		}
	}

	// Validate each argument
	for i, arg := range args {
		if err := validateCommandArg(arg); err != nil {
			return nil, fmt.Errorf("invalid argument %d: %w", i, err)
		}
	}

	// Create command with validated arguments
	cmd := exec.Command(name, args...)

	// Set secure environment - remove dangerous variables
	cmd.Env = getSecureEnv()

	return cmd, nil
}

// validateCommandArg validates individual command arguments
func validateCommandArg(arg string) error {
	// Check for null bytes
	if strings.Contains(arg, "\x00") {
		return fmt.Errorf("null byte detected in argument: %s", arg)
	}

	// Check for newline characters
	if strings.Contains(arg, "\n") || strings.Contains(arg, "\r") {
		return fmt.Errorf("newline character detected in argument: %s", arg)
	}

	// Check for environment variable expansion patterns
	if strings.Contains(arg, "${") || strings.Contains(arg, "$IFS") {
		return fmt.Errorf("environment variable expansion detected in argument: %s", arg)
	}

	// Check for path traversal patterns
	if strings.Contains(arg, "../") || strings.Contains(arg, "..\\") {
		return fmt.Errorf("path traversal detected in argument: %s", arg)
	}

	// Check for command injection patterns
	injectionPatterns := []string{
		";", "|", "&", "$(", "`", "&&", "||", ">>", "<<", "$(",
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(arg, pattern) {
			return fmt.Errorf("potential command injection in argument: %s", arg)
		}
	}

	return nil
}

// getSecureEnv returns a secure environment for command execution
func getSecureEnv() []string {
	// Define allowed environment variables
	allowedEnvVars := []string{
		"PATH", "HOME", "USER", "SHELL", "TERM",
		"GH_TOKEN", "GITHUB_TOKEN", // Git-specific tokens
	}

	var secureEnv []string
	for _, envVar := range allowedEnvVars {
		if value := getEnvVar(envVar); value != "" {
			secureEnv = append(secureEnv, fmt.Sprintf("%s=%s", envVar, value))
		}
	}

	return secureEnv
}

// getEnvVar safely gets environment variable
var getEnvVar = os.Getenv
