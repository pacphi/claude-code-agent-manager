package util

import (
	"net/url"
	"strings"
)

// ExtractSlugFromURL extracts a slug from a URL path
func ExtractSlugFromURL(urlStr string) string {
	if urlStr == "" {
		return ""
	}

	// Parse the URL to extract the path
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Extract slug from path like "/categories/ai-ml" -> "ai-ml"
	path := strings.Trim(parsedURL.Path, "/")
	parts := strings.Split(path, "/")

	// Look for the pattern "categories/slug"
	for i, part := range parts {
		if part == "categories" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	// If not in expected format, return the last non-empty path segment
	if len(parts) > 0 && parts[len(parts)-1] != "" {
		return parts[len(parts)-1]
	}

	return ""
}
