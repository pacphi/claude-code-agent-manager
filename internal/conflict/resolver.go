package conflict

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/epiclabs-io/diff3"
	"github.com/pacphi/claude-code-agent-manager/internal/util"
)

// Resolver handles file conflict resolution
type Resolver struct {
	strategy  string
	backupDir string
}

// NewResolver creates a new conflict resolver
func NewResolver(strategy, backupDir string) *Resolver {
	return &Resolver{
		strategy:  strategy,
		backupDir: backupDir,
	}
}

// Resolve resolves a file conflict based on the configured strategy
func (r *Resolver) Resolve(existingPath, newPath, strategy string) (bool, error) {
	// Use override strategy if provided
	if strategy == "" {
		strategy = r.strategy
	}

	switch strategy {
	case "backup":
		return r.resolveWithBackup(existingPath, newPath)
	case "overwrite":
		return true, nil // Allow overwrite
	case "skip":
		return false, nil // Skip the file
	case "merge":
		return r.resolveWithMerge(existingPath, newPath)
	default:
		return false, fmt.Errorf("unknown conflict strategy: %s", strategy)
	}
}

// resolveWithBackup creates a backup of the existing file
func (r *Resolver) resolveWithBackup(existingPath, newPath string) (bool, error) {
	// Create backup directory if it doesn't exist
	backupPath := r.getBackupPath(existingPath)
	backupDir := filepath.Dir(backupPath)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy existing file to backup
	if err := r.copyFile(existingPath, backupPath); err != nil {
		return false, fmt.Errorf("failed to backup file: %w", err)
	}

	return true, nil // Allow overwrite after backup
}

// resolveWithMerge attempts to merge files using three-way merge
func (r *Resolver) resolveWithMerge(existingPath, newPath string) (bool, error) {
	// First create a backup like in backup strategy
	backupPath := r.getBackupPath(existingPath)
	backupDir := filepath.Dir(backupPath)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy existing file to backup
	if err := r.copyFile(existingPath, backupPath); err != nil {
		return false, fmt.Errorf("failed to backup file: %w", err)
	}

	// Now attempt three-way merge
	mergedContent, err := r.performThreeWayMerge(backupPath, existingPath, newPath)
	if err != nil {
		// If merge fails, fall back to backup strategy behavior
		return true, nil
	}

	// Write merged content to the existing path
	if err := os.WriteFile(existingPath, mergedContent, 0644); err != nil {
		return false, fmt.Errorf("failed to write merged content: %w", err)
	}

	return true, nil
}

// performThreeWayMerge performs intelligent merge of three file versions
func (r *Resolver) performThreeWayMerge(originalPath, currentPath, incomingPath string) ([]byte, error) {
	// Read the three versions
	originalFile, err := os.Open(originalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read original file: %w", err)
	}
	defer originalFile.Close()

	currentFile, err := os.Open(currentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read current file: %w", err)
	}
	defer currentFile.Close()

	incomingFile, err := os.Open(incomingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read incoming file: %w", err)
	}
	defer incomingFile.Close()

	// Perform three-way merge using diff3
	// Parameters: current (a), original (o), incoming (b)
	result, err := diff3.Merge(currentFile, originalFile, incomingFile, true, "Current", "Incoming")
	if err != nil {
		return nil, fmt.Errorf("merge failed: %w", err)
	}

	// Read the merged result
	mergedContent, err := io.ReadAll(result.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to read merge result: %w", err)
	}

	return mergedContent, nil
}

// CreateBackup creates a backup of all files for a source
func (r *Resolver) CreateBackup(sourceName string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s-%s", sourceName, timestamp)
	backupPath := filepath.Join(r.backupDir, backupName)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Store backup metadata
	metadataPath := filepath.Join(backupPath, ".backup-info")
	metadata := fmt.Sprintf("Source: %s\nTimestamp: %s\n", sourceName, timestamp)

	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
		return fmt.Errorf("failed to write backup metadata: %w", err)
	}

	return nil
}

// RestoreBackup restores files from a backup (legacy directory-based backups)
func (r *Resolver) RestoreBackup(sourceName string) error {
	// Find the most recent backup for the source
	backupDir := r.findLatestBackup(sourceName)
	if backupDir == "" {
		return fmt.Errorf("no backup found for source: %s", sourceName)
	}

	// Read backup metadata
	metadataPath := filepath.Join(backupDir, ".backup-info")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return fmt.Errorf("backup metadata not found")
	}

	// Restore files from backup
	err := filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip metadata file and directories
		if info.IsDir() || filepath.Base(path) == ".backup-info" {
			return nil
		}

		// Calculate relative path and restore location
		relPath, err := filepath.Rel(backupDir, path)
		if err != nil {
			return err
		}

		// Restore file
		restorePath := filepath.Join(".", relPath)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(restorePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Copy file back
		if err := r.copyFile(path, restorePath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", relPath, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to restore from backup %s: %w", backupDir, err)
	}
	return nil
}

// RestoreBackupFiles restores files from flat file backups
func (r *Resolver) RestoreBackupFiles() error {
	if r.backupDir == "" {
		return nil
	}

	// Read all backup files
	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Find the most recent timestamp
	var latestTimestamp string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Extract timestamp from filename (format: path_timestamp)
		name := entry.Name()
		underscorePos := strings.LastIndex(name, "_")
		if underscorePos > 0 {
			timestamp := name[underscorePos+1:]
			if timestamp > latestTimestamp {
				latestTimestamp = timestamp
			}
		}
	}

	if latestTimestamp == "" {
		return nil // No backup files found
	}

	// Restore all files with the latest timestamp
	restoredCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, "_"+latestTimestamp) {
			continue
		}

		// Remove timestamp to get the flattened path
		flatPath := strings.TrimSuffix(name, "_"+latestTimestamp)

		// Reconstruct original path under .claude/agents/
		// Replace underscores with slashes to restore directory structure
		relativePath := strings.ReplaceAll(flatPath, "_", "/")
		originalPath := filepath.Join(".claude", "agents", relativePath)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(originalPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", originalPath, err)
		}

		// Copy backup file to original location
		backupPath := filepath.Join(r.backupDir, name)
		if err := r.copyFile(backupPath, originalPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", originalPath, err)
		}

		restoredCount++
	}

	if restoredCount > 0 {
		fmt.Printf("Restored %d files from backup\n", restoredCount)
	}

	return nil
}

// RestoreBackupFilesWithTracking restores files from flat file backups and returns which files were restored
func (r *Resolver) RestoreBackupFilesWithTracking() (map[string]bool, error) {
	restoredFiles := make(map[string]bool)

	if r.backupDir == "" {
		return restoredFiles, nil
	}

	// Read all backup files
	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return restoredFiles, nil
		}
		return restoredFiles, fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Find the most recent timestamp
	var latestTimestamp string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Extract timestamp from filename (format: path_timestamp)
		name := entry.Name()
		underscorePos := strings.LastIndex(name, "_")
		if underscorePos > 0 {
			timestamp := name[underscorePos+1:]
			if timestamp > latestTimestamp {
				latestTimestamp = timestamp
			}
		}
	}

	if latestTimestamp == "" {
		return restoredFiles, nil // No backup files found
	}

	// Restore all files with the latest timestamp
	restoredCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, "_"+latestTimestamp) {
			continue
		}

		// Remove timestamp to get the flattened path
		flatPath := strings.TrimSuffix(name, "_"+latestTimestamp)

		// Reconstruct original path under .claude/agents/
		// Replace underscores with slashes to restore directory structure
		relativePath := strings.ReplaceAll(flatPath, "_", "/")
		originalPath := filepath.Join(".claude", "agents", relativePath)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(originalPath), 0755); err != nil {
			return restoredFiles, fmt.Errorf("failed to create directory for %s: %w", originalPath, err)
		}

		// Copy backup file to original location
		backupPath := filepath.Join(r.backupDir, name)
		if err := r.copyFile(backupPath, originalPath); err != nil {
			return restoredFiles, fmt.Errorf("failed to restore %s: %w", originalPath, err)
		}

		// Track restored file
		restoredFiles[originalPath] = true
		restoredCount++
	}

	if restoredCount > 0 {
		fmt.Printf("Restored %d files from backup\n", restoredCount)
	}

	return restoredFiles, nil
}

// CleanupBackups removes backups for a specific source (legacy directory-based)
func (r *Resolver) CleanupBackups(sourceName string) error {
	if r.backupDir == "" {
		return nil
	}

	// Try new flat file cleanup first
	if err := r.CleanupBackupFiles(); err != nil {
		return err
	}

	// Also check for legacy directory-based backups
	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	prefix := sourceName + "-"
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			backupPath := filepath.Join(r.backupDir, entry.Name())
			if err := os.RemoveAll(backupPath); err != nil {
				return fmt.Errorf("failed to remove backup %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// CleanupBackupFiles removes flat file backups
func (r *Resolver) CleanupBackupFiles() error {
	if r.backupDir == "" {
		return nil
	}

	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Remove all backup files (they have underscore followed by timestamp)
	removedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Check if this looks like a backup file (contains underscore followed by timestamp)
		if strings.Contains(name, "_") {
			underscorePos := strings.LastIndex(name, "_")
			if underscorePos > 0 {
				timestamp := name[underscorePos+1:]
				// Simple check: timestamp should be 15 chars (20060102-150405)
				if len(timestamp) == 15 && strings.Contains(timestamp, "-") {
					backupPath := filepath.Join(r.backupDir, name)
					if err := os.Remove(backupPath); err != nil {
						return fmt.Errorf("failed to remove backup %s: %w", name, err)
					}
					removedCount++
				}
			}
		}
	}

	if removedCount > 0 {
		fmt.Printf("Removed %d backup files\n", removedCount)
	}

	return nil
}

// CleanupOldBackups removes backups older than the specified duration
func (r *Resolver) CleanupOldBackups(maxAge time.Duration) error {
	if r.backupDir == "" {
		return nil
	}

	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cutoff := time.Now().Add(-maxAge)

	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				backupPath := filepath.Join(r.backupDir, entry.Name())
				if err := os.RemoveAll(backupPath); err != nil {
					return fmt.Errorf("failed to remove old backup %s: %w", entry.Name(), err)
				}
			}
		}
	}

	return nil
}

// Private helper methods

func (r *Resolver) getBackupPath(originalPath string) string {
	timestamp := time.Now().Format("20060102-150405")

	// Clean the original path
	cleanedPath := filepath.Clean(originalPath)

	// Check if this is a file under .claude/agents/
	agentsPrefix := ".claude/agents/"
	if strings.HasPrefix(cleanedPath, agentsPrefix) {
		// Extract relative path from .claude/agents/
		relativePath := strings.TrimPrefix(cleanedPath, agentsPrefix)

		// Replace path separators with underscores to create flat backup filename
		// Example: "foo/agent.md" becomes "foo_agent.md"
		flatPath := strings.ReplaceAll(relativePath, "/", "_")
		backupName := fmt.Sprintf("%s_%s", flatPath, timestamp)

		return filepath.Join(r.backupDir, backupName)
	}

	// Fallback for files not in .claude/agents/ - just use filename
	filename := filepath.Base(cleanedPath)
	backupName := fmt.Sprintf("%s_%s", filename, timestamp)
	return filepath.Join(r.backupDir, backupName)
}

func (r *Resolver) findLatestBackup(sourceName string) string {
	if r.backupDir == "" {
		return ""
	}

	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		return ""
	}

	prefix := sourceName + "-"
	var latestBackup string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestBackup = filepath.Join(r.backupDir, entry.Name())
			}
		}
	}

	return latestBackup
}

func (r *Resolver) copyFile(src, dst string) error {
	fm := util.NewFileManager()
	return fm.Copy(src, dst)
}

// Removed custom hasPrefix - using strings.HasPrefix instead

// BackupInfo contains information about a backup
type BackupInfo struct {
	Source    string
	Timestamp time.Time
	Path      string
	Size      int64
}

// ListBackups returns information about all backups
func (r *Resolver) ListBackups() ([]BackupInfo, error) {
	if r.backupDir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var backups []BackupInfo

	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Parse backup name (format: sourcename-timestamp)
			name := entry.Name()
			lastDash := -1
			for i := len(name) - 1; i >= 0; i-- {
				if name[i] == '-' {
					lastDash = i
					break
				}
			}

			if lastDash > 0 {
				sourceName := name[:lastDash]

				// Calculate backup size
				var size int64
				backupPath := filepath.Join(r.backupDir, name)
				if walkErr := filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() {
						size += info.Size()
					}
					return nil
				}); walkErr != nil {
					// Log error but continue with size calculation
					fmt.Printf("Warning: failed to calculate size for backup %s: %v\n", name, walkErr)
				}

				backups = append(backups, BackupInfo{
					Source:    sourceName,
					Timestamp: info.ModTime(),
					Path:      backupPath,
					Size:      size,
				})
			}
		}
	}

	return backups, nil
}
