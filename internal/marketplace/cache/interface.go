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

	// General operations
	Clear()
	IsExpired(key string) bool
	Size() int
}

// Config holds cache configuration
type Config struct {
	Enabled   bool
	TTL       time.Duration
	MaxSizeMB int64
}

// DefaultConfig returns sensible default cache configuration
func DefaultConfig() Config {
	return Config{
		Enabled:   true,
		TTL:       1 * time.Hour,
		MaxSizeMB: 50,
	}
}
