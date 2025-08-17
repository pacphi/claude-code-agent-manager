package util

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// IsDebugEnabled returns true if DEBUG environment variable is set to true, 1, or yes (case-insensitive)
func IsDebugEnabled() bool {
	debug := strings.ToLower(os.Getenv("DEBUG"))
	return debug == "true" || debug == "1" || debug == "yes"
}

// DebugPrintf prints a debug message only if DEBUG environment variable is enabled
func DebugPrintf(format string, args ...interface{}) {
	if IsDebugEnabled() {
		fmt.Printf(format, args...)
	}
}

// DebugLogf logs a debug message only if DEBUG environment variable is enabled
func DebugLogf(format string, args ...interface{}) {
	if IsDebugEnabled() {
		log.Printf(format, args...)
	}
}
