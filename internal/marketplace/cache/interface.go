package cache

import (
	"time"
)

// Manager defines the interface for cache operations
// Uses interface{} to avoid import cycles - type assertions handled by callers
type Manager interface {
	// Categories
	GetCategories() interface{}
	SetCategories(categories interface{})

	// Agents
	GetAgents(category string) interface{}
	SetAgents(category string, agents interface{})
	GetAgent(agentID string) interface{}
	SetAgent(agentID string, agent interface{})

	// Agent-to-category mapping for fast lookups
	GetAgentCategory(agentID string) string
	SetAgentCategory(agentID string, category string)

	// General operations
	Clear()
	IsExpired(key string) bool
	Size() int
	GetStats() CacheStats
}

// CacheStats holds cache performance metrics
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int
	HitRate   float64
}

// Config holds cache configuration
type Config struct {
	Enabled   bool
	TTL       time.Duration
	MaxSizeMB int64
}
