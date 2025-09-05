package validator

import (
	"strings"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
)

// TestNewValidator tests validator creation
func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Error("NewValidator should not return nil")
	}
}

// TestValidate_ValidAgent tests validation of a completely valid agent
func TestValidate_ValidAgent(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "test-agent",
		Description: "A comprehensive test agent for validation",
		Tools:       []string{"Read", "Write", "Edit"},
		Prompt:      "This is a valid prompt content.",
	}

	err := validator.Validate(spec)
	if err != nil {
		t.Errorf("Expected valid agent to pass validation, got error: %v", err)
	}
}

// TestValidate_ValidAgentWithInheritedTools tests validation of agent with inherited tools
func TestValidate_ValidAgentWithInheritedTools(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:           "simple-agent",
		Description:    "A simple agent with inherited tools",
		Tools:          []string{}, // Empty tools array
		ToolsInherited: true,
		Prompt:         "This is a valid prompt content.",
	}

	err := validator.Validate(spec)
	if err != nil {
		t.Errorf("Expected agent with inherited tools to pass validation, got error: %v", err)
	}
}

// TestValidate_MissingName tests validation failure for missing name
func TestValidate_MissingName(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "", // Missing name
		Description: "Agent without name",
		Prompt:      "Prompt content here.",
	}

	err := validator.Validate(spec)
	if err == nil {
		t.Error("Expected validation to fail for missing name")
	}

	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("Expected error message about required name, got: %v", err)
	}
}

// TestValidate_InvalidNameFormat tests validation failure for invalid name formats
func TestValidate_InvalidNameFormat(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name        string
		expectError bool
		description string
	}{
		{"valid-agent", false, "valid lowercase with hyphens"},
		{"valid-agent-123", false, "valid with numbers"},
		{"a", false, "valid single character"},
		{"Valid-Agent", true, "invalid starting with uppercase"},
		{"Invalid_Agent", true, "invalid underscore"},
		{"InvalidAgent", true, "invalid camelCase"},
		{"INVALID-AGENT", true, "invalid uppercase"},
		{"invalid agent", true, "invalid space"},
		{"invalid-agent!", true, "invalid special character"},
		{"123invalid", true, "invalid starting with number"},
		{"-invalid", true, "invalid starting with hyphen"},
		{"", true, "invalid empty string"},
	}

	for _, tc := range testCases {
		spec := &parser.AgentSpec{
			Name:        tc.name,
			Description: "Test description",
			Prompt:      "Test prompt.",
		}

		err := validator.Validate(spec)

		if tc.expectError {
			if err == nil {
				t.Errorf("Expected validation to fail for name '%s' (%s)", tc.name, tc.description)
			} else if tc.name == "" {
				// Empty name should give "name is required" error
				if !strings.Contains(err.Error(), "name is required") {
					t.Errorf("Expected error about required name for empty string, got: %v", err)
				}
			} else if !strings.Contains(err.Error(), "name must be lowercase") {
				t.Errorf("Expected error about name format for '%s', got: %v", tc.name, err)
			}
		} else {
			if err != nil {
				t.Errorf("Expected validation to pass for name '%s' (%s), got error: %v", tc.name, tc.description, err)
			}
		}
	}
}

// TestValidate_MissingDescription tests validation failure for missing description
func TestValidate_MissingDescription(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "test-agent",
		Description: "", // Missing description
		Prompt:      "Prompt content here.",
	}

	err := validator.Validate(spec)
	if err == nil {
		t.Error("Expected validation to fail for missing description")
	}

	if !strings.Contains(err.Error(), "description is required") {
		t.Errorf("Expected error message about required description, got: %v", err)
	}
}

// TestValidate_MissingPrompt tests validation failure for missing prompt
func TestValidate_MissingPrompt(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "test-agent",
		Description: "Test description",
		Prompt:      "", // Missing prompt
	}

	err := validator.Validate(spec)
	if err == nil {
		t.Error("Expected validation to fail for missing prompt")
	}

	if !strings.Contains(err.Error(), "agent prompt is required") {
		t.Errorf("Expected error message about required prompt, got: %v", err)
	}
}

// TestValidate_InvalidTools tests validation failure for invalid tool names
func TestValidate_InvalidTools(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "test-agent",
		Description: "Test description",
		Tools:       []string{"Read", "InvalidTool", "Write"}, // Contains invalid tool
		Prompt:      "Prompt content.",
	}

	err := validator.Validate(spec)
	if err == nil {
		t.Error("Expected validation to fail for invalid tool")
	}

	if !strings.Contains(err.Error(), "invalid tool: InvalidTool") {
		t.Errorf("Expected error message about invalid tool, got: %v", err)
	}
}

// TestValidate_ValidTools tests validation success for all valid tools
func TestValidate_ValidTools(t *testing.T) {
	validator := NewValidator()

	validTools := []string{
		"Read", "Write", "Edit", "MultiEdit", "Task",
		"Bash", "Grep", "Glob", "WebFetch", "WebSearch",
	}

	spec := &parser.AgentSpec{
		Name:        "test-agent",
		Description: "Test description",
		Tools:       validTools,
		Prompt:      "Prompt content.",
	}

	err := validator.Validate(spec)
	if err != nil {
		t.Errorf("Expected validation to pass for valid tools, got error: %v", err)
	}
}

// TestValidateWithReport_ValidAgent tests detailed validation report for valid agent
func TestValidateWithReport_ValidAgent(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "complete-agent",
		Description: "A complete agent with all fields",
		Tools:       []string{"Read", "Write"},
		Prompt:      "Complete prompt content.",
	}

	report := validator.ValidateWithReport(spec)

	if !report.Valid {
		t.Error("Expected report to indicate valid agent")
	}

	if len(report.Errors) > 0 {
		t.Errorf("Expected no errors for valid agent, got: %v", report.Errors)
	}

	if report.Coverage != 100.0 {
		t.Errorf("Expected 100%% coverage for complete agent, got %.1f%%", report.Coverage)
	}
}

// TestValidateWithReport_PartialAgent tests detailed validation report for partial agent
func TestValidateWithReport_PartialAgent(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "partial-agent",
		Description: "A partial agent missing tools",
		Tools:       []string{}, // No tools specified
		Prompt:      "Partial prompt content.",
	}

	report := validator.ValidateWithReport(spec)

	if !report.Valid {
		t.Error("Expected report to indicate valid agent (tools optional)")
	}

	if len(report.Errors) > 0 {
		t.Errorf("Expected no errors for valid partial agent, got: %v", report.Errors)
	}

	// Should have 75% coverage (3 out of 4 fields: name, description, prompt)
	expectedCoverage := 75.0
	if report.Coverage != expectedCoverage {
		t.Errorf("Expected %.1f%% coverage for partial agent, got %.1f%%", expectedCoverage, report.Coverage)
	}
}

// TestValidateWithReport_InvalidAgent tests detailed validation report for invalid agent
func TestValidateWithReport_InvalidAgent(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "Invalid_Name",          // Invalid format
		Description: "",                      // Missing description
		Tools:       []string{"InvalidTool"}, // Invalid tool
		Prompt:      "",                      // Missing prompt
	}

	report := validator.ValidateWithReport(spec)

	if report.Valid {
		t.Error("Expected report to indicate invalid agent")
	}

	expectedErrors := 4 // Invalid name format, missing description, invalid tool, missing prompt
	if len(report.Errors) != expectedErrors {
		t.Errorf("Expected %d errors, got %d: %v", expectedErrors, len(report.Errors), report.Errors)
	}

	// Should have 50% coverage (name and tools fields present, even if invalid)
	expectedCoverage := 50.0
	if report.Coverage != expectedCoverage {
		t.Errorf("Expected %.1f%% coverage for invalid agent, got %.1f%%", expectedCoverage, report.Coverage)
	}
}

// TestValidateWithReport_ShortDescription tests warning for very short description
func TestValidateWithReport_ShortDescription(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "test-agent",
		Description: "Short", // Very short description (should trigger warning)
		Prompt:      "Test prompt.",
	}

	report := validator.ValidateWithReport(spec)

	if !report.Valid {
		t.Error("Expected agent with short description to be valid")
	}

	if len(report.Warnings) == 0 {
		t.Error("Expected warning for very short description")
	}

	found := false
	for _, warning := range report.Warnings {
		if strings.Contains(warning, "Description very short") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected warning about short description, got warnings: %v", report.Warnings)
	}
}

// TestValidateWithReport_CoverageCalculation tests accuracy of coverage calculation
func TestValidateWithReport_CoverageCalculation(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		spec             *parser.AgentSpec
		expectedCoverage float64
		description      string
	}{
		{
			&parser.AgentSpec{
				Name:        "full-agent",
				Description: "Full description",
				Tools:       []string{"Read"},
				Prompt:      "Full prompt",
			},
			100.0,
			"all four fields present",
		},
		{
			&parser.AgentSpec{
				Name:        "three-fields",
				Description: "Description present",
				Prompt:      "Prompt present",
			},
			75.0,
			"three out of four fields",
		},
		{
			&parser.AgentSpec{
				Name:        "two-fields",
				Description: "Description present",
			},
			50.0,
			"two out of four fields",
		},
		{
			&parser.AgentSpec{
				Name: "one-field",
			},
			25.0,
			"one out of four fields",
		},
		{
			&parser.AgentSpec{},
			0.0,
			"no fields present",
		},
	}

	for _, tc := range testCases {
		report := validator.ValidateWithReport(tc.spec)

		if report.Coverage != tc.expectedCoverage {
			t.Errorf("Coverage calculation error for %s: expected %.1f%%, got %.1f%%",
				tc.description, tc.expectedCoverage, report.Coverage)
		}
	}
}

// TestValidateWithReport_MultipleErrors tests handling multiple validation errors
func TestValidateWithReport_MultipleErrors(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "",                                                     // Missing name (error 1)
		Description: "",                                                     // Missing description (error 2)
		Tools:       []string{"Read", "BadTool", "Write", "AnotherBadTool"}, // Multiple invalid tools (errors 3, 4)
		Prompt:      "",                                                     // Missing prompt (error 5)
	}

	report := validator.ValidateWithReport(spec)

	if report.Valid {
		t.Error("Expected report to indicate invalid agent")
	}

	expectedMinErrors := 5
	if len(report.Errors) < expectedMinErrors {
		t.Errorf("Expected at least %d errors, got %d: %v", expectedMinErrors, len(report.Errors), report.Errors)
	}

	// Check specific error messages are present
	errorMessages := strings.Join(report.Errors, " | ")

	requiredErrors := []string{
		"Missing name",
		"Missing description",
		"Invalid tool: BadTool",
		"Invalid tool: AnotherBadTool",
		"Missing prompt",
	}

	for _, requiredError := range requiredErrors {
		if !strings.Contains(errorMessages, requiredError) {
			t.Errorf("Expected error message containing '%s', got errors: %v", requiredError, report.Errors)
		}
	}
}

// TestValidateWithReport_EmptySpec tests validation of completely empty spec
func TestValidateWithReport_EmptySpec(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{} // Completely empty

	report := validator.ValidateWithReport(spec)

	if report.Valid {
		t.Error("Expected empty spec to be invalid")
	}

	if report.Coverage != 0.0 {
		t.Errorf("Expected 0%% coverage for empty spec, got %.1f%%", report.Coverage)
	}

	if len(report.Errors) < 3 {
		t.Errorf("Expected at least 3 errors for empty spec, got %d", len(report.Errors))
	}
}

// TestValidateWithReport_AllFieldsPopulated tests that all report fields are properly set
func TestValidateWithReport_AllFieldsPopulated(t *testing.T) {
	validator := NewValidator()

	spec := &parser.AgentSpec{
		Name:        "test-agent",
		Description: "Hi", // Short description to trigger warning
		Tools:       []string{"Read"},
		Prompt:      "Test prompt",
	}

	report := validator.ValidateWithReport(spec)

	// Check all report fields are initialized
	if report == nil {
		t.Fatal("ValidateWithReport returned nil")
	}

	if report.Errors == nil {
		t.Error("Report.Errors should be initialized (empty slice, not nil)")
	}

	if report.Warnings == nil {
		t.Error("Report.Warnings should be initialized (empty slice, not nil)")
	}

	if report.Coverage < 0 || report.Coverage > 100 {
		t.Errorf("Report.Coverage should be between 0 and 100, got %.1f", report.Coverage)
	}
}
