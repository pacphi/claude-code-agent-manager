package conflict

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewResolver(t *testing.T) {
	strategy := "backup"
	backupDir := "/tmp/backups"

	resolver := NewResolver(strategy, backupDir)

	if resolver == nil {
		t.Fatal("Expected resolver but got nil")
	}

	if resolver.strategy != strategy {
		t.Errorf("Expected strategy %s, got %s", strategy, resolver.strategy)
	}

	if resolver.backupDir != backupDir {
		t.Errorf("Expected backupDir %s, got %s", backupDir, resolver.backupDir)
	}
}

func TestResolve(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "conflict-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	backupDir := filepath.Join(tempDir, "backups")
	resolver := NewResolver("backup", backupDir)

	existingFile := filepath.Join(tempDir, "existing.txt")
	newFile := filepath.Join(tempDir, "new.txt")

	if err := os.WriteFile(existingFile, []byte("existing content"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	if err := os.WriteFile(newFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	tests := []struct {
		name     string
		strategy string
		want     bool
		wantErr  bool
	}{
		{
			name:     "backup strategy",
			strategy: "backup",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "overwrite strategy",
			strategy: "overwrite",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "skip strategy",
			strategy: "skip",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "merge strategy",
			strategy: "merge",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "unknown strategy",
			strategy: "unknown",
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.Resolve(existingFile, newFile, tt.strategy)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateBackup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "create-backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	backupDir := filepath.Join(tempDir, "backups")
	resolver := NewResolver("backup", backupDir)

	sourceName := "test-source"
	err = resolver.CreateBackup(sourceName)
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("Failed to read backup directory: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 backup directory, got %d", len(entries))
	}

	backupPath := filepath.Join(backupDir, entries[0].Name())
	metadataPath := filepath.Join(backupPath, ".backup-info")

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("Backup metadata file not found")
	}
}

func TestCleanupOldBackups(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "old-backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	backupDir := filepath.Join(tempDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}

	resolver := NewResolver("backup", backupDir)

	oldBackupPath := filepath.Join(backupDir, "old-backup-20230101-120000")
	if err := os.MkdirAll(oldBackupPath, 0755); err != nil {
		t.Fatalf("Failed to create old backup dir: %v", err)
	}

	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldBackupPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old timestamp: %v", err)
	}

	recentBackupPath := filepath.Join(backupDir, "recent-backup-20230102-120000")
	if err := os.MkdirAll(recentBackupPath, 0755); err != nil {
		t.Fatalf("Failed to create recent backup dir: %v", err)
	}

	maxAge := 24 * time.Hour
	err = resolver.CleanupOldBackups(maxAge)
	if err != nil {
		t.Fatalf("CleanupOldBackups() error = %v", err)
	}

	if _, err := os.Stat(oldBackupPath); !os.IsNotExist(err) {
		t.Error("Old backup should have been removed")
	}

	if _, err := os.Stat(recentBackupPath); os.IsNotExist(err) {
		t.Error("Recent backup should not have been removed")
	}
}
