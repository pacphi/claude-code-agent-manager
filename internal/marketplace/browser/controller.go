package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// ChromeController implements Controller interface using chromedp
type ChromeController struct {
	allocCtx      context.Context
	cancel        context.CancelFunc
	browserCtx    context.Context
	browserCancel context.CancelFunc
	opts          Options
}

// NewController creates a new browser controller
func NewController(opts Options) (*ChromeController, error) {
	execPath := findChromeExecutable()
	if execPath == "" {
		return nil, ErrChromeNotFound
	}

	allocOpts := buildChromeOptions(execPath, opts)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)

	// Create a persistent browser context
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)

	return &ChromeController{
		allocCtx:      allocCtx,
		cancel:        cancel,
		browserCtx:    browserCtx,
		browserCancel: browserCancel,
		opts:          opts,
	}, nil
}

// Navigate navigates to the specified URL
func (c *ChromeController) Navigate(ctx context.Context, url string) error {
	err := chromedp.Run(c.browserCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		// Wait longer for dynamic content to load and retry if needed
		c.waitForContent(url),
	)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrNavigationTimeout, err)
	}

	return nil
}

// waitForContent waits for dynamic content to load on marketplace pages
func (c *ChromeController) waitForContent(url string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// For marketplace category pages, wait for agents to load
		if strings.Contains(url, "/categories/") && !strings.HasSuffix(url, "/categories") {
			// Wait for either "Agents Found" text or skeleton content to be replaced
			maxRetries := 25 // 25 seconds total for Next.js hydration
			for i := 0; i < maxRetries; i++ {
				var bodyText string
				err := chromedp.Text("body", &bodyText, chromedp.ByQuery).Do(ctx)
				if err == nil {
					// Check if content has loaded (no "Searching..." and has "Agents Found")
					if strings.Contains(bodyText, "Agents Found") && !strings.Contains(bodyText, "Searching...") {
						// Additional wait to ensure React hydration is complete
						time.Sleep(3000 * time.Millisecond)

						// Scroll down to load all agents (lazy loading)
						_ = chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil).Do(ctx)

						// Wait for lazy-loaded content
						time.Sleep(2000 * time.Millisecond)

						return nil // Content loaded
					}
				}
				time.Sleep(1000 * time.Millisecond) // Wait 1 second before retry
			}
		} else {
			// For other pages, use standard wait
			time.Sleep(3000 * time.Millisecond)
		}
		return nil
	})
}

// ExecuteScript executes JavaScript and returns the result
func (c *ChromeController) ExecuteScript(ctx context.Context, script string) (interface{}, error) {
	var result interface{}
	err := chromedp.Run(c.browserCtx,
		chromedp.Evaluate(script, &result),
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScriptExecution, err)
	}

	return result, nil
}

// WaitForElement waits for an element to be present
func (c *ChromeController) WaitForElement(ctx context.Context, selector string) error {
	err := chromedp.Run(c.browserCtx,
		chromedp.WaitVisible(selector),
	)

	if err != nil {
		return fmt.Errorf("element %s not found: %w", selector, err)
	}

	return nil
}

// ScrollPage scrolls the page by the specified offset
func (c *ChromeController) ScrollPage(ctx context.Context, offset int) error {
	script := fmt.Sprintf("window.scrollBy(0, %d);", offset)
	_, err := c.ExecuteScript(ctx, script)
	return err
}

// Close closes the browser context
func (c *ChromeController) Close() error {
	if c.browserCancel != nil {
		c.browserCancel()
	}
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

// buildChromeOptions creates chromedp allocator options
func buildChromeOptions(execPath string, opts Options) []chromedp.ExecAllocatorOption {
	allocOpts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(execPath),
		chromedp.UserAgent(opts.UserAgent),
		chromedp.WindowSize(opts.WindowWidth, opts.WindowHeight),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
	}

	if opts.Headless {
		allocOpts = append(allocOpts, chromedp.Headless)
	}

	return allocOpts
}
