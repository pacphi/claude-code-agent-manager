package benchmark

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/util"
)

// BenchmarkFileCopyOptimization benchmarks the file copy optimization
func BenchmarkFileCopyOptimization(b *testing.B) {
	// Test different file sizes
	fileSizes := []struct {
		name string
		size int64
	}{
		{"Small_1KB", 1 * 1024},
		{"Medium_100KB", 100 * 1024},
		{"Large_10MB", 10 * 1024 * 1024},
		{"XLarge_100MB", 100 * 1024 * 1024},
	}

	for _, fs := range fileSizes {
		b.Run(fs.name, func(b *testing.B) {
			// Create test file content
			content := make([]byte, fs.size)
			_, err := rand.Read(content)
			if err != nil {
				b.Fatalf("Failed to generate test content: %v", err)
			}

			tempDir := b.TempDir()
			srcPath := filepath.Join(tempDir, "source.dat")

			// Write test file
			if err := os.WriteFile(srcPath, content, 0644); err != nil {
				b.Fatalf("Failed to write test file: %v", err)
			}

			fm := util.NewFileManager()
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				dstPath := filepath.Join(tempDir, fmt.Sprintf("dest-%d.dat", i))
				if err := fm.Copy(srcPath, dstPath); err != nil {
					b.Fatalf("Copy failed: %v", err)
				}
				// Clean up destination file to avoid filling disk
				os.Remove(dstPath)
			}
		})
	}
}

// BenchmarkBufferSizes tests different buffer sizes to validate our 64KB choice
func BenchmarkBufferSizes(b *testing.B) {
	// Test different buffer sizes
	bufferSizes := []struct {
		name string
		size int
	}{
		{"Default", 0}, // io.Copy default (32KB)
		{"Small_4KB", 4 * 1024},
		{"Medium_32KB", 32 * 1024},
		{"Optimal_64KB", 64 * 1024},
		{"Large_128KB", 128 * 1024},
		{"XLarge_1MB", 1024 * 1024},
	}

	// Create 10MB test data
	testData := make([]byte, 10*1024*1024)
	_, err := rand.Read(testData)
	if err != nil {
		b.Fatalf("Failed to generate test data: %v", err)
	}

	for _, bs := range bufferSizes {
		b.Run(bs.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				src := bytes.NewReader(testData)
				dst := &bytes.Buffer{}

				if bs.size == 0 {
					// Use standard io.Copy
					_, err := io.Copy(dst, src)
					if err != nil {
						b.Fatalf("Copy failed: %v", err)
					}
				} else {
					// Use io.CopyBuffer with specific buffer size
					buffer := make([]byte, bs.size)
					_, err := io.CopyBuffer(dst, src, buffer)
					if err != nil {
						b.Fatalf("CopyBuffer failed: %v", err)
					}
				}
			}
		})
	}
}

// BenchmarkRealFileOperations tests with actual file I/O
func BenchmarkRealFileOperations(b *testing.B) {
	tempDir := b.TempDir()

	// Create a 5MB test file
	testSize := 5 * 1024 * 1024
	testContent := make([]byte, testSize)
	_, err := rand.Read(testContent)
	if err != nil {
		b.Fatalf("Failed to generate test content: %v", err)
	}

	srcPath := filepath.Join(tempDir, "test_source.bin")
	if err := os.WriteFile(srcPath, testContent, 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	fm := util.NewFileManager()

	b.Run("OptimizedFileCopy", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			dstPath := filepath.Join(tempDir, fmt.Sprintf("optimized_dest_%d.bin", i))
			if err := fm.Copy(srcPath, dstPath); err != nil {
				b.Fatalf("Optimized copy failed: %v", err)
			}
			os.Remove(dstPath) // Cleanup
		}
	})
}
