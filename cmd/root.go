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

	flagArticle     bool
	flagMobile      bool
	flagImages      bool
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
		Long:    "Fetch a URL using headless Chrome and convert it to clean markdown.\nDefault mode converts the full page; use --article to extract main content via readability.",
		Version: version,
		Args:    cobra.ExactArgs(1),
		RunE:    runRoot,
	}

	cmd.PersistentFlags().StringVar(&flagBrowserPath, "browser-path", "", "Path to Chrome/Chromium binary (overrides auto-detect)")
	cmd.PersistentFlags().BoolVar(&flagNoDownload, "no-download", false, "Disable auto-download of Chromium; fail if no system Chrome found")

	cmd.Flags().BoolVar(&flagArticle, "article", false, "Extract main article content via readability")
	cmd.Flags().BoolVar(&flagMobile, "mobile", false, "Emulate a mobile device (iPhone viewport and user-agent)")
	cmd.Flags().BoolVar(&flagImages, "images", false, "Include images in markdown output")
	cmd.Flags().DurationVar(&flagTimeout, "timeout", 15*time.Second, "Page load timeout")
	cmd.Flags().DurationVar(&flagWait, "wait", 0, "Extra wait after page load for JS-heavy sites")
	cmd.Flags().StringVar(&flagUserAgent, "user-agent", "", "Custom User-Agent string")
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write to file instead of stdout")

	cmd.AddCommand(newServeCmd())

	return cmd
}

func runRoot(cmd *cobra.Command, args []string) error {
	targetURL := args[0]

	// Try markdown content negotiation first — skip the browser entirely if the server supports it.
	if md := fetch.Markdown(targetURL, flagTimeout); md != "" {
		return writeOutput(cmd, md)
	}

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
		Mobile:    flagMobile,
	})
	if err != nil {
		return err
	}

	html := result.HTML
	if !flagImages {
		html = convert.StripImages(html)
	}

	var md string
	if html == "" {
		md = ""
	} else if flagArticle {
		md, err = convert.Readability(html)
	} else {
		md, err = convert.Full(html)
	}
	if err != nil {
		return err
	}

	md = convert.StripJunkLinks(md)

	if result.TimedOut {
		md = fmt.Sprintf("[webmd: page timed out after %s; content may be incomplete]\n\n%s", flagTimeout, md)
	}

	return writeOutput(cmd, md)
}

func writeOutput(cmd *cobra.Command, md string) error {
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
