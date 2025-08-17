package cache

import (
	"sync"
)

// MockManager implements Manager interface for testing
type MockManager struct {
	mu sync.RWMutex

	categories      map[string]interface{}
	agents          map[string]interface{}
	singleAgent     map[string]interface{}
	agentCategories map[string]string

	// Call tracking
	GetCategoriesCalls    int
	SetCategoriesCalls    int
	GetAgentsCalls        []string
	SetAgentsCalls        []string
	GetAgentCalls         []string
	SetAgentCalls         []string
	GetAgentCategoryCalls []string
	SetAgentCategoryCalls []string
	ClearCalls            int
	IsExpiredCalls        []string
	SizeCalls             int
	GetStatsCalls         int

	// Configuration
	shouldExpire bool
	disabled     bool
}

// NewMockManager creates a new mock cache manager
func NewMockManager() *MockManager {
	return &MockManager{
		categories:            make(map[string]interface{}),
		agents:                make(map[string]interface{}),
		singleAgent:           make(map[string]interface{}),
		agentCategories:       make(map[string]string),
		GetAgentsCalls:        make([]string, 0),
		SetAgentsCalls:        make([]string, 0),
		GetAgentCalls:         make([]string, 0),
		SetAgentCalls:         make([]string, 0),
		GetAgentCategoryCalls: make([]string, 0),
		SetAgentCategoryCalls: make([]string, 0),
		IsExpiredCalls:        make([]string, 0),
	}
}

// GetCategories retrieves cached categories
func (m *MockManager) GetCategories() interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetCategoriesCalls++

	if m.disabled || m.shouldExpire {
		return nil
	}

	if categories, exists := m.categories["default"]; exists {
		return categories
	}

	return nil
}

// SetCategories caches categories
func (m *MockManager) SetCategories(categories interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SetCategoriesCalls++

	if !m.disabled {
		m.categories["default"] = categories
	}
}

// GetAgents retrieves cached agents for a category
func (m *MockManager) GetAgents(category string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetAgentsCalls = append(m.GetAgentsCalls, category)

	if m.disabled || m.shouldExpire {
		return nil
	}

	if agents, exists := m.agents[category]; exists {
		return agents
	}

	return nil
}

// SetAgents caches agents for a category
func (m *MockManager) SetAgents(category string, agents interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SetAgentsCalls = append(m.SetAgentsCalls, category)

	if !m.disabled {
		m.agents[category] = agents
	}
}

// GetAgent retrieves a cached agent by ID
func (m *MockManager) GetAgent(agentID string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetAgentCalls = append(m.GetAgentCalls, agentID)

	if m.disabled || m.shouldExpire {
		return nil
	}

	if agent, exists := m.singleAgent[agentID]; exists {
		return agent
	}

	return nil
}

// SetAgent caches an individual agent
func (m *MockManager) SetAgent(agentID string, agent interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SetAgentCalls = append(m.SetAgentCalls, agentID)

	if !m.disabled {
		m.singleAgent[agentID] = agent
	}
}

// GetAgentCategory retrieves the cached category for an agent
func (m *MockManager) GetAgentCategory(agentID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetAgentCategoryCalls = append(m.GetAgentCategoryCalls, agentID)

	if m.disabled || m.shouldExpire {
		return ""
	}

	if category, exists := m.agentCategories[agentID]; exists {
		return category
	}

	return ""
}

// SetAgentCategory caches the category for an agent
func (m *MockManager) SetAgentCategory(agentID string, category string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SetAgentCategoryCalls = append(m.SetAgentCategoryCalls, agentID)

	if !m.disabled {
		m.agentCategories[agentID] = category
	}
}

// Clear removes all cached data
func (m *MockManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ClearCalls++

	m.categories = make(map[string]interface{})
	m.agents = make(map[string]interface{})
	m.singleAgent = make(map[string]interface{})
	m.agentCategories = make(map[string]string)
}

// IsExpired checks if a cache key is expired
func (m *MockManager) IsExpired(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.IsExpiredCalls = append(m.IsExpiredCalls, key)

	return m.shouldExpire
}

// Size returns the number of items in cache
func (m *MockManager) Size() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SizeCalls++

	return len(m.categories) + len(m.agents) + len(m.singleAgent) + len(m.agentCategories)
}

// GetStats returns cache performance statistics
func (m *MockManager) GetStats() CacheStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetStatsCalls++

	// For mock, return simple stats
	size := len(m.categories) + len(m.agents) + len(m.singleAgent) + len(m.agentCategories)

	return CacheStats{
		Hits:      100, // Mock values
		Misses:    10,
		Evictions: 0,
		Size:      size,
		HitRate:   0.9,
	}
}

// SetExpired configures whether items should be considered expired
func (m *MockManager) SetExpired(expired bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shouldExpire = expired
}

// SetDisabled configures whether the cache is disabled
func (m *MockManager) SetDisabled(disabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.disabled = disabled
}

// GetCallCounts returns call counts for testing
func (m *MockManager) GetCallCounts() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]int{
		"GetCategories":    m.GetCategoriesCalls,
		"SetCategories":    m.SetCategoriesCalls,
		"GetAgents":        len(m.GetAgentsCalls),
		"SetAgents":        len(m.SetAgentsCalls),
		"GetAgent":         len(m.GetAgentCalls),
		"SetAgent":         len(m.SetAgentCalls),
		"GetAgentCategory": len(m.GetAgentCategoryCalls),
		"SetAgentCategory": len(m.SetAgentCategoryCalls),
		"Clear":            m.ClearCalls,
		"IsExpired":        len(m.IsExpiredCalls),
		"Size":             m.SizeCalls,
		"GetStats":         m.GetStatsCalls,
	}
}

// Reset clears all call tracking and cached data
func (m *MockManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear data
	m.categories = make(map[string]interface{})
	m.agents = make(map[string]interface{})
	m.singleAgent = make(map[string]interface{})
	m.agentCategories = make(map[string]string)

	// Reset call tracking
	m.GetCategoriesCalls = 0
	m.SetCategoriesCalls = 0
	m.GetAgentsCalls = make([]string, 0)
	m.SetAgentsCalls = make([]string, 0)
	m.GetAgentCalls = make([]string, 0)
	m.SetAgentCalls = make([]string, 0)
	m.GetAgentCategoryCalls = make([]string, 0)
	m.SetAgentCategoryCalls = make([]string, 0)
	m.ClearCalls = 0
	m.IsExpiredCalls = make([]string, 0)
	m.SizeCalls = 0
	m.GetStatsCalls = 0

	// Reset configuration
	m.shouldExpire = false
	m.disabled = false
}
