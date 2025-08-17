package transformer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/config"
)

func TestNew(t *testing.T) {
	settings := config.Settings{
		BaseDir: "/tmp/test",
	}

	transformer := New(settings)

	if transformer == nil {
		t.Fatal("Expected transformer but got nil")
	}

	if transformer.settings.BaseDir != settings.BaseDir {
		t.Errorf("Expected BaseDir %s, got %s", settings.BaseDir, transformer.settings.BaseDir)
	}
}

func TestApply(t *testing.T) {
	transformer := New(config.Settings{})

	tests := []struct {
		name        string
		transform   config.Transformation
		files       []string
		wantErr     bool
		expectFiles int
	}{
		{
			name: "remove_numeric_prefix",
			transform: config.Transformation{
				Type:    "remove_numeric_prefix",
				Pattern: "^[0-9]{2}-",
			},
			files:       []string{"01-test/file.md", "02-example/other.md"},
			wantErr:     false,
			expectFiles: 2,
		},
		{
			name: "unknown transformation",
			transform: config.Transformation{
				Type: "unknown_transform",
			},
			files:   []string{"test.md"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.Apply(tt.files, tt.transform, "/src", "/dst")
			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(result) != tt.expectFiles {
				t.Errorf("Apply() returned %d files, expected %d", len(result), tt.expectFiles)
			}
		})
	}
}

func TestRemoveNumericPrefix(t *testing.T) {
	transformer := New(config.Settings{})

	tests := []struct {
		name     string
		pattern  string
		files    []string
		expected []string
		wantErr  bool
	}{
		{
			name:     "default pattern",
			pattern:  "",
			files:    []string{"01-intro/file.md", "02-advanced/guide.md"},
			expected: []string{"intro/file.md", "advanced/guide.md"},
			wantErr:  false,
		},
		{
			name:     "custom pattern",
			pattern:  "^[0-9]{3}_",
			files:    []string{"001_basics/start.md", "002_expert/end.md"},
			expected: []string{"basics/start.md", "expert/end.md"},
			wantErr:  false,
		},
		{
			name:     "no match",
			pattern:  "^[0-9]{2}-",
			files:    []string{"intro/file.md", "advanced/guide.md"},
			expected: []string{"intro/file.md", "advanced/guide.md"},
			wantErr:  false,
		},
		{
			name:    "invalid pattern",
			pattern: "[",
			files:   []string{"test.md"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transform := config.Transformation{
				Type:    "remove_numeric_prefix",
				Pattern: tt.pattern,
			}

			result, err := transformer.removeNumericPrefix(tt.files, transform, "/src", "/dst")
			if (err != nil) != tt.wantErr {
				t.Errorf("removeNumericPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) != len(tt.expected) {
					t.Errorf("removeNumericPrefix() returned %d files, expected %d", len(result), len(tt.expected))
					return
				}

				for i, expected := range tt.expected {
					if result[i] != expected {
						t.Errorf("removeNumericPrefix() result[%d] = %s, expected %s", i, result[i], expected)
					}
				}
			}
		})
	}
}

func TestExtractDocs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "transformer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	transformer := New(config.Settings{})

	// Create test files
	sourcePath := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create a file with documentation content
	testFile := filepath.Join(sourcePath, "agent.md")
	content := `# Agent Documentation

This is a test agent with documentation.

## Usage

Example usage here.`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targetPath := filepath.Join(tempDir, "docs")

	transform := config.Transformation{
		Type:      "extract_docs",
		TargetDir: "docs/",
	}

	files := []string{testFile}
	result, err := transformer.extractDocs(files, transform, sourcePath, targetPath)
	if err != nil {
		t.Fatalf("extractDocs() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("extractDocs() should return at least one file")
	}
}

func TestRenameFiles(t *testing.T) {
	transformer := New(config.Settings{})

	tests := []struct {
		name    string
		pattern string
		files   []string
		wantErr bool
	}{
		{
			name:    "valid pattern",
			pattern: "\\.txt$",
			files:   []string{"readme.txt", "guide.txt"},
			wantErr: false,
		},
		{
			name:    "empty pattern",
			pattern: "",
			files:   []string{"test.txt"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transform := config.Transformation{
				Type:    "rename_files",
				Pattern: tt.pattern,
			}

			result, err := transformer.renameFiles(tt.files, transform, "/src", "/dst")
			if (err != nil) != tt.wantErr {
				t.Errorf("renameFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(result) == 0 {
				t.Error("renameFiles() should return files")
			}
		})
	}
}

func TestReplaceContent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "replace-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	transformer := New(config.Settings{})

	testFile := filepath.Join(tempDir, "test.md")
	originalContent := "Hello World! This is a test."
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	transform := config.Transformation{
		Type:    "replace_content",
		Pattern: "World",
	}

	files := []string{testFile}
	result, err := transformer.replaceContent(files, transform, tempDir, tempDir)
	if err != nil {
		t.Fatalf("replaceContent() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("replaceContent() returned %d files, expected 1", len(result))
	}
}
