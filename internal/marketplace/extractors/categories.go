package extractors

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/browser"
	"github.com/pacphi/claude-code-agent-manager/internal/types"
	"github.com/pacphi/claude-code-agent-manager/internal/util"
)

// categoryExtractor implements CategoryExtractor interface
type categoryExtractor struct {
	scriptProvider ScriptProvider
}

// NewCategoryExtractor creates a new category extractor
func NewCategoryExtractor(scriptProvider ScriptProvider) CategoryExtractor {
	return &categoryExtractor{
		scriptProvider: scriptProvider,
	}
}

// Extract extracts categories from the marketplace page
func (e *categoryExtractor) Extract(ctx context.Context, browser browser.Controller) ([]types.Category, error) {
	if browser == nil {
		return nil, errors.New("browser controller is nil")
	}

	// Load the categories extraction script
	script, err := e.scriptProvider.LoadCategoriesScript()
	if err != nil {
		return nil, fmt.Errorf("script loading failed: %w", err)
	}

	if script == "" {
		return nil, errors.New("categories script is empty")
	}

	// Execute the script with error context
	util.DebugLogf("About to execute JavaScript script (length: %d chars)", len(script))
	result, err := browser.ExecuteScript(ctx, script)
	if err != nil {
		util.DebugLogf("Script execution error: %v", err)
		return nil, fmt.Errorf("script execution failed: %w", err)
	}
	util.DebugLogf("Script execution completed successfully")

	if result == nil {
		util.DebugLogf("Script returned nil result")
		return nil, errors.New("script returned nil result")
	}
	util.DebugLogf("Script returned non-nil result")

	// Parse the results with validation
	categories, err := e.parseCategories(result)
	if err != nil {
		return nil, fmt.Errorf("data parsing failed: %w", err)
	}

	util.DebugLogf("Raw categories parsed: %d", len(categories))
	if len(categories) == 0 {
		util.DebugLogf("Warning: No categories extracted - website structure may have changed")
		return []types.Category{}, nil
	}

	// Validate extracted categories
	validCategories := make([]types.Category, 0, len(categories))
	for _, category := range categories {
		if err := e.validateCategory(category); err != nil {
			util.DebugLogf("Warning: Skipping invalid category %s: %v", category.Name, err)
			continue
		}
		validCategories = append(validCategories, category)
	}

	// Sort categories alphabetically by name
	if len(validCategories) > 0 {
		sort.Slice(validCategories, func(i, j int) bool {
			return strings.ToLower(validCategories[i].Name) < strings.ToLower(validCategories[j].Name)
		})
	}

	util.DebugLogf("Successfully extracted %d valid categories", len(validCategories))
	return validCategories, nil
}

// parseCategories converts JavaScript result to Category slice
func (e *categoryExtractor) parseCategories(result interface{}) ([]types.Category, error) {
	util.DebugLogf("JavaScript result type: %T", result)
	util.DebugLogf("JavaScript result value: %+v", result)

	if result == nil {
		return nil, fmt.Errorf("JavaScript result is nil")
	}

	// Handle new return format with diagnostic info
	resultMap, ok := result.(map[string]interface{})
	if ok {
		// New format: {categories: [...], diagnostic: {...}, error: "..."}
		if errorMsg, hasError := resultMap["error"]; hasError && errorMsg != nil {
			if errorStr, isStr := errorMsg.(string); isStr && errorStr != "" {
				util.DebugLogf("JavaScript extraction error: %s", errorStr)
			}
		}

		if diagnostic, hasDiagnostic := resultMap["diagnostic"]; hasDiagnostic && diagnostic != nil {
			util.DebugLogf("Diagnostic info: %+v", diagnostic)
		}

		categoriesInterface, hasCats := resultMap["categories"]
		if !hasCats {
			return nil, fmt.Errorf("no categories field in result")
		}

		categoriesSlice, ok := categoriesInterface.([]interface{})
		if !ok {
			return nil, fmt.Errorf("categories field is not an array: %T", categoriesInterface)
		}

		return e.parseSliceToCategories(categoriesSlice)
	}

	// Fallback: try to parse as array directly (old format)
	categoriesSlice, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T, value: %+v", result, result)
	}

	return e.parseSliceToCategories(categoriesSlice)
}

// parseSliceToCategories converts a slice of category objects to Category slice
func (e *categoryExtractor) parseSliceToCategories(categoriesSlice []interface{}) ([]types.Category, error) {
	var categories []types.Category
	for _, catItem := range categoriesSlice {
		catMap, ok := catItem.(map[string]interface{})
		if !ok {
			continue
		}

		category := types.Category{
			Name:        util.GetString(catMap, "name"),
			Description: util.GetString(catMap, "description"),
			AgentCount:  util.GetInt(catMap, "agentCount"),
			URL:         util.GetString(catMap, "url"),
		}

		if category.Name != "" {
			// Extract slug from URL path instead of generating it
			category.Slug = util.ExtractSlugFromURL(category.URL)
			if category.Slug == "" {
				// Fallback to generated slug if URL extraction fails
				category.Slug = util.GenerateSlug(category.Name)
			}
			category.ID = category.Slug
			categories = append(categories, category)
		}
	}

	return categories, nil
}

// validateCategory checks if a category has valid data
func (e *categoryExtractor) validateCategory(category types.Category) error {
	if category.Name == "" {
		return errors.New("name is required")
	}
	if len(category.Name) > 100 {
		return errors.New("name is too long")
	}
	if category.Slug == "" {
		return errors.New("slug is required")
	}
	if len(category.Description) > 500 {
		return errors.New("description is too long")
	}
	if category.AgentCount < 0 {
		return errors.New("agent count cannot be negative")
	}
	return nil
}
