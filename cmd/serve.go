package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/boozedog/webmd/internal/browser"
	"github.com/boozedog/webmd/internal/convert"
	"github.com/boozedog/webmd/internal/fetch"
	"github.com/boozedog/webmd/internal/preview"
	"github.com/go-rod/rod"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var (
		flagPort int
		flagHost string
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start an HTTP server that converts URLs to markdown",
		Long:  "Launch a persistent HTTP server that keeps a headless Chrome instance running.\nSend GET /?url=https://example.com to convert pages to markdown.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd, flagHost, flagPort)
		},
	}

	cmd.Flags().IntVar(&flagPort, "port", 8080, "Port to listen on")
	cmd.Flags().StringVar(&flagHost, "host", "0.0.0.0", "Host to bind to")

	return cmd
}

func runServe(cmd *cobra.Command, host string, port int) error {
	controlURL, cleanup, err := browser.Launch(browser.Options{
		BrowserPath: flagBrowserPath,
		NoDownload:  flagNoDownload,
	})
	if err != nil {
		return err
	}
	defer cleanup()

	b := rod.New().ControlURL(controlURL)
	if err := b.Connect(); err != nil {
		return fmt.Errorf("connecting to browser: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleConvert(b))

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	srv := &http.Server{Addr: addr, Handler: mux}

	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	fmt.Fprintf(cmd.OutOrStderr(), "webmd server listening on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func handleConvert(b *rod.Browser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.URL.Query().Get("url")
		if targetURL == "" {
			http.Error(w, "missing required 'url' query parameter", http.StatusBadRequest)
			return
		}

		article := r.URL.Query().Has("article") && r.URL.Query().Get("article") != "false" && r.URL.Query().Get("article") != "0"

		timeout := 15 * time.Second
		if t := r.URL.Query().Get("timeout"); t != "" {
			if d, err := time.ParseDuration(t); err == nil {
				timeout = d
			}
		}

		var wait time.Duration
		if ws := r.URL.Query().Get("wait"); ws != "" {
			if d, err := time.ParseDuration(ws); err == nil {
				wait = d
			}
		}

		userAgent := r.URL.Query().Get("user-agent")
		images := r.URL.Query().Has("images") && r.URL.Query().Get("images") != "false" && r.URL.Query().Get("images") != "0"
		mobile := r.URL.Query().Has("mobile") && r.URL.Query().Get("mobile") != "false" && r.URL.Query().Get("mobile") != "0"

		// Try markdown content negotiation first.
		if md := fetch.Markdown(targetURL, timeout); md != "" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte(md))
			return
		}

		result, err := fetch.PageOnBrowser(b, fetch.Options{
			URL:       targetURL,
			Timeout:   timeout,
			Wait:      wait,
			UserAgent: userAgent,
			Mobile:    mobile,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		html := result.HTML
		if !images {
			html = convert.StripImages(html)
		}

		var md string
		if html == "" {
			md = ""
		} else if article {
			md, err = convert.Readability(html)
		} else {
			md, err = convert.Full(html)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		md = convert.StripJunkLinks(md)

		if result.TimedOut {
			md = fmt.Sprintf("[webmd: page timed out after %s; content may be incomplete]\n\n%s", timeout, md)
		}

		wantPreview := r.URL.Query().Has("preview") && r.URL.Query().Get("preview") != "false" && r.URL.Query().Get("preview") != "0"
		if wantPreview {
			rendered, err := preview.Render(md)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(rendered))
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(md))
	}
}
