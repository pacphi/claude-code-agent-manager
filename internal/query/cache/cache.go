package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config holds cache configuration
type Config struct {
	MaxSize       int           // Maximum number of entries
	TTL           time.Duration // Time to live for entries
	CleanupPeriod time.Duration // How often to run cleanup (defaults to TTL/4)
}

// Entry represents a cached entry with metadata
type Entry struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	CreatedAt  time.Time   `json:"created_at"`
	AccessedAt time.Time   `json:"accessed_at"`
}

// CacheManager provides TTL-based caching with thread-safety and persistence
type CacheManager struct {
	entries map[string]*Entry
	config  Config
	path    string
	mu      sync.RWMutex
	stats   CacheStats
	cleanup *time.Ticker
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits   int `json:"hits"`
	Misses int `json:"misses"`
	Size   int `json:"size"`
}

// NewCacheManager creates a new cache manager with the specified configuration
func NewCacheManager(path string, config Config) (*CacheManager, error) {
	// Set default cleanup period if not specified
	if config.CleanupPeriod == 0 {
		config.CleanupPeriod = config.TTL / 4
		if config.CleanupPeriod < time.Minute {
			config.CleanupPeriod = time.Minute
		}
	}

	cm := &CacheManager{
		entries: make(map[string]*Entry),
		config:  config,
		path:    path,
		stats:   CacheStats{},
	}

	// Try to load existing cache
	if err := cm.load(); err != nil {
		// Log error but continue with empty cache
		fmt.Fprintf(os.Stderr, "Warning: failed to load cache from %s: %v\n", path, err)
	}

	// Start cleanup goroutine if cleanup period is set
	if config.CleanupPeriod > 0 {
		cm.cleanup = time.NewTicker(config.CleanupPeriod)
		go cm.cleanupExpired()
	}

	return cm, nil
}

// Set stores a value in the cache with the given key
func (cm *CacheManager) Set(key string, value interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()

	// Check if we need to evict due to size limit
	if len(cm.entries) >= cm.config.MaxSize {
		cm.evictOldest()
	}

	// Store the entry
	cm.entries[key] = &Entry{
		Key:        key,
		Value:      value,
		CreatedAt:  now,
		AccessedAt: now,
	}

	cm.stats.Size = len(cm.entries)
}

// Get retrieves a value from the cache by key
func (cm *CacheManager) Get(key string) interface{} {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry, exists := cm.entries[key]
	if !exists {
		cm.stats.Misses++
		return nil
	}

	// Check if entry has expired
	if time.Since(entry.CreatedAt) > cm.config.TTL {
		delete(cm.entries, key)
		cm.stats.Size = len(cm.entries)
		cm.stats.Misses++
		return nil
	}

	// Update access time for LRU
	entry.AccessedAt = time.Now()
	cm.stats.Hits++

	return entry.Value
}

// Clear removes all entries from the cache
func (cm *CacheManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.entries = make(map[string]*Entry)
	cm.stats.Size = 0
}

// Stats returns cache performance statistics
func (cm *CacheManager) Stats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	total := cm.stats.Hits + cm.stats.Misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(cm.stats.Hits) / float64(total)
	}

	return map[string]interface{}{
		"size":     cm.stats.Size,
		"hits":     cm.stats.Hits,
		"misses":   cm.stats.Misses,
		"hit_rate": hitRate,
		"max_size": cm.config.MaxSize,
		"ttl":      cm.config.TTL,
	}
}

// Save persists the cache to disk
func (cm *CacheManager) Save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Create directory if it doesn't exist
	if dir := filepath.Dir(cm.path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory: %w", err)
		}
	}

	// Prepare data for serialization
	data := struct {
		Entries map[string]*Entry `json:"entries"`
		Stats   CacheStats        `json:"stats"`
		Config  Config            `json:"config"`
	}{
		Entries: cm.entries,
		Stats:   cm.stats,
		Config:  cm.config,
	}

	// Write to temporary file first, then rename (atomic operation)
	tempPath := cm.path + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp cache file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close temp cache file: %v\n", closeErr)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode cache data: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, cm.path); err != nil {
		return fmt.Errorf("failed to save cache file: %w", err)
	}

	return nil
}

// load reads the cache from disk
func (cm *CacheManager) load() error {
	file, err := os.Open(cm.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file is OK
		}
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close cache file: %v\n", closeErr)
		}
	}()

	var data struct {
		Entries map[string]*Entry `json:"entries"`
		Stats   CacheStats        `json:"stats"`
		Config  Config            `json:"config"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode cache file: %w", err)
	}

	// Filter out expired entries during load
	now := time.Now()
	validEntries := make(map[string]*Entry)

	for key, entry := range data.Entries {
		if now.Sub(entry.CreatedAt) <= cm.config.TTL {
			validEntries[key] = entry
		}
	}

	cm.entries = validEntries
	cm.stats.Size = len(validEntries)
	// Reset hit/miss stats on load
	cm.stats.Hits = 0
	cm.stats.Misses = 0

	return nil
}

// evictOldest removes the oldest entry based on access time
func (cm *CacheManager) evictOldest() {
	if len(cm.entries) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	// Find oldest entry by access time
	for key, entry := range cm.entries {
		if oldestKey == "" || entry.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.AccessedAt
		}
	}

	if oldestKey != "" {
		delete(cm.entries, oldestKey)
	}
}

// cleanupExpired removes expired entries periodically
func (cm *CacheManager) cleanupExpired() {
	for range cm.cleanup.C {
		cm.mu.Lock()

		now := time.Now()
		expiredKeys := make([]string, 0)

		// Find expired entries
		for key, entry := range cm.entries {
			if now.Sub(entry.CreatedAt) > cm.config.TTL {
				expiredKeys = append(expiredKeys, key)
			}
		}

		// Remove expired entries
		for _, key := range expiredKeys {
			delete(cm.entries, key)
		}

		cm.stats.Size = len(cm.entries)
		cm.mu.Unlock()

		// Save periodically during cleanup
		if len(expiredKeys) > 0 {
			_ = cm.Save() // Ignore errors in background cleanup
		}
	}
}

// Close shuts down the cache manager and saves to disk
func (cm *CacheManager) Close() error {
	if cm.cleanup != nil {
		cm.cleanup.Stop()
	}

	return cm.Save()
}
