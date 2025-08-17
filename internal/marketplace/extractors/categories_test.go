package extractors

import (
	"context"
	"errors"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/browser"
	"github.com/pacphi/claude-code-agent-manager/internal/types"
)

func TestCategoryExtractor_Extract(t *testing.T) {
	tests := []struct {
		name          string
		browserResult interface{}
		browserError  error
		scriptError   error
		expectError   bool
		expectedCount int
	}{
		{
			name: "success_with_valid_data",
			browserResult: []interface{}{
				map[string]interface{}{
					"name":        "AI & ML",
					"description": "Artificial Intelligence",
					"agentCount":  5,
					"url":         "https://test.com/categories/ai-ml",
				},
				map[string]interface{}{
					"name":        "Development",
					"description": "Software Development",
					"agentCount":  10,
					"url":         "https://test.com/categories/dev",
				},
			},
			expectedCount: 2,
		},
		{
			name:        "error_with_nil_browser",
			expectError: true,
		},
		{
			name:         "error_with_script_execution_failure",
			browserError: browser.ErrScriptExecution,
			expectError:  true,
		},
		{
			name:        "error_with_script_loading_failure",
			scriptError: errors.New("script not found"),
			expectError: true,
		},
		{
			name:          "success_with_empty_result",
			browserResult: []interface{}{},
			expectedCount: 0,
		},
		{
			name:          "success_with_nil_result",
			browserResult: nil,
			expectError:   true,
		},
		{
			name: "success_with_invalid_category_filtered_out",
			browserResult: []interface{}{
				map[string]interface{}{
					"name":        "", // Invalid - empty name
					"description": "Empty name category",
					"agentCount":  5,
					"url":         "https://test.com/categories/empty",
				},
				map[string]interface{}{
					"name":        "Valid Category",
					"description": "This is a valid category",
					"agentCount":  3,
					"url":         "https://test.com/categories/valid",
				},
			},
			expectedCount: 1, // Only the valid category should be included
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock script provider
			mockScript := &mockScriptProvider{
				scriptError: tt.scriptError,
			}

			// Setup mock browser
			var mockBrowser browser.Controller
			if tt.name != "error_with_nil_browser" {
				mock := browser.NewMockController()
				if tt.browserError != nil {
					mock.SetExecuteScriptError(tt.browserError)
				} else {
					mock.SetExecuteScriptResult(tt.browserResult)
				}
				mockBrowser = mock
			}

			// Create extractor
			extractor := NewCategoryExtractor(mockScript)

			// Execute
			categories, err := extractor.Extract(context.Background(), mockBrowser)

			// Verify
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(categories) != tt.expectedCount {
				t.Errorf("Expected %d categories, got %d", tt.expectedCount, len(categories))
			}

			// Verify category data structure
			for _, category := range categories {
				if category.Name == "" {
					t.Errorf("Category name should not be empty")
				}
				if category.Slug == "" {
					t.Errorf("Category slug should not be empty")
				}
				if category.ID == "" {
					t.Errorf("Category ID should not be empty")
				}
			}
		})
	}
}

func TestCategoryExtractor_parseCategories(t *testing.T) {
	extractor := &categoryExtractor{}

	tests := []struct {
		name          string
		input         interface{}
		expectedCount int
		expectError   bool
	}{
		{
			name: "valid_categories_array",
			input: []interface{}{
				map[string]interface{}{
					"name":        "AI & ML",
					"description": "Artificial Intelligence",
					"agentCount":  5,
					"url":         "https://test.com/categories/ai-ml",
				},
			},
			expectedCount: 1,
		},
		{
			name:        "invalid_input_type",
			input:       "not an array",
			expectError: true,
		},
		{
			name:          "empty_array",
			input:         []interface{}{},
			expectedCount: 0,
		},
		{
			name: "invalid_category_objects",
			input: []interface{}{
				"not a map",
				map[string]interface{}{
					"name": "Valid Category",
					"url":  "https://test.com/categories/valid",
				},
			},
			expectedCount: 1, // Only the valid map should be processed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categories, err := extractor.parseCategories(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(categories) != tt.expectedCount {
				t.Errorf("Expected %d categories, got %d", tt.expectedCount, len(categories))
			}
		})
	}
}

func TestCategoryExtractor_validateCategory(t *testing.T) {
	extractor := &categoryExtractor{}

	tests := []struct {
		name     string
		category types.Category
		wantErr  bool
	}{
		{
			name: "valid_category",
			category: types.Category{
				Name:        "AI & ML",
				Slug:        "ai-ml",
				Description: "Artificial Intelligence and Machine Learning",
				AgentCount:  5,
			},
		},
		{
			name: "empty_name",
			category: types.Category{
				Name: "",
				Slug: "test",
			},
			wantErr: true,
		},
		{
			name: "empty_slug",
			category: types.Category{
				Name: "Test",
				Slug: "",
			},
			wantErr: true,
		},
		{
			name: "name_too_long",
			category: types.Category{
				Name: "This is a very long category name that exceeds the maximum allowed length for category names and goes beyond 100 characters",
				Slug: "long-name",
			},
			wantErr: true,
		},
		{
			name: "description_too_long",
			category: types.Category{
				Name: "Test",
				Slug: "test",
				Description: "This is a very long description that exceeds the maximum allowed length for category descriptions. " +
					"It contains way too much text and should be rejected by the validation function because it's longer than 500 characters. " +
					"This is still part of the very long description that we're testing to make sure it gets properly rejected. " +
					"And this is even more text to ensure we definitely exceed the limit and trigger the validation error as expected. " +
					"Adding even more text here to make absolutely sure we go over the 500 character limit for validation testing purposes.",
			},
			wantErr: true,
		},
		{
			name: "negative_agent_count",
			category: types.Category{
				Name:       "Test",
				Slug:       "test",
				AgentCount: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := extractor.validateCategory(tt.category)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

// Mock script provider for testing
type mockScriptProvider struct {
	scriptError error
}

func (m *mockScriptProvider) LoadCategoriesScript() (string, error) {
	if m.scriptError != nil {
		return "", m.scriptError
	}
	return "mock categories script", nil
}

func (m *mockScriptProvider) LoadAgentsScript() (string, error) {
	return "mock agents script", nil
}

func (m *mockScriptProvider) LoadContentScript() (string, error) {
	return "mock content script", nil
}
