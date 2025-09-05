package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	securejoin "github.com/cyphar/filepath-securejoin"
)

// FileManager provides secure file operations
type FileManager struct{}

// NewFileManager creates a new FileManager instance
func NewFileManager() *FileManager {
	return &FileManager{}
}

// Copy safely copies a file from src to dst with validation
func (fm *FileManager) Copy(src, dst string) error {
	// Validate both paths
	if err := ValidatePath(src); err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	if err := ValidatePath(dst); err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// Clean paths
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close source file: %v\n", closeErr)
		}
	}()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Use atomic file copy: write to temp file first
	tempPath := dst + ".tmp"
	dstFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		if closeErr := dstFile.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close temp file during cleanup: %v\n", closeErr)
		}
		if removeErr := os.Remove(tempPath); removeErr != nil {
			fmt.Printf("Warning: failed to remove temp file during cleanup: %v\n", removeErr)
		}
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Set same permissions as source
	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		if closeErr := dstFile.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close temp file during cleanup: %v\n", closeErr)
		}
		if removeErr := os.Remove(tempPath); removeErr != nil {
			fmt.Printf("Warning: failed to remove temp file during cleanup: %v\n", removeErr)
		}
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Force sync to disk
	if err := dstFile.Sync(); err != nil {
		if closeErr := dstFile.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close temp file during cleanup: %v\n", closeErr)
		}
		if removeErr := os.Remove(tempPath); removeErr != nil {
			fmt.Printf("Warning: failed to remove temp file during cleanup: %v\n", removeErr)
		}
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Close file before atomic rename
	if err := dstFile.Close(); err != nil {
		if removeErr := os.Remove(tempPath); removeErr != nil {
			fmt.Printf("Warning: failed to remove temp file during cleanup: %v\n", removeErr)
		}
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename with Windows compatibility
	if err := atomicRename(tempPath, dst); err != nil {
		if removeErr := os.Remove(tempPath); removeErr != nil {
			fmt.Printf("Warning: failed to remove temp file during cleanup: %v\n", removeErr)
		}
		return fmt.Errorf("failed to finalize file copy: %w", err)
	}

	return nil
}

// Move safely moves a file from src to dst
func (fm *FileManager) Move(src, dst string) error {
	// First copy the file
	if err := fm.Copy(src, dst); err != nil {
		return err
	}

	// Then remove the source
	if err := fm.Remove(src); err != nil {
		// If remove fails, try to cleanup the destination
		_ = fm.Remove(dst)
		return fmt.Errorf("failed to remove source after copy: %w", err)
	}

	return nil
}

// Remove safely removes a file or directory
func (fm *FileManager) Remove(path string) error {
	if err := ValidatePath(path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	path = filepath.Clean(path)

	// Normalize for cross-platform comparison
	normalizedPath := strings.ToLower(strings.ReplaceAll(path, "\\", "/"))

	// Critical system directories (normalized to forward slashes)
	criticalPaths := []string{
		"/", "/bin", "/sbin", "/usr", "/etc", "/proc", "/sys", "/dev", "/boot", "/root",
		"c:/", "c:/windows", "c:/program files", "c:/program files (x86)",
		"c:/users/all users", "c:/documents and settings",
	}

	for _, criticalPath := range criticalPaths {
		if normalizedPath == criticalPath || strings.HasPrefix(normalizedPath, criticalPath+"/") {
			return fmt.Errorf("refusing to remove system path: %s", path)
		}
	}

	// Allow /tmp and /var/folders (temp directories) but not other /var paths
	if strings.HasPrefix(normalizedPath, "/var/") && !strings.HasPrefix(normalizedPath, "/var/folders/") && !strings.HasPrefix(normalizedPath, "/var/tmp/") {
		return fmt.Errorf("refusing to remove system path: %s", path)
	}

	return os.RemoveAll(path)
}

// Exists checks if a path exists
func (fm *FileManager) Exists(path string) (bool, error) {
	if err := ValidatePath(path); err != nil {
		return false, fmt.Errorf("invalid path: %w", err)
	}

	_, err := os.Stat(filepath.Clean(path))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// IsDir checks if a path is a directory
func (fm *FileManager) IsDir(path string) (bool, error) {
	if err := ValidatePath(path); err != nil {
		return false, fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(filepath.Clean(path))
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// SecureJoin safely joins path components and validates the result
func SecureJoin(base string, elem ...string) (string, error) {
	// Validate base path
	if err := ValidatePath(base); err != nil {
		return "", fmt.Errorf("invalid base path: %w", err)
	}

	// Validate all elements
	for i, el := range elem {
		if err := ValidatePath(el); err != nil {
			return "", fmt.Errorf("invalid path element %d: %w", i, err)
		}
	}

	// Use the secure join library which properly handles symlinks and path traversal
	// Join all elements first
	joinedElem := filepath.Join(elem...)

	// Use securejoin to safely join with the base, preventing any path traversal
	result, err := securejoin.SecureJoin(base, joinedElem)
	if err != nil {
		return "", fmt.Errorf("secure join failed: %w", err)
	}

	return result, nil
}

// ExpandPath safely expands ~ to home directory
func ExpandPath(path string) (string, error) {
	if err := ValidatePath(path); err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, path[2:]), nil
}

// atomicRename performs an atomic rename with proper Windows compatibility
// It uses a retry mechanism to handle file locking issues on Windows
func atomicRename(oldPath, newPath string) error {
	retries := 3
	delay := 10 * time.Millisecond

	for i := 0; i < retries; i++ {
		if runtime.GOOS == "windows" {
			// On Windows, check if the target exists and is accessible
			if _, err := os.Stat(newPath); err == nil {
				// Try to remove existing file to avoid "file in use" errors
				if removeErr := os.Remove(newPath); removeErr != nil {
					// If remove fails, wait and retry
					if i < retries-1 {
						time.Sleep(delay)
						delay *= 2 // Exponential backoff
						continue
					}
					return fmt.Errorf("failed to remove existing file %s: %w", newPath, removeErr)
				}
			}
		}

		// Attempt the rename
		if err := os.Rename(oldPath, newPath); err != nil {
			// If rename fails and we have retries left, wait and try again
			if i < retries-1 {
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
				continue
			}
			return fmt.Errorf("failed to rename %s to %s: %w", oldPath, newPath, err)
		}

		// Success
		return nil
	}

	return fmt.Errorf("failed to rename %s to %s after %d retries", oldPath, newPath, retries)
}
