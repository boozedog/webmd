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
	flagKeepNav     bool
	flagFrontmatter bool
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
	cmd.Flags().BoolVar(&flagKeepNav, "keep-nav", false, "Keep nav, header, footer, and aside elements")
	cmd.Flags().BoolVar(&flagFrontmatter, "frontmatter", false, "Prepend YAML frontmatter with source URL, fetch method, and timing")
	cmd.Flags().DurationVar(&flagTimeout, "timeout", 15*time.Second, "Page load timeout")
	cmd.Flags().DurationVar(&flagWait, "wait", 0, "Extra wait after page load for JS-heavy sites")
	cmd.Flags().StringVar(&flagUserAgent, "user-agent", "", "Custom User-Agent string")
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write to file instead of stdout")

	cmd.AddCommand(newServeCmd())

	return cmd
}

func runRoot(cmd *cobra.Command, args []string) error {
	targetURL := args[0]
	start := time.Now()
	var timing []convert.TimingStep

	// Try markdown content negotiation first — skip the browser entirely if the server supports it.
	fetchStart := time.Now()
	if md := fetch.Markdown(targetURL, flagTimeout); md != "" {
		timing = append(timing, convert.TimingStep{Name: "fetch", Duration: time.Since(fetchStart)})

		stepStart := time.Now()
		md, err := convert.FormatMarkdown(md)
		if err != nil {
			return err
		}
		timing = append(timing, convert.TimingStep{Name: "format", Duration: time.Since(stepStart)})

		if flagFrontmatter {
			timing = append(timing, convert.TimingStep{Name: "total", Duration: time.Since(start)})
			md = convert.Frontmatter(convert.Metadata{
				SourceURL:   targetURL,
				FetchMethod: "markdown",
				Timing:      timing,
			}) + md
		}
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
	timing = append(timing, convert.TimingStep{Name: "fetch", Duration: time.Since(fetchStart)})

	html := result.HTML

	stepStart := time.Now()
	html = convert.StripHidden(html)
	timing = append(timing, convert.TimingStep{Name: "strip_hidden", Duration: time.Since(stepStart)})

	if !flagKeepNav {
		stepStart = time.Now()
		html = convert.StripNav(html)
		timing = append(timing, convert.TimingStep{Name: "strip_nav", Duration: time.Since(stepStart)})
	}

	if !flagImages {
		stepStart = time.Now()
		html = convert.StripImages(html)
		timing = append(timing, convert.TimingStep{Name: "strip_images", Duration: time.Since(stepStart)})
	}

	var md string
	stepStart = time.Now()
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
	timing = append(timing, convert.TimingStep{Name: "convert", Duration: time.Since(stepStart)})

	stepStart = time.Now()
	md = convert.StripJunkLinks(md)
	timing = append(timing, convert.TimingStep{Name: "strip_junk_links", Duration: time.Since(stepStart)})

	stepStart = time.Now()
	md, err = convert.FormatMarkdown(md)
	if err != nil {
		return err
	}
	timing = append(timing, convert.TimingStep{Name: "format", Duration: time.Since(stepStart)})

	if result.TimedOut {
		md = fmt.Sprintf("[webmd: page timed out after %s; content may be incomplete]\n\n%s", flagTimeout, md)
	}

	if flagFrontmatter {
		timing = append(timing, convert.TimingStep{Name: "total", Duration: time.Since(start)})
		md = convert.Frontmatter(convert.Metadata{
			SourceURL:   targetURL,
			FetchMethod: "browser",
			TimedOut:    result.TimedOut,
			Timing:      timing,
		}) + md
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
