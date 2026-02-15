package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Options struct {
	URL       string
	Timeout   time.Duration
	Wait      time.Duration
	UserAgent string
	Mobile    bool
}

// iPhone 14 Pro Max dimensions and UA for mobile emulation.
const mobileUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"

// Result holds the fetched content and metadata about the fetch.
type Result struct {
	HTML     string
	Markdown string // Set when the server provided markdown directly.
	TimedOut bool
}

// Markdown attempts a lightweight HTTP GET with Accept: text/markdown.
// Returns the markdown body if the server responds with text/markdown,
// or empty string if not supported.
func Markdown(url string, timeout time.Duration) string {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "text/markdown")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/markdown") {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

// Page connects to a browser via controlURL, navigates to the target URL,
// and returns the fully rendered HTML. The browser connection is closed when done.
// If the page times out waiting for the DOM to stabilize, it returns whatever
// HTML is available along with TimedOut=true.
func Page(controlURL string, opts Options) (*Result, error) {
	b := rod.New().ControlURL(controlURL)
	if err := b.Connect(); err != nil {
		return nil, fmt.Errorf("connecting to browser: %w", err)
	}
	defer b.MustClose()

	return PageOnBrowser(b, opts)
}

// PageOnBrowser navigates to the target URL using an existing browser connection
// and returns the fully rendered HTML. The browser is NOT closed after the fetch,
// making this suitable for server mode with a persistent browser.
func PageOnBrowser(b *rod.Browser, opts Options) (*Result, error) {
	page, err := b.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, fmt.Errorf("creating page: %w", err)
	}
	defer page.MustClose()

	userAgent := opts.UserAgent
	if opts.Mobile {
		// Set mobile viewport: iPhone 14 Pro Max logical resolution.
		err := proto.EmulationSetDeviceMetricsOverride{
			Width:             430,
			Height:            932,
			DeviceScaleFactor: 3,
			Mobile:            true,
		}.Call(page)
		if err != nil {
			return nil, fmt.Errorf("setting mobile viewport: %w", err)
		}
		if userAgent == "" {
			userAgent = mobileUserAgent
		}
	}

	if userAgent != "" {
		err := proto.NetworkSetUserAgentOverride{UserAgent: userAgent}.Call(page)
		if err != nil {
			return nil, fmt.Errorf("setting user agent: %w", err)
		}
	}

	// Apply timeout to navigation and DOM wait, but not HTML extraction.
	timedPage := page
	if opts.Timeout > 0 {
		timedPage = page.Timeout(opts.Timeout)
	}

	if err := timedPage.Navigate(opts.URL); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return &Result{TimedOut: true}, nil
		}
		return nil, fmt.Errorf("navigating to %s: %w", opts.URL, err)
	}

	// Wait for the load event first (resources + scripts loaded), then for
	// DOM mutations to settle. This ensures JS-heavy sites that render
	// content asynchronously have time to populate the page.
	timedOut := false
	if err := timedPage.WaitLoad(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			timedOut = true
		} else {
			return nil, fmt.Errorf("waiting for page load: %w", err)
		}
	}

	if !timedOut {
		if err := timedPage.WaitDOMStable(300*time.Millisecond, 0.1); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				timedOut = true
			} else {
				return nil, fmt.Errorf("waiting for DOM stable: %w", err)
			}
		}
	}

	if !timedOut && opts.Wait > 0 {
		time.Sleep(opts.Wait)
	}

	// Use the original page (no timeout) to extract HTML.
	html, err := page.HTML()
	if err != nil {
		return nil, fmt.Errorf("extracting HTML: %w", err)
	}

	return &Result{HTML: html, TimedOut: timedOut}, nil
}
