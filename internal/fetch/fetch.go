package fetch

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Options struct {
	URL       string
	Timeout   time.Duration
	Wait      time.Duration
	UserAgent string
}

// Result holds the fetched HTML and metadata about the fetch.
type Result struct {
	HTML     string
	TimedOut bool
}

// Page connects to a browser via controlURL, navigates to the target URL,
// and returns the fully rendered HTML. If the page times out waiting for the
// DOM to stabilize, it returns whatever HTML is available along with TimedOut=true.
func Page(controlURL string, opts Options) (*Result, error) {
	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connecting to browser: %w", err)
	}
	defer browser.MustClose()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, fmt.Errorf("creating page: %w", err)
	}
	defer page.MustClose()

	if opts.UserAgent != "" {
		err := proto.NetworkSetUserAgentOverride{UserAgent: opts.UserAgent}.Call(page)
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

	timedOut := false
	if err := timedPage.WaitDOMStable(300*time.Millisecond, 0.1); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			timedOut = true
		} else {
			return nil, fmt.Errorf("waiting for DOM stable: %w", err)
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
