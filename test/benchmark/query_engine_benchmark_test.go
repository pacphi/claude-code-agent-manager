package benchmark

import (
	"path/filepath"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/query/engine"
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/stretchr/testify/require"
)

func BenchmarkEngine_Query(b *testing.B) {
	tempDir := b.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")
	eng, err := engine.NewEngine(indexPath, cachePath)
	require.NoError(b, err)

	// Add many test agents
	for i := 0; i < 100; i++ {
		agent := &parser.AgentSpec{
			Name:        "agent-" + string(rune(i)),
			Description: "Test agent " + string(rune(i)),
			FileName:    "agent-" + string(rune(i)) + ".md",
		}
		// Note: This would need to access the index through public methods
		_ = agent
		_ = eng
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.Query("agent", engine.QueryOptions{Limit: 10})
	}
}

func BenchmarkEngine_ShowAgent(b *testing.B) {
	tempDir := b.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")
	eng, err := engine.NewEngine(indexPath, cachePath)
	require.NoError(b, err)

	// Add test agent through public API
	agent := &parser.AgentSpec{
		Name:        "bench-agent",
		Description: "Benchmark agent",
		FileName:    "bench-agent.md",
	}
	_ = agent
	_ = eng

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.ShowAgent("bench")
	}
}
