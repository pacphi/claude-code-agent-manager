package service

import (
	"context"
	"errors"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/browser"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/cache"
	"github.com/pacphi/claude-code-agent-manager/internal/types"
)

func TestMarketplaceService_GetCategories(t *testing.T) {
	tests := []struct {
		name           string
		cacheEnabled   bool
		cacheData      []types.Category
		extractorError error
		wantErr        bool
		wantCategories int
	}{
		{
			name:         "success_with_cache_hit",
			cacheEnabled: true,
			cacheData: []types.Category{
				{ID: "ai-ml", Name: "AI & ML", Slug: "ai-ml", AgentCount: 5},
			},
			wantCategories: 1,
		},
		{
			name:           "success_with_cache_miss",
			cacheEnabled:   true,
			cacheData:      nil,
			wantCategories: 1,
		},
		{
			name:           "success_without_cache",
			cacheEnabled:   false,
			wantCategories: 1,
		},
		{
			name:           "extractor_error",
			cacheEnabled:   false,
			extractorError: errors.New("extraction failed"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockBrowser := browser.NewMockController()
			mockCache := cache.NewMockManager()
			mockExtractor := &mockCategoryExtractor{
				err: tt.extractorError,
			}

			// Configure cache
			if tt.cacheEnabled && tt.cacheData != nil {
				mockCache.SetCategories(tt.cacheData)
			} else {
				mockCache.SetDisabled(!tt.cacheEnabled)
			}

			service := &marketplaceService{
				browser: mockBrowser,
				cache:   mockCache,
				extractors: ExtractorSet{
					Categories: mockExtractor,
				},
				baseURL: "https://test.com",
				config: Config{
					CacheEnabled: tt.cacheEnabled,
				},
			}

			// Execute
			categories, err := service.GetCategories(context.Background())

			// Verify
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(categories) != tt.wantCategories {
				t.Errorf("Expected %d categories, got %d", tt.wantCategories, len(categories))
			}

			// Verify cache interaction
			if tt.cacheEnabled && tt.cacheData == nil {
				// Should have called SetCategories for cache miss
				calls := mockCache.GetCallCounts()
				if calls["SetCategories"] != 1 {
					t.Errorf("Expected 1 SetCategories call, got %d", calls["SetCategories"])
				}
			}
		})
	}
}

func TestMarketplaceService_GetAgents(t *testing.T) {
	tests := []struct {
		name       string
		category   string
		cacheData  []types.Agent
		wantErr    bool
		wantAgents int
	}{
		{
			name:       "success_with_valid_category",
			category:   "ai-ml",
			wantAgents: 1,
		},
		{
			name:     "error_with_empty_category",
			category: "",
			wantErr:  true,
		},
		{
			name:     "success_with_cache_hit",
			category: "dev",
			cacheData: []types.Agent{
				{ID: "code-reviewer", Name: "Code Reviewer", Category: "dev"},
			},
			wantAgents: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBrowser := browser.NewMockController()
			mockCache := cache.NewMockManager()
			mockExtractor := &mockAgentExtractor{}

			// Setup cache data
			if tt.cacheData != nil {
				mockCache.SetAgents(tt.category, tt.cacheData)
			}

			service := &marketplaceService{
				browser: mockBrowser,
				cache:   mockCache,
				extractors: ExtractorSet{
					Agents: mockExtractor,
				},
				baseURL: "https://test.com",
				config: Config{
					CacheEnabled: true,
				},
			}

			// Execute
			agents, err := service.GetAgents(context.Background(), tt.category)

			// Verify
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(agents) != tt.wantAgents {
				t.Errorf("Expected %d agents, got %d", tt.wantAgents, len(agents))
			}
		})
	}
}

// Mock implementations

type mockCategoryExtractor struct {
	categories []types.Category
	err        error
}

func (m *mockCategoryExtractor) Extract(ctx context.Context, browser browser.Controller) ([]types.Category, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.categories != nil {
		return m.categories, nil
	}

	// Default test data
	return []types.Category{
		{ID: "ai-ml", Name: "AI & ML", Slug: "ai-ml", AgentCount: 5},
	}, nil
}

type mockAgentExtractor struct {
	agents map[string][]types.Agent
	err    error
}

func (m *mockAgentExtractor) Extract(ctx context.Context, browser browser.Controller, category string) ([]types.Agent, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.agents != nil {
		if agents, exists := m.agents[category]; exists {
			return agents, nil
		}
		return []types.Agent{}, nil
	}

	// Default test data
	return []types.Agent{
		{ID: "code-reviewer", Name: "Code Reviewer", Category: category},
	}, nil
}
