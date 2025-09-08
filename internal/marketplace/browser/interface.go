package browser

import (
	"context"
	"errors"
)

var (
	ErrChromeNotFound    = errors.New("chrome executable not found")
	ErrBrowserClosed     = errors.New("browser context is closed")
	ErrScriptExecution   = errors.New("script execution failed")
	ErrNavigationTimeout = errors.New("navigation timeout")
)

// Controller defines the interface for browser automation
type Controller interface {
	Navigate(ctx context.Context, url string) error
	ExecuteScript(ctx context.Context, script string) (interface{}, error)
	WaitForElement(ctx context.Context, selector string) error
	ScrollPage(ctx context.Context, offset int) error
	Close() error
}

// Options configures browser behavior
type Options struct {
	Headless     bool
	Timeout      int // seconds
	UserAgent    string
	WindowWidth  int
	WindowHeight int
}
