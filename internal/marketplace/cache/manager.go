package cache

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// ristrettoManager implements Manager interface using ristretto cache
type ristrettoManager struct {
	cache  *ristretto.Cache[string, interface{}]
	config Config
	stats  cacheStats
}

// cacheStats tracks cache performance with atomic operations for thread safety
type cacheStats struct {
	hits   int64
	misses int64
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
		stats:  cacheStats{},
	}, nil
}

// GetCategories retrieves cached categories
func (m *ristrettoManager) GetCategories() interface{} {
	if !m.config.Enabled {
		return nil
	}

	entry, found := m.cache.Get("categories")
	if !found {
		atomic.AddInt64(&m.stats.misses, 1)
		return nil
	}

	cacheEntry, ok := entry.(*cacheEntry)
	if !ok {
		atomic.AddInt64(&m.stats.misses, 1)
		return nil
	}

	if m.isEntryExpired(cacheEntry) {
		m.cache.Del("categories")
		atomic.AddInt64(&m.stats.misses, 1)
		return nil
	}

	atomic.AddInt64(&m.stats.hits, 1)
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
		atomic.AddInt64(&m.stats.misses, 1)
		return nil
	}

	cacheEntry, ok := entry.(*cacheEntry)
	if !ok {
		atomic.AddInt64(&m.stats.misses, 1)
		return nil
	}

	if m.isEntryExpired(cacheEntry) {
		m.cache.Del(key)
		atomic.AddInt64(&m.stats.misses, 1)
		return nil
	}

	atomic.AddInt64(&m.stats.hits, 1)
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
		atomic.AddInt64(&m.stats.misses, 1)
		return nil
	}

	cacheEntry, ok := entry.(*cacheEntry)
	if !ok {
		atomic.AddInt64(&m.stats.misses, 1)
		return nil
	}

	if m.isEntryExpired(cacheEntry) {
		m.cache.Del(key)
		atomic.AddInt64(&m.stats.misses, 1)
		return nil
	}

	atomic.AddInt64(&m.stats.hits, 1)
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

// GetAgentCategory retrieves the cached category for an agent
func (m *ristrettoManager) GetAgentCategory(agentID string) string {
	if !m.config.Enabled {
		return ""
	}

	key := fmt.Sprintf("agent_category:%s", agentID)
	entry, found := m.cache.Get(key)
	if !found {
		atomic.AddInt64(&m.stats.misses, 1)
		return ""
	}

	cacheEntry, ok := entry.(*cacheEntry)
	if !ok {
		atomic.AddInt64(&m.stats.misses, 1)
		return ""
	}

	if m.isEntryExpired(cacheEntry) {
		m.cache.Del(key)
		atomic.AddInt64(&m.stats.misses, 1)
		return ""
	}

	atomic.AddInt64(&m.stats.hits, 1)
	if category, ok := cacheEntry.Data.(string); ok {
		return category
	}
	return ""
}

// SetAgentCategory caches the category for an agent
func (m *ristrettoManager) SetAgentCategory(agentID string, category string) {
	if !m.config.Enabled {
		return
	}

	key := fmt.Sprintf("agent_category:%s", agentID)
	entry := &cacheEntry{
		Data:      category,
		Timestamp: time.Now(),
	}

	m.cache.Set(key, entry, 1)
}

// Clear removes all cached data
func (m *ristrettoManager) Clear() {
	m.cache.Clear()
	// Reset stats when clearing cache
	m.stats = cacheStats{}
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

// GetStats returns cache performance statistics
func (m *ristrettoManager) GetStats() CacheStats {
	size := m.Size()

	// Get atomic values
	hits := atomic.LoadInt64(&m.stats.hits)
	misses := atomic.LoadInt64(&m.stats.misses)

	// Calculate hit rate
	total := hits + misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	// Get evictions from ristretto metrics
	metrics := m.cache.Metrics
	evictions := int64(metrics.KeysEvicted())

	return CacheStats{
		Hits:      hits,
		Misses:    misses,
		Evictions: evictions,
		Size:      size,
		HitRate:   hitRate,
	}
}

// isEntryExpired checks if a cache entry is expired
func (m *ristrettoManager) isEntryExpired(entry *cacheEntry) bool {
	return time.Since(entry.Timestamp) > m.config.TTL
}
