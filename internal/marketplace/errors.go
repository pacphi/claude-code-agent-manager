package marketplace

import (
	"errors"
)

// Common marketplace errors
var (
	ErrBrowserNotAvailable   = errors.New("browser not available")
	ErrExtractionFailed      = errors.New("data extraction failed")
	ErrNavigationFailed      = errors.New("navigation failed")
	ErrScriptExecutionFailed = errors.New("script execution failed")
	ErrInvalidConfiguration  = errors.New("invalid configuration")
	ErrCacheOperationFailed  = errors.New("cache operation failed")
	ErrNoDataFound           = errors.New("no data found")
	ErrInvalidURL            = errors.New("invalid URL")
	ErrTimeout               = errors.New("operation timeout")
	ErrRateLimited           = errors.New("rate limited")
)

// MarketplaceError wraps errors with additional context
type MarketplaceError struct {
	Operation string
	URL       string
	Cause     error
}
