package validator

import (
	"fmt"
	"regexp"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
)

// Validator validates agent specifications
type Validator struct {
	namePattern *regexp.Regexp
	validTools  map[string]bool
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		namePattern: regexp.MustCompile("^[a-z][a-z0-9-]*$"),
		validTools: map[string]bool{
			"Read":      true,
			"Write":     true,
			"Edit":      true,
			"MultiEdit": true,
			"Task":      true,
			"Bash":      true,
			"Grep":      true,
			"Glob":      true,
			"WebFetch":  true,
			"WebSearch": true,
			// Add other Claude Code tools as needed
		},
	}
}

// Validate checks if an agent spec is valid
func (v *Validator) Validate(spec *parser.AgentSpec) error {
	// Check required field: name
	if spec.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !v.namePattern.MatchString(spec.Name) {
		return fmt.Errorf("name must be lowercase with hyphens: %s", spec.Name)
	}

	// Check required field: description
	if spec.Description == "" {
		return fmt.Errorf("description is required")
	}

	// Check optional field: tools
	for _, tool := range spec.Tools {
		if !v.validTools[tool] {
			return fmt.Errorf("invalid tool: %s", tool)
		}
	}

	// Check prompt exists
	if spec.Prompt == "" {
		return fmt.Errorf("agent prompt is required")
	}

	return nil
}

// ValidationReport provides validation details
type ValidationReport struct {
	Valid    bool
	Errors   []string
	Warnings []string
	Coverage float64
}

// ValidateWithReport provides detailed validation
func (v *Validator) ValidateWithReport(spec *parser.AgentSpec) *ValidationReport {
	report := &ValidationReport{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Calculate coverage (4 fields total)
	fieldsPresent := 0
	totalFields := 4

	// Name (required)
	if spec.Name != "" {
		fieldsPresent++
		if !v.namePattern.MatchString(spec.Name) {
			report.Errors = append(report.Errors, "Invalid name format")
			report.Valid = false
		}
	} else {
		report.Errors = append(report.Errors, "Missing name")
		report.Valid = false
	}

	// Description (required)
	if spec.Description != "" {
		fieldsPresent++
		if len(spec.Description) < 10 {
			report.Warnings = append(report.Warnings, "Description very short")
		}
	} else {
		report.Errors = append(report.Errors, "Missing description")
		report.Valid = false
	}

	// Tools (optional)
	if len(spec.Tools) > 0 {
		fieldsPresent++
		for _, tool := range spec.Tools {
			if !v.validTools[tool] {
				report.Errors = append(report.Errors, fmt.Sprintf("Invalid tool: %s", tool))
				report.Valid = false
			}
		}
	}

	// Prompt (required)
	if spec.Prompt != "" {
		fieldsPresent++
	} else {
		report.Errors = append(report.Errors, "Missing prompt")
		report.Valid = false
	}

	report.Coverage = float64(fieldsPresent) / float64(totalFields) * 100

	return report
}
