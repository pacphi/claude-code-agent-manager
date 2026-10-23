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
	err := chromedp.Do(c.browserCtx,
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

// waitForContent waits for dynamic content to load on marketplace pages - HYBRID OPTIMIZED VERSION
func (c *ChromeController) waitForContent(url string) chromedp.Action[chromedp.Void] {
	return chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
		// For marketplace category pages, wait for agents to load
		if strings.Contains(url, "/categories/") && !strings.HasSuffix(url, "/categories") {
			// HYBRID: Reasonable wait time
			maxRetries := 5 // 5 seconds max for Next.js hydration
			for i := 0; i < maxRetries; i++ {
				bodyText, err := chromedp.Text(chromedp.CSS("body"))(ctx, t)
				if err == nil {
					// Check if content has loaded (no "Searching..." and has "Agents Found")
					if strings.Contains(bodyText, "Agents Found") && !strings.Contains(bodyText, "Searching...") {
						// HYBRID: Moderate wait for React hydration - not too short, not too long
						time.Sleep(3000 * time.Millisecond)

						// HYBRID: Strategic scrolling to trigger lazy loading without overdoing it
						for initialScroll := 0; initialScroll < 3; initialScroll++ {
							_, _ = chromedp.Evaluate[chromedp.Void](`window.scrollTo(0, document.body.scrollHeight)`)(ctx, t)
							time.Sleep(800 * time.Millisecond)

							// Scroll to middle and back to trigger intersection observers
							_, _ = chromedp.Evaluate[chromedp.Void](`window.scrollTo(0, document.body.scrollHeight / 2)`)(ctx, t)
							time.Sleep(500 * time.Millisecond)
							_, _ = chromedp.Evaluate[chromedp.Void](`window.scrollTo(0, document.body.scrollHeight)`)(ctx, t)
							time.Sleep(500 * time.Millisecond)
						}

						c.handleLoadMorePagination(ctx, t)

						return nil // Content loaded
					}
				}
				time.Sleep(250 * time.Millisecond) // Wait 250ms before retry
			}
		} else {
			// For other pages, use standard wait
			time.Sleep(500 * time.Millisecond)
		}
		return nil
	})
}

func (c *ChromeController) handleLoadMorePagination(ctx context.Context, t *chromedp.Target) {
	maxClicks := 10 // Reasonable limit - most categories have 1-2 pages

	// Get expected count to know when we're done
	expectedCount, _ := chromedp.Evaluate[int](`
		(function() {
			const pageText = document.body.textContent;
			const countMatch = pageText.match(/(\d+)\s+Agents?\s+Found/i);
			console.log('Expected count search in:', pageText.substring(0, 200));
			return countMatch ? parseInt(countMatch[1]) : 0;
		})()
	`)(ctx, t)

	for i := 0; i < maxClicks; i++ {
		// Quick check for Load More button
		hasLoadMore, err := chromedp.Evaluate[bool](`
			(function() {
				const buttons = Array.from(document.querySelectorAll('button'));
				const loadMore = buttons.find(btn => {
					const text = btn.textContent?.trim().toLowerCase();
					return text === 'load more' || text.includes('load more');
				});
				return loadMore && !loadMore.disabled && loadMore.style.display !== 'none';
			})()
		`)(ctx, t)

		if err != nil || !hasLoadMore {
			// No Load More button, we're done
			break
		}

		// Click Load More button quickly
		_, err = chromedp.Click(chromedp.Search(`//button[contains(translate(., 'LOAD MORE', 'load more'), 'load more')]`))(ctx, t)
		if err != nil {
			// Try JavaScript click as fallback
			clicked, _ := chromedp.Evaluate[bool](`
				(function() {
					const buttons = Array.from(document.querySelectorAll('button'));
					const loadMore = buttons.find(btn => {
						const text = btn.textContent?.trim().toLowerCase();
						return text === 'load more' || text.includes('load more');
					});
					if (loadMore && !loadMore.disabled) {
						loadMore.click();
						return true;
					}
					return false;
				})()
			`)(ctx, t)

			if !clicked {
				break // Can't click, we're done
			}
		}

		// OPTIMIZED: Wait for new content - but not too long
		time.Sleep(2000 * time.Millisecond)

		// OPTIMIZED: Single scroll to trigger any new Load More buttons
		_, _ = chromedp.Evaluate[chromedp.Void](`window.scrollTo(0, document.body.scrollHeight)`)(ctx, t)
		time.Sleep(50 * time.Millisecond)

		// Check if we've reached expected count, but don't trust the expected count completely
		if expectedCount > 0 {
			newCount, _ := chromedp.Evaluate[int](`Array.from(document.querySelectorAll('button')).filter(b => b.textContent?.trim().toLowerCase() === 'view').length`)(ctx, t)
			// Only exit if we have significantly more agents than expected
			if newCount >= expectedCount+5 {
				break
			}
		}
	}

	// OPTIMIZED: Final single scroll to ensure everything is visible
	_, _ = chromedp.Evaluate[chromedp.Void](`window.scrollTo(0, document.body.scrollHeight)`)(ctx, t)
	time.Sleep(250 * time.Millisecond)
}

// ExecuteScript executes JavaScript and returns the result
func (c *ChromeController) ExecuteScript(ctx context.Context, script string) (interface{}, error) {
	result, err := chromedp.Run(c.browserCtx, chromedp.Evaluate[any](script))

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScriptExecution, err)
	}

	return result, nil
}

// WaitForElement waits for an element to be present
func (c *ChromeController) WaitForElement(ctx context.Context, selector string) error {
	err := chromedp.Do(c.browserCtx,
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
