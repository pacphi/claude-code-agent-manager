package benchmark

import (
	"fmt"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/query/fuzzy"
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
)

func BenchmarkFuzzyMatcher_Score(b *testing.B) {
	fm := fuzzy.NewFuzzyMatcher(0.7)
	s1 := "test-agent-helper"
	s2 := "test-agent-helper.md"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This would need to be exposed as a public method or tested via public API
		_ = fm
		_ = s1
		_ = s2
	}
}

func BenchmarkFuzzyMatcher_FindBest(b *testing.B) {
	fm := fuzzy.NewFuzzyMatcher(0.7)

	// Create a larger dataset
	agents := make([]*parser.AgentSpec, 100)
	for i := 0; i < 100; i++ {
		agents[i] = &parser.AgentSpec{
			Name:     fmt.Sprintf("agent-%d", i),
			FileName: fmt.Sprintf("agent-%d.md", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fm.FindBest("agent-50", agents)
	}
}
