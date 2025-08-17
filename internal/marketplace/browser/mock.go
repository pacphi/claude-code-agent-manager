package browser

import (
	"context"
	"sync"
)

// MockController implements Controller interface for testing
type MockController struct {
	mu sync.RWMutex

	// Test data
	NavigateFunc       func(ctx context.Context, url string) error
	ExecuteScriptFunc  func(ctx context.Context, script string) (interface{}, error)
	WaitForElementFunc func(ctx context.Context, selector string) error
	ScrollPageFunc     func(ctx context.Context, offset int) error
	CloseFunc          func() error

	// Call tracking
	NavigateCalls       []string
	ExecuteScriptCalls  []string
	WaitForElementCalls []string
	ScrollPageCalls     []int
	CloseCalls          int

	// State
	Closed bool
}

// NewMockController creates a new mock browser controller
func NewMockController() *MockController {
	return &MockController{
		NavigateCalls:       make([]string, 0),
		ExecuteScriptCalls:  make([]string, 0),
		WaitForElementCalls: make([]string, 0),
		ScrollPageCalls:     make([]int, 0),
	}
}

// Navigate mocks navigation to a URL
func (m *MockController) Navigate(ctx context.Context, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.NavigateCalls = append(m.NavigateCalls, url)

	if m.NavigateFunc != nil {
		return m.NavigateFunc(ctx, url)
	}

	if m.Closed {
		return ErrBrowserClosed
	}

	return nil
}

// ExecuteScript mocks JavaScript execution
func (m *MockController) ExecuteScript(ctx context.Context, script string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ExecuteScriptCalls = append(m.ExecuteScriptCalls, script)

	if m.ExecuteScriptFunc != nil {
		return m.ExecuteScriptFunc(ctx, script)
	}

	if m.Closed {
		return nil, ErrBrowserClosed
	}

	// Return mock data based on script content
	if contains(script, "categories") {
		return []interface{}{
			map[string]interface{}{
				"name":        "AI & ML",
				"description": "Artificial Intelligence and Machine Learning",
				"agentCount":  5,
				"url":         "https://subagents.sh/categories/ai-ml",
			},
		}, nil
	}

	if contains(script, "agents") {
		return map[string]interface{}{
			"agents": []interface{}{
				map[string]interface{}{
					"name":        "Code Reviewer",
					"description": "Reviews code for quality and bugs",
					"author":      "testuser",
					"rating":      4.5,
					"url":         "https://subagents.sh/agent/code-reviewer",
				},
			},
		}, nil
	}

	return "mock content", nil
}

// WaitForElement mocks waiting for an element
func (m *MockController) WaitForElement(ctx context.Context, selector string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.WaitForElementCalls = append(m.WaitForElementCalls, selector)

	if m.WaitForElementFunc != nil {
		return m.WaitForElementFunc(ctx, selector)
	}

	if m.Closed {
		return ErrBrowserClosed
	}

	return nil
}

// ScrollPage mocks page scrolling
func (m *MockController) ScrollPage(ctx context.Context, offset int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ScrollPageCalls = append(m.ScrollPageCalls, offset)

	if m.ScrollPageFunc != nil {
		return m.ScrollPageFunc(ctx, offset)
	}

	if m.Closed {
		return ErrBrowserClosed
	}

	return nil
}

// Close mocks browser closure
func (m *MockController) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CloseCalls++
	m.Closed = true

	if m.CloseFunc != nil {
		return m.CloseFunc()
	}

	return nil
}

// SetNavigateError sets an error to be returned by Navigate
func (m *MockController) SetNavigateError(err error) {
	m.NavigateFunc = func(ctx context.Context, url string) error {
		return err
	}
}

// SetExecuteScriptError sets an error to be returned by ExecuteScript
func (m *MockController) SetExecuteScriptError(err error) {
	m.ExecuteScriptFunc = func(ctx context.Context, script string) (interface{}, error) {
		return nil, err
	}
}

// SetExecuteScriptResult sets a custom result to be returned by ExecuteScript
func (m *MockController) SetExecuteScriptResult(result interface{}) {
	m.ExecuteScriptFunc = func(ctx context.Context, script string) (interface{}, error) {
		return result, nil
	}
}

// GetNavigateCalls returns all Navigate calls
func (m *MockController) GetNavigateCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]string, len(m.NavigateCalls))
	copy(calls, m.NavigateCalls)
	return calls
}

// GetExecuteScriptCalls returns all ExecuteScript calls
func (m *MockController) GetExecuteScriptCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]string, len(m.ExecuteScriptCalls))
	copy(calls, m.ExecuteScriptCalls)
	return calls
}

// Reset clears all call tracking and resets state
func (m *MockController) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.NavigateCalls = make([]string, 0)
	m.ExecuteScriptCalls = make([]string, 0)
	m.WaitForElementCalls = make([]string, 0)
	m.ScrollPageCalls = make([]int, 0)
	m.CloseCalls = 0
	m.Closed = false

	m.NavigateFunc = nil
	m.ExecuteScriptFunc = nil
	m.WaitForElementFunc = nil
	m.ScrollPageFunc = nil
	m.CloseFunc = nil
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

// findSubstring is a simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
