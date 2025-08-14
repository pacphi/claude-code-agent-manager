package tracker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Tracker manages installation tracking
type Tracker struct {
	filePath string
	mu       sync.RWMutex
}

// Installation represents an installed source
type Installation struct {
	Timestamp     time.Time           `json:"timestamp"`
	SourceCommit  string              `json:"source_commit,omitempty"`
	Files         map[string]FileInfo `json:"files"`
	Directories   []string            `json:"directories"`
	DocsGenerated []string            `json:"docs_generated,omitempty"`
}

// FileInfo contains information about an installed file
type FileInfo struct {
	Path           string    `json:"path"`
	Hash           string    `json:"hash,omitempty"`
	Size           int64     `json:"size"`
	Modified       time.Time `json:"modified"`
	WasPreExisting bool      `json:"was_pre_existing,omitempty"`
}

// TrackingData represents the complete tracking data
type TrackingData struct {
	Version       string                   `json:"version"`
	LastUpdated   time.Time                `json:"last_updated"`
	Installations map[string]*Installation `json:"installations"`
}

// New creates a new tracker
func New(filePath string) *Tracker {
	return &Tracker{
		filePath: filePath,
	}
}

// RecordInstallation records a new installation
func (t *Tracker) RecordInstallation(sourceName string, installation Installation) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Load existing data
	data, err := t.load()
	if err != nil {
		// If file doesn't exist, create new tracking data
		if os.IsNotExist(err) {
			data = &TrackingData{
				Version:       "1.0",
				Installations: make(map[string]*Installation),
			}
		} else {
			return fmt.Errorf("failed to load tracking data: %w", err)
		}
	}

	// Update installation
	installation.Timestamp = time.Now()
	data.Installations[sourceName] = &installation
	data.LastUpdated = time.Now()

	// Save data
	return t.save(data)
}

// GetInstallation retrieves installation information for a source
func (t *Tracker) GetInstallation(sourceName string) (*Installation, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	data, err := t.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load tracking data: %w", err)
	}

	installation, exists := data.Installations[sourceName]
	if !exists {
		return nil, fmt.Errorf("installation not found: %s", sourceName)
	}

	return installation, nil
}

// RemoveInstallation removes installation tracking for a source
func (t *Tracker) RemoveInstallation(sourceName string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := t.load()
	if err != nil {
		return fmt.Errorf("failed to load tracking data: %w", err)
	}

	delete(data.Installations, sourceName)
	data.LastUpdated = time.Now()

	return t.save(data)
}

// List returns all installations
func (t *Tracker) List() (map[string]*Installation, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	data, err := t.load()
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*Installation), nil
		}
		return nil, fmt.Errorf("failed to load tracking data: %w", err)
	}

	return data.Installations, nil
}

// Clear removes all tracking data
func (t *Tracker) Clear() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data := &TrackingData{
		Version:       "1.0",
		LastUpdated:   time.Now(),
		Installations: make(map[string]*Installation),
	}

	return t.save(data)
}

// IsInstalled checks if a source is installed
func (t *Tracker) IsInstalled(sourceName string) bool {
	installation, err := t.GetInstallation(sourceName)
	return err == nil && installation != nil
}

// GetInstalledFiles returns all files installed by a source
func (t *Tracker) GetInstalledFiles(sourceName string) ([]string, error) {
	installation, err := t.GetInstallation(sourceName)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(installation.Files))
	for path := range installation.Files {
		files = append(files, path)
	}

	return files, nil
}

// UpdateFile updates tracking for a single file
func (t *Tracker) UpdateFile(sourceName, filePath string, info FileInfo) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := t.load()
	if err != nil {
		return fmt.Errorf("failed to load tracking data: %w", err)
	}

	installation, exists := data.Installations[sourceName]
	if !exists {
		return fmt.Errorf("installation not found: %s", sourceName)
	}

	installation.Files[filePath] = info
	data.LastUpdated = time.Now()

	return t.save(data)
}

// Private methods

func (t *Tracker) load() (*TrackingData, error) {
	// Check if file exists
	if _, err := os.Stat(t.filePath); os.IsNotExist(err) {
		return nil, err
	}

	// Read file
	content, err := os.ReadFile(t.filePath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var data TrackingData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse tracking data: %w", err)
	}

	// Initialize map if nil
	if data.Installations == nil {
		data.Installations = make(map[string]*Installation)
	}

	return &data, nil
}

func (t *Tracker) save(data *TrackingData) error {
	// Ensure parent directory exists
	dir := filepath.Dir(t.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create tracking directory: %w", err)
	}

	// Marshal to JSON with indentation
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tracking data: %w", err)
	}

	// Write atomically using temp file
	tempFile := t.filePath + ".tmp"
	if err := os.WriteFile(tempFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write tracking data: %w", err)
	}

	// Rename temp file to actual file
	if err := os.Rename(tempFile, t.filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to save tracking data: %w", err)
	}

	return nil
}

// Backup creates a backup of the current tracking data
func (t *Tracker) Backup() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Check if tracking file exists
	if _, err := os.Stat(t.filePath); os.IsNotExist(err) {
		return nil // Nothing to backup
	}

	// Read current data
	content, err := os.ReadFile(t.filePath)
	if err != nil {
		return fmt.Errorf("failed to read tracking data: %w", err)
	}

	// Create backup filename with timestamp
	backupPath := fmt.Sprintf("%s.backup.%s", t.filePath, time.Now().Format("20060102-150405"))

	// Write backup
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

// Restore restores tracking data from a backup
func (t *Tracker) Restore(backupPath string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Read backup file
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Validate it's valid JSON
	var data TrackingData
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Write to tracking file
	if err := os.WriteFile(t.filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}
