package util

import (
	"regexp"
	"strings"
)

// CleanText removes extra whitespace and cleans up text
func CleanText(text string) string {
	// Remove extra whitespace and clean up text
	re := regexp.MustCompile(`\s+`)
	cleaned := strings.TrimSpace(re.ReplaceAllString(text, " "))
	// Remove common unwanted characters
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")
	return cleaned
}
