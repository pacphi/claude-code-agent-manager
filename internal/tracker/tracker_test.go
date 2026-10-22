package tracker

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	filePath := "/tmp/tracking.json"
	tracker := New(filePath)

	if tracker == nil {
		t.Fatal("Expected tracker but got nil")
	}

	if tracker.filePath != filePath {
		t.Errorf("Expected filePath %s, got %s", filePath, tracker.filePath)
	}
}

func TestRecordInstallation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tracker-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	trackingFile := filepath.Join(tempDir, "tracking.json")
	tracker := New(trackingFile)

	installation := Installation{
		SourceCommit: "commit123",
		Files: map[string]FileInfo{
			"test.md": {
				Path:     "test.md",
				Hash:     "hash123",
				Size:     100,
				Modified: time.Now(),
			},
		},
		Directories: []string{"dir1", "dir2"},
	}

	err = tracker.RecordInstallation("test-source", installation)
	if err != nil {
		t.Fatalf("RecordInstallation() error = %v", err)
	}

	if _, err := os.Stat(trackingFile); os.IsNotExist(err) {
		t.Error("Tracking file should have been created")
	}

	retrieved, err := tracker.GetInstallation("test-source")
	if err != nil {
		t.Fatalf("GetInstallation() error = %v", err)
	}

	if retrieved.SourceCommit != installation.SourceCommit {
		t.Errorf("Expected SourceCommit %s, got %s", installation.SourceCommit, retrieved.SourceCommit)
	}

	if len(retrieved.Files) != len(installation.Files) {
		t.Errorf("Expected %d files, got %d", len(installation.Files), len(retrieved.Files))
	}
}

func TestGetInstallation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tracker-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	trackingFile := filepath.Join(tempDir, "tracking.json")
	tracker := New(trackingFile)

	_, err = tracker.GetInstallation("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent installation")
	}

	installation := Installation{
		SourceCommit: "test-commit",
		Files:        map[string]FileInfo{},
	}

	err = tracker.RecordInstallation("test-source", installation)
	if err != nil {
		t.Fatalf("RecordInstallation() error = %v", err)
	}

	retrieved, err := tracker.GetInstallation("test-source")
	if err != nil {
		t.Fatalf("GetInstallation() error = %v", err)
	}

	if retrieved.SourceCommit != installation.SourceCommit {
		t.Errorf("Expected SourceCommit %s, got %s", installation.SourceCommit, retrieved.SourceCommit)
	}
}

func TestList(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tracker-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	trackingFile := filepath.Join(tempDir, "tracking.json")
	tracker := New(trackingFile)

	installations, err := tracker.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(installations) != 0 {
		t.Errorf("Expected 0 installations, got %d", len(installations))
	}

	sources := []string{"source1", "source2"}
	for _, source := range sources {
		installation := Installation{
			SourceCommit: source + "-commit",
			Files:        map[string]FileInfo{},
		}

		err = tracker.RecordInstallation(source, installation)
		if err != nil {
			t.Fatalf("RecordInstallation() error = %v", err)
		}
	}

	installations, err = tracker.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(installations) != len(sources) {
		t.Errorf("Expected %d installations, got %d", len(sources), len(installations))
	}
}

func TestBackupAndRestore(t *testing.T) {
	tempDir := t.TempDir()
	trackingFile := filepath.Join(tempDir, "tracking.json")
	tracker := New(trackingFile)

	if err := tracker.RecordInstallation("source-a", Installation{SourceCommit: "commit-a"}); err != nil {
		t.Fatalf("Failed to record installation: %v", err)
	}
	original, err := os.ReadFile(trackingFile)
	if err != nil {
		t.Fatalf("Failed to read tracking file: %v", err)
	}

	if err := tracker.Backup(); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}
	backups, err := filepath.Glob(trackingFile + ".backup.*")
	if err != nil || len(backups) != 1 {
		t.Fatalf("Expected one backup file, got %v (err %v)", backups, err)
	}
	backupContent, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	if string(backupContent) != string(original) {
		t.Error("Backup content does not match the tracking file")
	}
	if info, err := os.Stat(backups[0]); err != nil {
		t.Fatalf("Failed to stat backup: %v", err)
	} else if perm := info.Mode().Perm(); runtime.GOOS != "windows" && perm&0o077 != 0 {
		t.Errorf("Expected backup to be private (0600), got %v", perm)
	}

	if err := tracker.RecordInstallation("source-b", Installation{SourceCommit: "commit-b"}); err != nil {
		t.Fatalf("Failed to record second installation: %v", err)
	}

	if err := tracker.Restore(backups[0]); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if _, err := tracker.GetInstallation("source-b"); err == nil {
		t.Error("Expected source-b to be gone after restore")
	}
	restored, err := tracker.GetInstallation("source-a")
	if err != nil {
		t.Fatalf("Expected source-a after restore: %v", err)
	}
	if restored.SourceCommit != "commit-a" {
		t.Errorf("Expected commit-a, got %s", restored.SourceCommit)
	}
	if _, err := os.Stat(trackingFile + ".tmp"); !os.IsNotExist(err) {
		t.Error("Expected no leftover temp file after restore")
	}

	invalid := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(invalid, []byte("not json"), 0600); err != nil {
		t.Fatalf("Failed to write invalid backup: %v", err)
	}
	if err := tracker.Restore(invalid); err == nil {
		t.Error("Expected restore of invalid JSON to fail")
	}
}
