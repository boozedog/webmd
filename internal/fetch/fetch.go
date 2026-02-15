package fetch

import (
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

// Page connects to a browser via controlURL, navigates to the target URL,
// and returns the fully rendered HTML.
func Page(controlURL string, opts Options) (string, error) {
	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return "", fmt.Errorf("connecting to browser: %w", err)
	}
	defer browser.MustClose()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("creating page: %w", err)
	}
	defer page.MustClose()

	if opts.Timeout > 0 {
		page = page.Timeout(opts.Timeout)
	}

	if opts.UserAgent != "" {
		err := proto.NetworkSetUserAgentOverride{UserAgent: opts.UserAgent}.Call(page)
		if err != nil {
			return "", fmt.Errorf("setting user agent: %w", err)
		}
	}

	if err := page.Navigate(opts.URL); err != nil {
		return "", fmt.Errorf("navigating to %s: %w", opts.URL, err)
	}

	if err := page.WaitLoad(); err != nil {
		return "", fmt.Errorf("waiting for page load: %w", err)
	}

	if opts.Wait > 0 {
		time.Sleep(opts.Wait)
	}

	html, err := page.HTML()
	if err != nil {
		return "", fmt.Errorf("extracting HTML: %w", err)
	}

	return html, nil
}
