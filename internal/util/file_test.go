package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileManager_Copy(t *testing.T) {
	fm := NewFileManager()

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "file-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test source file
	srcFile := filepath.Join(tempDir, "source.txt")
	content := "test content"
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	tests := []struct {
		name    string
		src     string
		dst     string
		wantErr bool
	}{
		{
			name:    "valid copy",
			src:     srcFile,
			dst:     filepath.Join(tempDir, "dest.txt"),
			wantErr: false,
		},
		{
			name:    "invalid source path",
			src:     "../../../etc/passwd",
			dst:     filepath.Join(tempDir, "dest2.txt"),
			wantErr: true,
		},
		{
			name:    "invalid destination path",
			src:     srcFile,
			dst:     "../../../tmp/malicious.txt",
			wantErr: true,
		},
		{
			name:    "nonexistent source",
			src:     filepath.Join(tempDir, "nonexistent.txt"),
			dst:     filepath.Join(tempDir, "dest3.txt"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fm.Copy(tt.src, tt.dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileManager.Copy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If copy should succeed, verify the file was copied
			if !tt.wantErr {
				if _, err := os.Stat(tt.dst); os.IsNotExist(err) {
					t.Errorf("FileManager.Copy() did not create destination file")
				}

				// Verify content
				dstContent, err := os.ReadFile(tt.dst)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
				}
				if string(dstContent) != content {
					t.Errorf("FileManager.Copy() content mismatch, got %s, want %s", string(dstContent), content)
				}
			}
		})
	}
}

func TestFileManager_Remove(t *testing.T) {
	fm := NewFileManager()

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "remove-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file to remove
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid remove",
			path:    testFile,
			wantErr: false,
		},
		{
			name:    "system path protection",
			path:    "/etc",
			wantErr: true,
		},
		{
			name:    "path traversal protection",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "root path protection",
			path:    "/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fm.Remove(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileManager.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecureJoin(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "join-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		base    string
		elems   []string
		wantErr bool
	}{
		{
			name:    "valid join",
			base:    tempDir,
			elems:   []string{"subdir", "file.txt"},
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			base:    tempDir,
			elems:   []string{"..", "..", "etc", "passwd"},
			wantErr: true,
		},
		{
			name:    "invalid base path",
			base:    "../../../etc",
			elems:   []string{"passwd"},
			wantErr: true,
		},
		{
			name:    "null byte in element",
			base:    tempDir,
			elems:   []string{"test\x00file"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SecureJoin(tt.base, tt.elems...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureJoin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If join should succeed, verify result is under base
			if !tt.wantErr && result != "" {
				// The result should start with the base path
				absBase, _ := filepath.Abs(tt.base)
				absResult, _ := filepath.Abs(result)
				if !strings.HasPrefix(absResult, absBase) {
					t.Errorf("SecureJoin() result %s is not under base %s", absResult, absBase)
				}
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "regular path",
			path:    "/tmp/test",
			wantErr: false,
		},
		{
			name:    "home path expansion",
			path:    "~/test",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			path:    "~/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "null byte injection",
			path:    "~/test\x00file",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If expansion should succeed, verify result
			if !tt.wantErr {
				if tt.path == "~/test" {
					// Should be expanded to absolute path
					if !filepath.IsAbs(result) {
						t.Errorf("ExpandPath() should return absolute path for ~/test, got %s", result)
					}
				} else if tt.path == "/tmp/test" {
					// Should remain unchanged
					if result != tt.path {
						t.Errorf("ExpandPath() should not change absolute path, got %s, want %s", result, tt.path)
					}
				}
			}
		})
	}
}
