package util

import (
	"net/url"
	"strings"
)

// ResolveURL resolves a relative URL against a base URL
func ResolveURL(base, relative string) string {
	if relative == "" {
		return ""
	}

	// If relative is already absolute, return it
	if strings.HasPrefix(relative, "http://") || strings.HasPrefix(relative, "https://") {
		return relative
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return relative
	}

	relURL, err := url.Parse(relative)
	if err != nil {
		return relative
	}

	return baseURL.ResolveReference(relURL).String()
}

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
