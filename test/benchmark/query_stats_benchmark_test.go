package benchmark

import (
	"fmt"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/pacphi/claude-code-agent-manager/internal/query/stats"
)

func BenchmarkCalculator_Calculate(b *testing.B) {
	// Create a large set of test agents
	agents := make([]*parser.AgentSpec, 1000)
	for i := 0; i < 1000; i++ {
		agents[i] = &parser.AgentSpec{
			Name:        fmt.Sprintf("agent-%d", i),
			Description: fmt.Sprintf("Test agent number %d", i),
			Tools:       []string{"Read", "Write"},
			Prompt:      fmt.Sprintf("You are agent %d", i),
			Source:      "benchmark",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc := stats.NewCalculator(agents)
		_ = calc.Calculate()
	}
}

func BenchmarkCalculator_Coverage(b *testing.B) {
	// Create test agents
	agents := make([]*parser.AgentSpec, 100)
	for i := 0; i < 100; i++ {
		agents[i] = &parser.AgentSpec{
			Name:        fmt.Sprintf("agent-%d", i),
			Description: fmt.Sprintf("Test agent %d", i),
			Tools:       []string{"Read"},
			Prompt:      "Test prompt",
		}
	}

	calc := stats.NewCalculator(agents)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This would need to be exposed as a public method or tested via public API
		_ = calc // Placeholder - would need access to private method or test via public interface
	}
}
