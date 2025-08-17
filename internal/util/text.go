package util

import (
	"regexp"
	"strconv"
	"strings"
	"time"
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

// ParseCount extracts the first number from text
func ParseCount(text string) int {
	re := regexp.MustCompile(`\d+`)
	if match := re.FindString(text); match != "" {
		if count, err := strconv.Atoi(match); err == nil {
			return count
		}
	}
	return 0
}

// ParseRating extracts a floating point rating from text
func ParseRating(text string) float32 {
	re := regexp.MustCompile(`[\d.]+`)
	if match := re.FindString(text); match != "" {
		if rating, err := strconv.ParseFloat(match, 32); err == nil {
			return float32(rating)
		}
	}
	return 0.0
}

// ParseDownloads handles formats like "1.2k", "500", "2.5M"
func ParseDownloads(text string) int {
	re := regexp.MustCompile(`([\d.]+)([kKmM]?)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		if base, err := strconv.ParseFloat(matches[1], 64); err == nil {
			multiplier := 1.0
			switch strings.ToLower(matches[2]) {
			case "k":
				multiplier = 1000
			case "m":
				multiplier = 1000000
			}
			return int(base * multiplier)
		}
	}
	return 0
}

// ParseDate attempts to parse date from text using common formats
func ParseDate(text string) time.Time {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"Jan 2, 2006",
		"January 2, 2006",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"02/01/2006",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, text); err == nil {
			return t
		}
	}

	return time.Time{}
}
