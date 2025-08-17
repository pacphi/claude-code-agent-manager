package progress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Manager manages progress indicators for the application
type Manager struct {
	enabled  bool
	verbose  bool
	dryRun   bool
	noColor  bool
	bars     map[string]*progressbar.ProgressBar
	spinners map[string]*progressbar.ProgressBar
	mu       sync.Mutex
	output   io.Writer
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
		enabled:  opts.Enabled,
		verbose:  opts.Verbose,
		dryRun:   opts.DryRun,
		noColor:  opts.NoColor,
		bars:     make(map[string]*progressbar.ProgressBar),
		spinners: make(map[string]*progressbar.ProgressBar),
		output:   output,
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
