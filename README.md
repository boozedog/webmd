# webmd

Convert web pages to agent-friendly markdown using headless Chrome.

Fetches URLs with a real browser (renders JavaScript), then extracts clean markdown. Two modes:

- **Readability** (default) — extracts main article content
- **Full page** — converts the entire page

## Install

```bash
go install github.com/boozedog/webmd@latest
```

Or with Docker:

```bash
docker pull ghcr.io/boozedog/webmd
```

## Usage

```bash
# Extract main content (readability mode)
webmd https://example.com

# Full page conversion
webmd --full https://example.com

# Write to file
webmd -o article.md https://example.com

# Custom timeout and extra wait for JS-heavy sites
webmd --timeout 60s --wait 3s https://example.com

# Use system Chrome only (no auto-download)
webmd --no-download https://example.com

# Specify Chrome binary path
webmd --browser-path /usr/bin/chromium https://example.com
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--full` | `false` | Convert full page instead of readability extraction |
| `--browser-path` | | Path to Chrome/Chromium binary |
| `--no-download` | `false` | Disable auto-download of Chromium |
| `--timeout` | `30s` | Page load timeout |
| `--wait` | `0s` | Extra wait after page load for JS-heavy sites |
| `--user-agent` | | Custom User-Agent string |
| `-o, --output` | | Write to file instead of stdout |

## Docker

```bash
docker build -t webmd .
docker run --rm webmd https://example.com
```

The Docker image uses `chromedp/headless-shell` so no external Chrome is needed.

## Browser Resolution

webmd finds Chrome/Chromium in this order:

1. `--browser-path` flag
2. `WEBMD_BROWSER_PATH` environment variable
3. System Chrome (auto-detected)
4. Auto-download via rod (unless `--no-download`)

## Building from Source

```bash
make build    # → bin/webmd
make test
make lint
make clean
```
