package fuzzy

import (
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/stretchr/testify/assert"
)

func TestNewFuzzyMatcher(t *testing.T) {
	tests := []struct {
		name      string
		threshold float64
	}{
		{
			name:      "default threshold",
			threshold: 0.7,
		},
		{
			name:      "high threshold",
			threshold: 0.9,
		},
		{
			name:      "low threshold",
			threshold: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := NewFuzzyMatcher(tt.threshold)
			assert.NotNil(t, fm)
			assert.Equal(t, tt.threshold, fm.threshold)
		})
	}
}

func TestFuzzyMatcher_Score(t *testing.T) {
	fm := NewFuzzyMatcher(0.7)

	tests := []struct {
		name     string
		s1       string
		s2       string
		minScore float64 // Use minimum expected score instead of exact
	}{
		{
			name:     "exact match",
			s1:       "test",
			s2:       "test",
			minScore: 1.0,
		},
		{
			name:     "substring match",
			s1:       "test",
			s2:       "test-agent.md",
			minScore: 0.3, // Should be at least 30% match
		},
		{
			name:     "partial match",
			s1:       "agent",
			s2:       "test-agent-helper.md",
			minScore: 0.2, // Should be at least 20% match
		},
		{
			name:     "token match",
			s1:       "data processor",
			s2:       "data-processor-agent.md",
			minScore: 0.8, // Should be high score for token match
		},
		{
			name:     "partial token match",
			s1:       "data science helper",
			s2:       "data-processor.md",
			minScore: 0.2, // At least 20% for partial token match
		},
		{
			name:     "no match",
			s1:       "xyz",
			s2:       "test-agent.md",
			minScore: 0.0,
		},
		{
			name:     "case insensitive",
			s1:       "TEST",
			s2:       "test-agent.md",
			minScore: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := fm.score(tt.s1, tt.s2)
			switch tt.minScore {
			case 0.0:
				assert.Equal(t, 0.0, score, "Expected no match")
			case 1.0:
				assert.Equal(t, 1.0, score, "Expected exact match")
			default:
				assert.GreaterOrEqual(t, score, tt.minScore, "Expected score at least %f, got %f", tt.minScore, score)
			}
		})
	}
}

func TestFuzzyMatcher_FindBest(t *testing.T) {
	fm := NewFuzzyMatcher(0.3)

	agents := []*parser.AgentSpec{
		{
			Name:     "test-agent",
			FileName: "test-agent.md",
		},
		{
			Name:     "data-processor",
			FileName: "data-processor.md",
		},
		{
			Name:     "web-scraper",
			FileName: "web-scraper-helper.md",
		},
		{
			Name:     "code-reviewer",
			FileName: "code-reviewer.md",
		},
	}

	tests := []struct {
		name     string
		query    string
		expected *parser.AgentSpec
	}{
		{
			name:     "exact filename match",
			query:    "test-agent.md",
			expected: agents[0],
		},
		{
			name:     "partial filename match",
			query:    "test",
			expected: agents[0],
		},
		{
			name:     "fuzzy match",
			query:    "data",
			expected: agents[1],
		},
		{
			name:     "multi-word match",
			query:    "web scraper",
			expected: agents[2],
		},
		{
			name:     "no match above threshold",
			query:    "nonexistent",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.FindBest(tt.query, agents)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuzzyMatcher_FindBestWithHighThreshold(t *testing.T) {
	fm := NewFuzzyMatcher(0.8) // High threshold

	agents := []*parser.AgentSpec{
		{
			Name:     "test-agent",
			FileName: "test-agent.md",
		},
		{
			Name:     "different-name",
			FileName: "different-name.md",
		},
	}

	tests := []struct {
		name     string
		query    string
		expected *parser.AgentSpec
	}{
		{
			name:     "high score match",
			query:    "test-agent",
			expected: agents[0],
		},
		{
			name:     "low score filtered out",
			query:    "test", // Lower score due to partial match
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.FindBest(tt.query, agents)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuzzyMatcher_FindMultiple(t *testing.T) {
	fm := NewFuzzyMatcher(0.3)

	agents := []*parser.AgentSpec{
		{
			Name:     "test-agent",
			FileName: "test-agent.md",
		},
		{
			Name:     "test-helper",
			FileName: "test-helper.md",
		},
		{
			Name:     "data-processor",
			FileName: "data-processor.md",
		},
		{
			Name:     "testing-utils",
			FileName: "testing-utils.md",
		},
	}

	tests := []struct {
		name     string
		query    string
		limit    int
		expected int // Number of expected results
	}{
		{
			name:     "multiple matches",
			query:    "test",
			limit:    5,
			expected: 3, // test-agent, test-helper, testing-utils
		},
		{
			name:     "limited results",
			query:    "test",
			limit:    2,
			expected: 2,
		},
		{
			name:     "no limit",
			query:    "test",
			limit:    0, // No limit
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := fm.FindMultiple(tt.query, agents, tt.limit)
			assert.Len(t, results, tt.expected)
		})
	}
}

func TestFuzzyMatcher_ScoreByField(t *testing.T) {
	fm := NewFuzzyMatcher(0.5)

	agent := &parser.AgentSpec{
		Name:        "data-processor",
		Description: "Process data files efficiently",
		FileName:    "data-processor.md",
	}

	tests := []struct {
		name     string
		field    string
		query    string
		minScore float64
	}{
		{
			name:     "name field exact",
			field:    "name",
			query:    "data-processor",
			minScore: 1.0,
		},
		{
			name:     "name field partial",
			field:    "name",
			query:    "data",
			minScore: 0.2, // At least 20% match
		},
		{
			name:     "description field",
			field:    "description",
			query:    "process",
			minScore: 0.2, // At least 20% match
		},
		{
			name:     "filename field",
			field:    "filename",
			query:    "processor",
			minScore: 0.5, // Should be a good match
		},
		{
			name:     "invalid field",
			field:    "invalid",
			query:    "test",
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := fm.ScoreByField(agent, tt.field, tt.query)
			switch tt.minScore {
			case 0.0:
				assert.Equal(t, 0.0, score, "Expected no match")
			case 1.0:
				assert.Equal(t, 1.0, score, "Expected exact match")
			default:
				assert.GreaterOrEqual(t, score, tt.minScore, "Expected score at least %f, got %f", tt.minScore, score)
			}
		})
	}
}

func TestFuzzyMatcher_SetThreshold(t *testing.T) {
	fm := NewFuzzyMatcher(0.5)

	// Test initial threshold
	assert.Equal(t, 0.5, fm.threshold)

	// Test threshold update
	fm.SetThreshold(0.8)
	assert.Equal(t, 0.8, fm.threshold)

	// Test that new threshold affects matching
	agents := []*parser.AgentSpec{
		{
			Name:     "test-agent",
			FileName: "test-agent.md",
		},
	}

	// Should not match with high threshold
	result := fm.FindBest("test", agents)
	assert.Nil(t, result)

	// Lower threshold should match
	fm.SetThreshold(0.3)
	result = fm.FindBest("test", agents)
	assert.Equal(t, agents[0], result)
}
