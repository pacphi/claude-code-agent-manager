package cache

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// ristrettoManager implements Manager interface using ristretto cache
type ristrettoManager struct {
	cache  *ristretto.Cache[string, interface{}]
	config Config
}

// cacheEntry holds cached data with timestamp
type cacheEntry struct {
	Data      interface{}
	Timestamp time.Time
}

// NewManager creates a new cache manager
func NewManager(config Config) (Manager, error) {
	maxCost := config.MaxSizeMB * 1024 * 1024
	if maxCost == 0 {
		maxCost = 50 * 1024 * 1024 // 50MB default
	}

	cache, err := ristretto.NewCache(&ristretto.Config[string, interface{}]{
		NumCounters: 1000,
		MaxCost:     maxCost,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &ristrettoManager{
		cache:  cache,
		config: config,
	}, nil
}

// GetCategories retrieves cached categories
func (m *ristrettoManager) GetCategories() interface{} {
	if !m.config.Enabled {
		return nil
	}

	entry, found := m.cache.Get("categories")
	if !found {
		return nil
	}

	cacheEntry, ok := entry.(*cacheEntry)
	if !ok {
		return nil
	}

	if m.isEntryExpired(cacheEntry) {
		m.cache.Del("categories")
		return nil
	}

	// Return the cached data - let caller handle type assertion
	return cacheEntry.Data
}

// SetCategories caches categories
func (m *ristrettoManager) SetCategories(categories interface{}) {
	if !m.config.Enabled {
		return
	}

	entry := &cacheEntry{
		Data:      categories,
		Timestamp: time.Now(),
	}

	m.cache.Set("categories", entry, 1)
}

// GetAgents retrieves cached agents for a category
func (m *ristrettoManager) GetAgents(category string) interface{} {
	if !m.config.Enabled {
		return nil
	}

	key := fmt.Sprintf("agents:%s", category)
	entry, found := m.cache.Get(key)
	if !found {
		return nil
	}

	cacheEntry, ok := entry.(*cacheEntry)
	if !ok {
		return nil
	}

	if m.isEntryExpired(cacheEntry) {
		m.cache.Del(key)
		return nil
	}

	// Return the cached data - let caller handle type assertion
	return cacheEntry.Data
}

// SetAgents caches agents for a category
func (m *ristrettoManager) SetAgents(category string, agents interface{}) {
	if !m.config.Enabled {
		return
	}

	key := fmt.Sprintf("agents:%s", category)
	entry := &cacheEntry{
		Data:      agents,
		Timestamp: time.Now(),
	}

	m.cache.Set(key, entry, 1)
}

// GetAgent retrieves a cached agent by ID
func (m *ristrettoManager) GetAgent(agentID string) interface{} {
	if !m.config.Enabled {
		return nil
	}

	key := fmt.Sprintf("agent:%s", agentID)
	entry, found := m.cache.Get(key)
	if !found {
		return nil
	}

	cacheEntry, ok := entry.(*cacheEntry)
	if !ok {
		return nil
	}

	if m.isEntryExpired(cacheEntry) {
		m.cache.Del(key)
		return nil
	}

	// Return the cached data - let caller handle type assertion
	return cacheEntry.Data
}

// SetAgent caches an individual agent
func (m *ristrettoManager) SetAgent(agentID string, agent interface{}) {
	if !m.config.Enabled {
		return
	}

	key := fmt.Sprintf("agent:%s", agentID)
	entry := &cacheEntry{
		Data:      agent,
		Timestamp: time.Now(),
	}

	m.cache.Set(key, entry, 1)
}

// Clear removes all cached data
func (m *ristrettoManager) Clear() {
	m.cache.Clear()
}

// IsExpired checks if a cache key is expired
func (m *ristrettoManager) IsExpired(key string) bool {
	entry, found := m.cache.Get(key)
	if !found {
		return true
	}

	cacheEntry, ok := entry.(*cacheEntry)
	if !ok {
		return true
	}

	return m.isEntryExpired(cacheEntry)
}

// Size returns the number of items in cache
func (m *ristrettoManager) Size() int {
	// Ristretto doesn't provide a direct size method
	// This is an approximation based on metrics
	metrics := m.cache.Metrics
	return int(metrics.KeysAdded() - metrics.KeysEvicted())
}

// isEntryExpired checks if a cache entry is expired
func (m *ristrettoManager) isEntryExpired(entry *cacheEntry) bool {
	return time.Since(entry.Timestamp) > m.config.TTL
}
