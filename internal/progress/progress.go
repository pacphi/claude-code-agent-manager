package progress

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Manager manages progress indicators for the application
type Manager struct {
	enabled      bool
	verbose      bool
	dryRun       bool
	noColor      bool
	bars         map[string]*progressbar.ProgressBar
	spinners     map[string]*progressbar.ProgressBar
	mu           sync.Mutex
	output       io.Writer
	lastCleanup  time.Time
	cleanupQueue []string // Queue of IDs to cleanup
}

// Options configures the progress manager
type Options struct {
	Enabled bool
	Verbose bool
	DryRun  bool
	NoColor bool
	Output  io.Writer
}

// New creates a new progress manager
func New(opts Options) *Manager {
	output := opts.Output
	if output == nil {
		output = os.Stdout
	}

	// Disable progress in non-TTY environments
	if opts.Enabled && !isTerminal() {
		opts.Enabled = false
	}

	return &Manager{
		enabled:      opts.Enabled,
		verbose:      opts.Verbose,
		dryRun:       opts.DryRun,
		noColor:      opts.NoColor,
		bars:         make(map[string]*progressbar.ProgressBar),
		spinners:     make(map[string]*progressbar.ProgressBar),
		output:       output,
		lastCleanup:  time.Now(),
		cleanupQueue: make([]string, 0, 10),
	}
}

// StartSpinner starts an indeterminate progress spinner
func (m *Manager) StartSpinner(id, description string) {
	if !m.enabled || m.verbose {
		// In verbose mode, just print the description
		if m.verbose {
			_, _ = fmt.Fprintf(m.output, "%s...\n", description)
		}
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop any existing spinner with this ID
	if existing, ok := m.spinners[id]; ok {
		_ = existing.Finish()
	}

	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(m.output),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionEnableColorCodes(!m.noColor),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
	)

	m.spinners[id] = bar
}

// StopSpinner stops a spinner and optionally shows a completion message
func (m *Manager) StopSpinner(id string, success bool, message string) {
	if !m.enabled {
		if m.verbose && message != "" {
			_, _ = fmt.Fprintf(m.output, "%s\n", message)
		}
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bar, ok := m.spinners[id]; ok {
		_ = bar.Finish()
		delete(m.spinners, id)

		if message != "" {
			if success {
				_, _ = fmt.Fprintf(m.output, "✓ %s\n", message)
			} else {
				_, _ = fmt.Fprintf(m.output, "✗ %s\n", message)
			}
		}
	}

	// Trigger cleanup if needed
	m.conditionalCleanup()
}

// StartProgress starts a determinate progress bar
func (m *Manager) StartProgress(id, description string, total int) {
	if !m.enabled || m.verbose {
		if m.verbose {
			_, _ = fmt.Fprintf(m.output, "%s (0/%d)\n", description, total)
		}
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop any existing progress bar with this ID
	if existing, ok := m.bars[id]; ok {
		_ = existing.Finish()
	}

	bar := progressbar.NewOptions(total,
		progressbar.OptionSetWriter(m.output),
		progressbar.OptionSetDescription(description),
		progressbar.OptionEnableColorCodes(!m.noColor),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	m.bars[id] = bar
}

// UpdateProgress updates a progress bar
func (m *Manager) UpdateProgress(id string, increment int) {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bar, ok := m.bars[id]; ok {
		_ = bar.Add(increment)
	}
}

// UpdateDescription updates the description of a progress bar or spinner
func (m *Manager) UpdateDescription(id, description string) {
	if !m.enabled {
		if m.verbose {
			_, _ = fmt.Fprintf(m.output, "%s\n", description)
		}
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bar, ok := m.bars[id]; ok {
		bar.Describe(description)
	} else if spinner, ok := m.spinners[id]; ok {
		spinner.Describe(description)
	}
}

// FinishProgress completes a progress bar
func (m *Manager) FinishProgress(id string, success bool, message string) {
	if !m.enabled {
		if m.verbose && message != "" {
			_, _ = fmt.Fprintf(m.output, "%s\n", message)
		}
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bar, ok := m.bars[id]; ok {
		_ = bar.Finish()
		delete(m.bars, id)

		if message != "" {
			if success {
				_, _ = fmt.Fprintf(m.output, "✓ %s\n", message)
			} else {
				_, _ = fmt.Fprintf(m.output, "✗ %s\n", message)
			}
		}
	}

	// Trigger cleanup if needed
	m.conditionalCleanup()
}

// StopAll stops all active progress indicators
func (m *Manager) StopAll() {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, bar := range m.bars {
		_ = bar.Finish()
		delete(m.bars, id)
	}

	for id, spinner := range m.spinners {
		_ = spinner.Finish()
		delete(m.spinners, id)
	}
}

// WithSpinner executes a function with a spinner
func (m *Manager) WithSpinner(description string, fn func() error) error {
	spinnerID := fmt.Sprintf("spinner-%d", time.Now().UnixNano())
	m.StartSpinner(spinnerID, description)

	err := fn()

	if err != nil {
		m.StopSpinner(spinnerID, false, fmt.Sprintf("Failed: %s", description))
	} else {
		m.StopSpinner(spinnerID, true, fmt.Sprintf("Completed: %s", description))
	}

	return err
}

// WithProgress executes a function with a progress bar
func (m *Manager) WithProgress(description string, total int, fn func(update func(int)) error) error {
	progressID := fmt.Sprintf("progress-%d", time.Now().UnixNano())
	m.StartProgress(progressID, description, total)

	err := fn(func(increment int) {
		m.UpdateProgress(progressID, increment)
	})

	if err != nil {
		m.FinishProgress(progressID, false, fmt.Sprintf("Failed: %s", description))
	} else {
		m.FinishProgress(progressID, true, fmt.Sprintf("Completed: %s", description))
	}

	return err
}

// isTerminal checks if output is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// Global instance for convenience
var defaultManager *Manager

// Initialize sets up the global progress manager
func Initialize(opts Options) {
	defaultManager = New(opts)
}

// Default returns the default progress manager
func Default() *Manager {
	if defaultManager == nil {
		defaultManager = New(Options{Enabled: true})
	}
	return defaultManager
}

// conditionalCleanup performs cleanup if conditions are met (must be called with mutex held)
func (m *Manager) conditionalCleanup() {
	now := time.Now()

	// Only cleanup every 30 seconds to avoid excessive overhead
	if now.Sub(m.lastCleanup) < 30*time.Second {
		return
	}

	m.lastCleanup = now

	// Clean up any stale progress indicators
	// In practice, this should rarely happen due to proper lifecycle management
	// but provides protection against memory leaks in edge cases
	totalEntries := len(m.bars) + len(m.spinners)
	if totalEntries > 100 { // Only cleanup if we have excessive entries
		m.performCleanup()
	}
}

// performCleanup removes stale progress indicators (must be called with mutex held)
func (m *Manager) performCleanup() {
	// For progress bars and spinners created with timestamp IDs,
	// we can identify old ones by parsing the timestamp
	cutoff := time.Now().Add(-5 * time.Minute) // Consider entries older than 5 minutes as stale

	// Clean up old spinners
	for id := range m.spinners {
		if m.isStaleID(id, cutoff) {
			if bar := m.spinners[id]; bar != nil {
				_ = bar.Finish()
			}
			delete(m.spinners, id)
		}
	}

	// Clean up old progress bars
	for id := range m.bars {
		if m.isStaleID(id, cutoff) {
			if bar := m.bars[id]; bar != nil {
				_ = bar.Finish()
			}
			delete(m.bars, id)
		}
	}
}

// isStaleID checks if an ID represents a stale timestamp-based entry
func (m *Manager) isStaleID(id string, cutoff time.Time) bool {
	// Check if this looks like a timestamp-based ID (from WithSpinner/WithProgress)
	if len(id) > 10 && (id[:8] == "spinner-" || id[:9] == "progress-") {
		// Extract timestamp part
		var timestampStr string
		if id[:8] == "spinner-" {
			timestampStr = id[8:]
		} else if id[:9] == "progress-" {
			timestampStr = id[9:]
		}

		// Try to parse as Unix nano timestamp
		if nanos, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
			timestamp := time.Unix(0, nanos)
			return timestamp.Before(cutoff)
		}
	}

	return false
}

// GetMemoryStats returns memory usage statistics for monitoring
func (m *Manager) GetMemoryStats() (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.bars), len(m.spinners)
}
