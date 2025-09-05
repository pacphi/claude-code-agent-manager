package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCacheManager(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 100,
		TTL:     time.Hour,
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)
	assert.NotNil(t, cm)
	assert.Equal(t, 100, cm.config.MaxSize)
	assert.Equal(t, time.Hour, cm.config.TTL)
}

func TestCacheManager_SetGet(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 10,
		TTL:     time.Hour,
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	// Test basic set/get
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": []string{"a", "b", "c"},
		"key3": 42,
	}

	for key, value := range testData {
		cm.Set(key, value)
		retrieved := cm.Get(key)
		assert.Equal(t, value, retrieved)
	}

	// Test non-existent key
	assert.Nil(t, cm.Get("nonexistent"))
}

func TestCacheManager_TTL(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 10,
		TTL:     100 * time.Millisecond, // Very short TTL for testing
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	// Set a value
	cm.Set("expire-test", "will-expire")
	assert.Equal(t, "will-expire", cm.Get("expire-test"))

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Value should be expired
	assert.Nil(t, cm.Get("expire-test"))
}

func TestCacheManager_MaxSize(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 3, // Very small cache
		TTL:     time.Hour,
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	// Fill cache to capacity
	cm.Set("key1", "value1")
	cm.Set("key2", "value2")
	cm.Set("key3", "value3")

	// All values should be present
	assert.Equal(t, "value1", cm.Get("key1"))
	assert.Equal(t, "value2", cm.Get("key2"))
	assert.Equal(t, "value3", cm.Get("key3"))

	// Add one more item, should evict oldest
	cm.Set("key4", "value4")

	// key1 should be evicted (oldest), others should remain
	assert.Nil(t, cm.Get("key1"))
	assert.Equal(t, "value2", cm.Get("key2"))
	assert.Equal(t, "value3", cm.Get("key3"))
	assert.Equal(t, "value4", cm.Get("key4"))
}

func TestCacheManager_Clear(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 10,
		TTL:     time.Hour,
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	// Add some items
	cm.Set("key1", "value1")
	cm.Set("key2", "value2")

	assert.Equal(t, "value1", cm.Get("key1"))
	assert.Equal(t, "value2", cm.Get("key2"))

	// Clear cache
	cm.Clear()

	// All items should be gone
	assert.Nil(t, cm.Get("key1"))
	assert.Nil(t, cm.Get("key2"))
}

func TestCacheManager_Stats(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 10,
		TTL:     time.Hour,
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	// Initial stats
	stats := cm.Stats()
	assert.Equal(t, 0, stats["size"])
	assert.Equal(t, 0, stats["hits"])
	assert.Equal(t, 0, stats["misses"])

	// Add items and check stats
	cm.Set("key1", "value1")
	cm.Set("key2", "value2")

	stats = cm.Stats()
	assert.Equal(t, 2, stats["size"])

	// Test hits and misses
	cm.Get("key1")        // hit
	cm.Get("key2")        // hit
	cm.Get("nonexistent") // miss

	stats = cm.Stats()
	assert.Equal(t, 2, stats["hits"])
	assert.Equal(t, 1, stats["misses"])
	assert.Greater(t, stats["hit_rate"].(float64), 0.0)
}

func TestCacheManager_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 10,
		TTL:     time.Hour,
	}

	// Create cache and add data
	cm1, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	cm1.Set("persistent1", "value1")
	cm1.Set("persistent2", "value2")

	// Save to disk
	err = cm1.Save()
	require.NoError(t, err)

	// Create new cache manager with same path
	cm2, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	// Data should be loaded
	assert.Equal(t, "value1", cm2.Get("persistent1"))
	assert.Equal(t, "value2", cm2.Get("persistent2"))
}

func TestCacheManager_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 100,
		TTL:     time.Hour,
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	// Test concurrent writes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)
				cm.Set(key, value)
				retrieved := cm.Get(key)
				assert.Equal(t, value, retrieved)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Cache should have all items
	stats := cm.Stats()
	assert.Equal(t, 100, stats["size"])
}

func TestCacheManager_AutoCleanup(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize:       10,
		TTL:           50 * time.Millisecond, // Very short TTL
		CleanupPeriod: 25 * time.Millisecond, // Frequent cleanup
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	// Add items
	cm.Set("cleanup1", "value1")
	cm.Set("cleanup2", "value2")

	// Items should be present
	assert.Equal(t, "value1", cm.Get("cleanup1"))
	assert.Equal(t, "value2", cm.Get("cleanup2"))

	// Wait for cleanup to run
	time.Sleep(100 * time.Millisecond)

	// Items should be cleaned up
	assert.Nil(t, cm.Get("cleanup1"))
	assert.Nil(t, cm.Get("cleanup2"))

	stats := cm.Stats()
	assert.Equal(t, 0, stats["size"])
}

func TestCacheManager_InvalidPath(t *testing.T) {
	// Try to create cache in non-existent directory without creating it
	invalidPath := "/nonexistent/path/cache.json"

	config := Config{
		MaxSize: 10,
		TTL:     time.Hour,
	}

	_, err := NewCacheManager(invalidPath, config)
	// Should still work but won't be able to persist
	assert.NoError(t, err) // We handle this gracefully
}

func TestCacheManager_CorruptedFile(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	// Create a corrupted cache file
	err := os.WriteFile(cachePath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	config := Config{
		MaxSize: 10,
		TTL:     time.Hour,
	}

	// Should handle corrupted file gracefully
	cm, err := NewCacheManager(cachePath, config)
	require.NoError(t, err)

	// Should start with empty cache
	assert.Nil(t, cm.Get("any-key"))

	stats := cm.Stats()
	assert.Equal(t, 0, stats["size"])
}

// Benchmarks
func BenchmarkCacheManager_Set(b *testing.B) {
	tempDir := b.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 1000,
		TTL:     time.Hour,
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench-key-%d", i)
		cm.Set(key, "benchmark-value")
	}
}

func BenchmarkCacheManager_Get(b *testing.B) {
	tempDir := b.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := Config{
		MaxSize: 1000,
		TTL:     time.Hour,
	}

	cm, err := NewCacheManager(cachePath, config)
	require.NoError(b, err)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("bench-key-%d", i)
		cm.Set(key, fmt.Sprintf("benchmark-value-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench-key-%d", i%100)
		cm.Get(key)
	}
}
