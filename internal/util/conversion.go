package util

import (
	"strconv"
)

// GetString safely extracts a string value from a map
func GetString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return CleanText(str)
		}
	}
	return ""
}

// GetInt safely extracts an integer value from a map
func GetInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// GetFloat32 safely extracts a float32 value from a map
func GetFloat32(m map[string]interface{}, key string) float32 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float32:
			return v
		case float64:
			return float32(v)
		case int:
			return float32(v)
		case string:
			if f, err := strconv.ParseFloat(v, 32); err == nil {
				return float32(f)
			}
		}
	}
	return 0.0
}
