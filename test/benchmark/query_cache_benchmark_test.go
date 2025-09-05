package benchmark

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/query/cache"
	"github.com/stretchr/testify/require"
)

func BenchmarkCacheManager_Set(b *testing.B) {
	tempDir := b.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	config := cache.Config{
		MaxSize: 1000,
		TTL:     time.Hour,
	}

	cm, err := cache.NewCacheManager(cachePath, config)
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

	config := cache.Config{
		MaxSize: 1000,
		TTL:     time.Hour,
	}

	cm, err := cache.NewCacheManager(cachePath, config)
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
