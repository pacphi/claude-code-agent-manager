package benchmark

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/query/index"
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
)

func BenchmarkAddAgent(b *testing.B) {
	tmpDir := b.TempDir()
	indexPath := filepath.Join(tmpDir, "bench-index.json")
	im, err := index.NewIndexManager(indexPath)
	if err != nil {
		b.Fatalf("NewIndexManager failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent := createBenchmarkAgent(
			fmt.Sprintf("bench-agent%d", i),
			fmt.Sprintf("Benchmark agent %d", i),
			[]string{"Read", "Write"},
			fmt.Sprintf("Benchmark prompt %d with some content", i),
		)
		im.AddAgent(agent)
	}
}

func BenchmarkSearch(b *testing.B) {
	tmpDir := b.TempDir()
	indexPath := filepath.Join(tmpDir, "bench-index.json")
	im, err := index.NewIndexManager(indexPath)
	if err != nil {
		b.Fatalf("NewIndexManager failed: %v", err)
	}

	// Add many agents for realistic benchmarking
	for i := 0; i < 1000; i++ {
		tools := []string{"Read", "Write", "Edit"}
		agent := createBenchmarkAgent(
			fmt.Sprintf("agent%d", i),
			fmt.Sprintf("Agent %d for benchmarking search performance", i),
			tools[i%3:i%3+1],
			fmt.Sprintf("Agent %d prompt with benchmark content for testing", i),
		)
		im.AddAgent(agent)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = im.Search("agent", index.QueryOptions{})
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	tmpDir := b.TempDir()
	indexPath := filepath.Join(tmpDir, "bench-index.json")
	im, err := index.NewIndexManager(indexPath)
	if err != nil {
		b.Fatalf("NewIndexManager failed: %v", err)
	}

	// Add test agents
	for i := 0; i < 100; i++ {
		agent := createBenchmarkAgent(
			fmt.Sprintf("agent%d", i),
			fmt.Sprintf("Agent %d description", i),
			[]string{"Read"},
			fmt.Sprintf("Agent %d prompt", i),
		)
		im.AddAgent(agent)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = im.Search("agent", index.QueryOptions{})
		}
	})
}

// createBenchmarkAgent creates a test agent for benchmarking
func createBenchmarkAgent(name, description string, tools []string, prompt string) *parser.AgentSpec {
	return &parser.AgentSpec{
		Name:        name,
		Description: description,
		Tools:       tools,
		Prompt:      prompt,
		FileName:    name + ".md",
	}
}
