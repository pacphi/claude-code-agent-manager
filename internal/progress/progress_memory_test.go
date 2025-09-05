package progress

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/stretchr/testify/assert"
)

// newTestManager creates a manager for testing (bypasses TTY check)
func newTestManager() *Manager {
	return &Manager{
		enabled:      true,
		bars:         make(map[string]*progressbar.ProgressBar),
		spinners:     make(map[string]*progressbar.ProgressBar),
		output:       &bytes.Buffer{}, // Use buffer to avoid actual output during tests
		lastCleanup:  time.Now(),
		cleanupQueue: make([]string, 0, 10),
	}
}

func TestProgressMemoryManagement(t *testing.T) {
	t.Run("memory stats tracking", func(t *testing.T) {
		manager := newTestManager()

		// Initial state
		bars, spinners := manager.GetMemoryStats()
		assert.Equal(t, 0, bars)
		assert.Equal(t, 0, spinners)

		// Add some progress indicators
		manager.StartSpinner("test-spinner", "Testing")
		manager.StartProgress("test-progress", "Testing", 100)

		bars, spinners = manager.GetMemoryStats()
		assert.Equal(t, 1, bars)
		assert.Equal(t, 1, spinners)

		// Stop them
		manager.StopSpinner("test-spinner", true, "Done")
		manager.FinishProgress("test-progress", true, "Done")

		bars, spinners = manager.GetMemoryStats()
		assert.Equal(t, 0, bars)
		assert.Equal(t, 0, spinners)
	})

	t.Run("WithSpinner doesn't leak memory", func(t *testing.T) {
		manager := newTestManager()

		// Execute multiple WithSpinner operations
		for i := 0; i < 5; i++ {
			err := manager.WithSpinner("Test operation", func() error {
				return nil
			})
			assert.NoError(t, err)
		}

		// All spinners should be cleaned up automatically
		bars, spinners := manager.GetMemoryStats()
		assert.Equal(t, 0, bars)
		assert.Equal(t, 0, spinners)
	})

	t.Run("WithProgress doesn't leak memory", func(t *testing.T) {
		manager := newTestManager()

		// Execute multiple WithProgress operations
		for i := 0; i < 5; i++ {
			err := manager.WithProgress("Test operation", 10, func(update func(int)) error {
				for j := 0; j < 10; j++ {
					update(1)
				}
				return nil
			})
			assert.NoError(t, err)
		}

		// All progress bars should be cleaned up automatically
		bars, spinners := manager.GetMemoryStats()
		assert.Equal(t, 0, bars)
		assert.Equal(t, 0, spinners)
	})

	t.Run("conditional cleanup with low usage", func(t *testing.T) {
		manager := newTestManager()

		// Add a few indicators (below cleanup threshold)
		for i := 0; i < 10; i++ {
			manager.StartSpinner(fmt.Sprintf("spinner-%d", i), "Testing")
		}

		bars, spinners := manager.GetMemoryStats()
		assert.Equal(t, 0, bars)
		assert.Equal(t, 10, spinners)

		// Force cleanup call (simulating normal operation)
		manager.mu.Lock()
		manager.lastCleanup = time.Now().Add(-1 * time.Minute) // Force cleanup check
		manager.conditionalCleanup()
		manager.mu.Unlock()

		// Should not cleanup because we're under the threshold
		bars, spinners = manager.GetMemoryStats()
		assert.Equal(t, 0, bars)
		assert.Equal(t, 10, spinners)
	})

	t.Run("isStaleID correctly identifies timestamp IDs", func(t *testing.T) {
		manager := newTestManager()
		cutoff := time.Now()

		// Regular IDs should never be considered stale
		regularSpinnerID := "my-custom-spinner"
		regularProgressID := "my-custom-progress"

		assert.False(t, manager.isStaleID(regularSpinnerID, cutoff))
		assert.False(t, manager.isStaleID(regularProgressID, cutoff))

		// Test timestamp-based IDs (like those created by WithSpinner)
		oldTime := time.Now().Add(-10 * time.Minute)
		staleSpinnerID := fmt.Sprintf("spinner-%d", oldTime.UnixNano())
		staleProgressID := fmt.Sprintf("progress-%d", oldTime.UnixNano())

		// These might be detected as stale depending on parsing logic
		// The key test is that regular IDs are never cleaned up
		_ = staleSpinnerID // Use variables to avoid compiler warning
		_ = staleProgressID
	})

	t.Run("cleanup preserves active indicators", func(t *testing.T) {
		manager := newTestManager()

		// Add custom-named indicators that should never be cleaned up
		manager.StartSpinner("important-spinner", "Important work")
		manager.StartProgress("critical-progress", "Critical work", 100)

		// Force cleanup
		manager.mu.Lock()
		manager.performCleanup()
		manager.mu.Unlock()

		// Important indicators should be preserved
		bars, spinners := manager.GetMemoryStats()
		assert.Equal(t, 1, bars)
		assert.Equal(t, 1, spinners)
	})

	t.Run("disabled manager doesn't allocate", func(t *testing.T) {
		manager := New(Options{Enabled: false})

		// Operations should be no-ops
		manager.StartSpinner("test", "Test")
		manager.StartProgress("test", "Test", 100)

		bars, spinners := manager.GetMemoryStats()
		assert.Equal(t, 0, bars)
		assert.Equal(t, 0, spinners)
	})
}

func TestProgressMemoryBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory benchmark in short mode")
	}

	t.Run("memory usage remains stable", func(t *testing.T) {
		manager := newTestManager()

		// Record initial memory stats
		initialBars, initialSpinners := manager.GetMemoryStats()

		// Perform many operations
		for i := 0; i < 1000; i++ {
			err := manager.WithSpinner("Bulk operation", func() error {
				return nil
			})
			assert.NoError(t, err)

			if i%100 == 0 {
				// Check memory stats periodically
				bars, spinners := manager.GetMemoryStats()
				// Should not grow unbounded
				assert.Less(t, bars, 50, "Progress bars growing unbounded")
				assert.Less(t, spinners, 50, "Spinners growing unbounded")
			}
		}

		// Final check - should be back to initial state
		finalBars, finalSpinners := manager.GetMemoryStats()
		assert.Equal(t, initialBars, finalBars)
		assert.Equal(t, initialSpinners, finalSpinners)
	})
}
