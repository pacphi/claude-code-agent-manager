package fuzzy

import (
	"strings"
	"sync"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
)

// FuzzyMatcher provides fuzzy string matching capabilities for agent discovery
type FuzzyMatcher struct {
	threshold float64
	cache     map[string]float64
	mu        sync.RWMutex
}

// NewFuzzyMatcher creates a new fuzzy matcher with the specified threshold
// threshold should be between 0.0 and 1.0, where higher values require closer matches
func NewFuzzyMatcher(threshold float64) *FuzzyMatcher {
	return &FuzzyMatcher{
		threshold: threshold,
		cache:     make(map[string]float64),
	}
}

// SetThreshold updates the matching threshold
func (fm *FuzzyMatcher) SetThreshold(threshold float64) {
	fm.threshold = threshold
}

// FindBest finds the best matching agent from the provided list
// Returns nil if no match exceeds the threshold
func (fm *FuzzyMatcher) FindBest(query string, agents []*parser.AgentSpec) *parser.AgentSpec {
	var best *parser.AgentSpec
	var bestScore float64

	query = strings.ToLower(strings.TrimSpace(query))

	for _, agent := range agents {
		score := fm.score(query, agent.FileName)
		if score > bestScore && score >= fm.threshold {
			best = agent
			bestScore = score
		}
	}

	return best
}

// FindMultiple finds multiple matching agents, sorted by score (best first)
// If limit is 0, returns all matches above threshold
func (fm *FuzzyMatcher) FindMultiple(query string, agents []*parser.AgentSpec, limit int) []*parser.AgentSpec {
	query = strings.ToLower(strings.TrimSpace(query))

	type scoredAgent struct {
		agent *parser.AgentSpec
		score float64
	}

	var matches []scoredAgent

	// Score all agents
	for _, agent := range agents {
		score := fm.score(query, agent.FileName)
		if score >= fm.threshold {
			matches = append(matches, scoredAgent{agent, score})
		}
	}

	// Sort by score (descending)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[i].score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	// Extract agents
	result := make([]*parser.AgentSpec, len(matches))
	for i, match := range matches {
		result[i] = match.agent
	}

	return result
}

// ScoreByField calculates similarity score between query and specific agent field
func (fm *FuzzyMatcher) ScoreByField(agent *parser.AgentSpec, field, query string) float64 {
	var target string

	switch strings.ToLower(field) {
	case "name":
		target = agent.Name
	case "description":
		target = agent.Description
	case "filename":
		target = agent.FileName
	case "prompt", "content":
		target = agent.Prompt
	case "tools":
		target = strings.Join(agent.Tools, " ")
	case "source":
		target = agent.Source
	default:
		return 0.0
	}

	return fm.score(query, target)
}

// MultiFieldSearch searches across multiple fields with relevance scoring
func (fm *FuzzyMatcher) MultiFieldSearch(query string, agents []*parser.AgentSpec, fields []string, limit int) []*parser.AgentSpec {
	type scoredAgent struct {
		agent *parser.AgentSpec
		score float64
	}

	if len(fields) == 0 {
		// Default to searching all fields
		fields = []string{"name", "description", "filename", "prompt"}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var matches []scoredAgent

	// Score all agents across all fields
	for _, agent := range agents {
		maxScore := 0.0
		totalScore := 0.0
		validFields := 0

		for _, field := range fields {
			fieldScore := fm.ScoreByField(agent, field, query)
			if fieldScore > 0 {
				totalScore += fieldScore
				validFields++
				if fieldScore > maxScore {
					maxScore = fieldScore
				}
			}
		}

		// Use weighted combination of max score and average score
		// This gives preference to agents with high scores in any field
		// while also considering agents with decent scores across multiple fields
		if validFields > 0 {
			avgScore := totalScore / float64(validFields)
			finalScore := (maxScore * 0.7) + (avgScore * 0.3)

			if finalScore >= fm.threshold {
				matches = append(matches, scoredAgent{agent, finalScore})
			}
		}
	}

	// Sort by score (descending)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[i].score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	// Extract agents
	result := make([]*parser.AgentSpec, len(matches))
	for i, match := range matches {
		result[i] = match.agent
	}

	return result
}

// Score calculates similarity between two strings using multiple algorithms
func (fm *FuzzyMatcher) Score(s1, s2 string) float64 {
	return fm.score(s1, s2)
}

// score calculates similarity between two strings using multiple algorithms
func (fm *FuzzyMatcher) score(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == "" || s2 == "" {
		return 0
	}

	// Exact match gets highest score
	if s1 == s2 {
		return 1.0
	}

	// Substring matching - if s1 is contained in s2, give a good score
	if strings.Contains(s2, s1) {
		baseScore := float64(len(s1)) / float64(len(s2))
		// Boost substring matches significantly
		return baseScore * 2.0 // This will make "data" in "data-processor.md" score ~0.47
	}

	// Token-based matching for multi-word queries
	tokens1 := strings.Fields(strings.ReplaceAll(s1, "-", " "))
	tokens2 := strings.Fields(strings.ReplaceAll(s2, "-", " "))

	if len(tokens1) == 0 {
		return 0
	}

	matches := 0
	totalScore := 0.0
	for _, token1 := range tokens1 {
		bestTokenScore := 0.0
		for _, token2 := range tokens2 {
			if strings.Contains(token2, token1) {
				tokenScore := float64(len(token1)) / float64(len(token2))
				if tokenScore > bestTokenScore {
					bestTokenScore = tokenScore
				}
			} else if strings.Contains(token1, token2) {
				tokenScore := float64(len(token2)) / float64(len(token1))
				if tokenScore > bestTokenScore {
					bestTokenScore = tokenScore
				}
			}
		}
		if bestTokenScore > 0 {
			matches++
			totalScore += bestTokenScore
		}
	}

	if matches > 0 {
		return totalScore / float64(len(tokens1))
	}

	// Character-based similarity for very fuzzy matching
	charSim := fm.characterSimilarity(s1, s2)
	// Don't let character similarity go too low for reasonable matches
	if charSim > 0.1 {
		return charSim
	}

	return 0
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (fm *FuzzyMatcher) levenshteinDistance(s1, s2 string) int {
	cacheKey := s1 + "|" + s2

	// Check cache first
	fm.mu.RLock()
	if dist, exists := fm.cache[cacheKey]; exists {
		fm.mu.RUnlock()
		return int(dist)
	}
	fm.mu.RUnlock()

	len1, len2 := len(s1), len(s2)

	// Handle edge cases
	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	// Create matrix for dynamic programming
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	distance := matrix[len1][len2]

	// Cache the result
	fm.mu.Lock()
	if len(fm.cache) > 1000 { // Prevent unbounded cache growth
		fm.cache = make(map[string]float64)
	}
	fm.cache[cacheKey] = float64(distance)
	fm.mu.Unlock()

	return distance
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

// characterSimilarity calculates character-level similarity using Levenshtein distance
func (fm *FuzzyMatcher) characterSimilarity(s1, s2 string) float64 {
	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	distance := fm.levenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	// Convert distance to similarity score (0-1)
	similarity := 1.0 - float64(distance)/float64(maxLen)
	if similarity < 0 {
		similarity = 0
	}

	return similarity
}
