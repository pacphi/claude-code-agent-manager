package util

import (
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
