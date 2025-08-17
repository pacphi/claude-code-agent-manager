package marketplace

import (
	"errors"
	"fmt"
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

func (e *MarketplaceError) Error() string {
	if e.URL != "" {
		return fmt.Sprintf("marketplace %s failed for %s: %v", e.Operation, e.URL, e.Cause)
	}
	return fmt.Sprintf("marketplace %s failed: %v", e.Operation, e.Cause)
}

func (e *MarketplaceError) Unwrap() error {
	return e.Cause
}

// NewError creates a new marketplace error
func NewError(operation string, cause error) *MarketplaceError {
	return &MarketplaceError{
		Operation: operation,
		Cause:     cause,
	}
}

// NewErrorWithURL creates a new marketplace error with URL context
func NewErrorWithURL(operation, url string, cause error) *MarketplaceError {
	return &MarketplaceError{
		Operation: operation,
		URL:       url,
		Cause:     cause,
	}
}

// IsTemporary checks if an error is temporary and can be retried
func IsTemporary(err error) bool {
	var mErr *MarketplaceError
	if errors.As(err, &mErr) {
		return errors.Is(mErr.Cause, ErrTimeout) ||
			errors.Is(mErr.Cause, ErrRateLimited) ||
			errors.Is(mErr.Cause, ErrNavigationFailed)
	}
	return false
}

// IsPermanent checks if an error is permanent and should not be retried
func IsPermanent(err error) bool {
	var mErr *MarketplaceError
	if errors.As(err, &mErr) {
		return errors.Is(mErr.Cause, ErrInvalidURL) ||
			errors.Is(mErr.Cause, ErrInvalidConfiguration) ||
			errors.Is(mErr.Cause, ErrBrowserNotAvailable)
	}
	return false
}
