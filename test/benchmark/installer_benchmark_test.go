package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/conflict"
	"github.com/pacphi/claude-code-agent-manager/internal/installer"
	"github.com/pacphi/claude-code-agent-manager/internal/tracker"
)

func BenchmarkExtractAgentMetadata(b *testing.B) {
	// Create temporary test environment
	tempDir := b.TempDir()

	// Create test files with various extensions - smaller set for benchmark
	testFiles := createBenchmarkTestFiles(b, tempDir, 100) // 100 files total

	// Create installer
	cfg := &config.Config{
		Settings: config.Settings{
			BaseDir: tempDir,
		},
	}

	track := tracker.New(filepath.Join(tempDir, "test.json"))
	resolver := conflict.NewResolver("backup", tempDir)
	install := installer.New(cfg, track, resolver, installer.Options{})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Use reflection or testing.TB interface to access the private method
		// This is a simplified approach for the benchmark
		_ = testFiles // Placeholder for now - would need to expose the method or test the public API
		_ = install   // Prevent unused variable error
	}
}

func BenchmarkPreFilterMarkdownFiles(b *testing.B) {
	// Create installer
	cfg := &config.Config{}
	tracker := tracker.New("")
	resolver := conflict.NewResolver("backup", "")
	install := installer.New(cfg, tracker, resolver, installer.Options{})

	// Create test file lists with different ratios of markdown files
	testCases := []struct {
		name       string
		totalFiles int
		mdRatio    float64 // Percentage of files that are .md
	}{
		{"SmallSet_HighMD", 100, 0.8},    // 80 .md files out of 100
		{"MediumSet_MediumMD", 500, 0.5}, // 250 .md files out of 500
		{"LargeSet_LowMD", 1000, 0.2},    // 200 .md files out of 1000
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			files := createMockFileList(tc.totalFiles, tc.mdRatio)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Use a public method or interface instead of private method
				_ = install // Placeholder
				_ = files   // Placeholder
			}
		})
	}
}

// MemoryBenchmark tests memory usage of the optimization
func BenchmarkInstallerMemoryUsage(b *testing.B) {
	// Create test files - simulate large repository
	files := createMockFileList(10000, 0.1) // 10k files, 10% markdown

	cfg := &config.Config{}
	tracker := tracker.New("")
	resolver := conflict.NewResolver("backup", "")
	install := installer.New(cfg, tracker, resolver, installer.Options{})

	b.Run("PreFilter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Simulate filtering markdown files
			mdFiles := make([]string, 0)
			for _, file := range files {
				if filepath.Ext(file) == ".md" {
					mdFiles = append(mdFiles, file)
				}
			}
			_ = install // Prevent optimization
			_ = mdFiles
		}
	})

	b.Run("WithoutPreFilter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Simulate old approach - allocate for all files then filter
			allFilesList := make([]string, 0, len(files))
			for _, file := range files {
				if filepath.Ext(file) == ".md" {
					allFilesList = append(allFilesList, file)
				}
			}
			_ = allFilesList
		}
	})
}

// createBenchmarkTestFiles creates actual files for realistic benchmarking
func createBenchmarkTestFiles(b *testing.B, tempDir string, totalFiles int) []string {
	files := make([]string, 0, totalFiles)

	// Create 20% markdown files, 80% other files
	mdCount := totalFiles / 5
	otherCount := totalFiles - mdCount

	// Create markdown files with minimal agent content
	for i := 0; i < mdCount; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("agent-%d.md", i))
		content := `---
name: Test Agent
description: Test agent for benchmarking
tools: ["Read", "Write"]
---

# Test Agent

This is a test agent for benchmarking.`

		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		files = append(files, fmt.Sprintf("agent-%d.md", i))
	}

	// Create other files
	extensions := []string{".txt", ".go", ".js", ".py", ".json", ".yml", ".xml"}
	for i := 0; i < otherCount; i++ {
		ext := extensions[i%len(extensions)]
		filename := filepath.Join(tempDir, fmt.Sprintf("file-%d%s", i, ext))
		err := os.WriteFile(filename, []byte("test content"), 0644)
		if err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		files = append(files, fmt.Sprintf("file-%d%s", i, ext))
	}

	return files
}

// createMockFileList creates a list of file names for benchmarking (no actual files)
func createMockFileList(totalFiles int, mdRatio float64) []string {
	files := make([]string, 0, totalFiles)
	mdCount := int(float64(totalFiles) * mdRatio)

	// Add markdown files
	for i := 0; i < mdCount; i++ {
		files = append(files, fmt.Sprintf("agent-%d.md", i))
	}

	// Add other files
	extensions := []string{".txt", ".go", ".js", ".py", ".json", ".yml", ".xml"}
	remaining := totalFiles - mdCount
	for i := 0; i < remaining; i++ {
		ext := extensions[i%len(extensions)]
		files = append(files, fmt.Sprintf("file-%d%s", i, ext))
	}

	return files
}
