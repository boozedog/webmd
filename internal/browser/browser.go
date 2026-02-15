package browser

import (
	"fmt"
	"os"

	"github.com/go-rod/rod/lib/launcher"
)

type Options struct {
	BrowserPath string
	NoDownload  bool
}

// Launch starts a headless Chrome instance and returns the DevTools control URL
// along with a cleanup function that should be deferred.
func Launch(opts Options) (controlURL string, cleanup func(), err error) {
	path := opts.BrowserPath
	if path == "" {
		path = os.Getenv("WEBMD_BROWSER_PATH")
	}

	var l *launcher.Launcher

	switch {
	case path != "":
		l = launcher.New().Bin(path)
	case opts.NoDownload:
		found, ok := launcher.LookPath()
		if !ok {
			return "", nil, fmt.Errorf("no system Chrome/Chromium found (auto-download disabled)")
		}
		l = launcher.New().Bin(found)
	default:
		found, ok := launcher.LookPath()
		if ok {
			l = launcher.New().Bin(found)
		} else {
			l = launcher.New() // rod will auto-download
		}
	}

	l = l.Headless(true)

	// In containers, Chrome needs --no-sandbox
	if inContainer() {
		l = l.NoSandbox(true)
	}

	u, err := l.Launch()
	if err != nil {
		return "", nil, fmt.Errorf("launching browser: %w", err)
	}

	cleanup = func() {
		l.Kill()
	}

	return u, cleanup, nil
}

func inContainer() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}
