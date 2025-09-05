package benchmark

import (
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/pacphi/claude-code-agent-manager/internal/query/validator"
)

// BenchmarkValidate tests validation performance
func BenchmarkValidate(b *testing.B) {
	validator := validator.NewValidator()

	spec := &parser.AgentSpec{
		Name:        "benchmark-agent",
		Description: "Agent for benchmarking validation performance",
		Tools:       []string{"Read", "Write", "Edit", "Bash", "Grep"},
		Prompt:      "This is a benchmark agent prompt for testing validation speed.",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(spec)
	}
}

// BenchmarkValidateWithReport tests detailed validation performance
func BenchmarkValidateWithReport(b *testing.B) {
	validator := validator.NewValidator()

	spec := &parser.AgentSpec{
		Name:        "benchmark-agent",
		Description: "Agent for benchmarking detailed validation performance",
		Tools:       []string{"Read", "Write", "Edit", "Bash", "Grep"},
		Prompt:      "This is a benchmark agent prompt for testing detailed validation speed.",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateWithReport(spec)
	}
}
