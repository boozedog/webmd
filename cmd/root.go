package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/boozedog/webmd/internal/browser"
	"github.com/boozedog/webmd/internal/convert"
	"github.com/boozedog/webmd/internal/fetch"
	"github.com/spf13/cobra"
)

var (
	version string

	flagFull        bool
	flagBrowserPath string
	flagNoDownload  bool
	flagTimeout     time.Duration
	flagWait        time.Duration
	flagUserAgent   string
	flagOutput      string
)

func SetVersion(v string) {
	version = v
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webmd [flags] <url>",
		Short:   "Convert web pages to agent-friendly markdown",
		Long:    "Fetch a URL using headless Chrome and convert it to clean markdown.\nDefault mode extracts main content via readability; use --full for entire page.",
		Version: version,
		Args:    cobra.ExactArgs(1),
		RunE:    runRoot,
	}

	cmd.Flags().BoolVar(&flagFull, "full", false, "Convert full page instead of readability extraction")
	cmd.Flags().StringVar(&flagBrowserPath, "browser-path", "", "Path to Chrome/Chromium binary (overrides auto-detect)")
	cmd.Flags().BoolVar(&flagNoDownload, "no-download", false, "Disable auto-download of Chromium; fail if no system Chrome found")
	cmd.Flags().DurationVar(&flagTimeout, "timeout", 5*time.Second, "Page load timeout")
	cmd.Flags().DurationVar(&flagWait, "wait", 0, "Extra wait after page load for JS-heavy sites")
	cmd.Flags().StringVar(&flagUserAgent, "user-agent", "", "Custom User-Agent string")
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write to file instead of stdout")

	return cmd
}

func runRoot(cmd *cobra.Command, args []string) error {
	targetURL := args[0]

	controlURL, cleanup, err := browser.Launch(browser.Options{
		BrowserPath: flagBrowserPath,
		NoDownload:  flagNoDownload,
	})
	if err != nil {
		return err
	}
	defer cleanup()

	result, err := fetch.Page(controlURL, fetch.Options{
		URL:       targetURL,
		Timeout:   flagTimeout,
		Wait:      flagWait,
		UserAgent: flagUserAgent,
	})
	if err != nil {
		return err
	}

	var md string
	if result.HTML == "" {
		md = ""
	} else if flagFull {
		md, err = convert.Full(result.HTML)
	} else {
		md, err = convert.Readability(result.HTML)
	}
	if err != nil {
		return err
	}

	if result.TimedOut {
		md = fmt.Sprintf("[webmd: page timed out after %s; content may be incomplete]\n\n%s", flagTimeout, md)
	}

	if flagOutput != "" {
		if err := os.WriteFile(flagOutput, []byte(md), 0o644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		return nil
	}

	fmt.Fprint(cmd.OutOrStdout(), md)
	return nil
}

func Execute() error {
	return newRootCmd().Execute()
}
