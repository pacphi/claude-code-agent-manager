package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Set same permissions as source
	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
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

	// Additional safety check - don't allow removing critical system directories
	criticalPaths := []string{
		"/", "/bin", "/sbin", "/usr", "/etc", "/proc", "/sys", "/dev",
	}

	for _, criticalPath := range criticalPaths {
		if path == criticalPath || strings.HasPrefix(path, criticalPath+"/") {
			return fmt.Errorf("refusing to remove system path: %s", path)
		}
	}

	// Allow /tmp and /var/folders (temp directories) but not other /var paths
	if strings.HasPrefix(path, "/var/") && !strings.HasPrefix(path, "/var/folders/") && !strings.HasPrefix(path, "/var/tmp/") {
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

	// Join and clean the path
	result := filepath.Join(append([]string{base}, elem...)...)
	result = filepath.Clean(result)

	// Ensure the result is still under the base path
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute base path: %w", err)
	}

	absResult, err := filepath.Abs(result)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute result path: %w", err)
	}

	if !strings.HasPrefix(absResult, absBase) {
		return "", fmt.Errorf("path traversal detected: result %s is outside base %s", absResult, absBase)
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
