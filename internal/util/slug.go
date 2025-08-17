package util

import (
	"regexp"
	"strings"
)

// GenerateSlug converts text to a URL-friendly slug
func GenerateSlug(text string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(strings.ToLower(text), "-")
	return strings.Trim(slug, "-")
}
