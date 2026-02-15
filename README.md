# webmd

Convert web pages to agent-friendly markdown using headless Chrome.

Fetches URLs with a real browser (renders JavaScript), then extracts clean markdown. Two modes:

- **Full page** (default) — converts the entire page
- **Article** — extracts main article content via readability

## Install

```bash
go install github.com/boozedog/webmd@latest
```

From a local clone:

```bash
go install .
```

Or with Docker:

```bash
docker pull ghcr.io/boozedog/webmd
```

## Usage

```bash
# Full page conversion (default)
webmd https://example.com

# Extract main article content via readability
webmd --article https://example.com

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
| `--article` | `false` | Extract main article content via readability |
| `--mobile` | `false` | Emulate a mobile device (iPhone viewport and user-agent) |
| `--images` | `false` | Include images in markdown output |
| `--browser-path` | | Path to Chrome/Chromium binary |
| `--no-download` | `false` | Disable auto-download of Chromium |
| `--timeout` | `15s` | Page load timeout |
| `--wait` | `0s` | Extra wait after page load for JS-heavy sites |
| `--user-agent` | | Custom User-Agent string |
| `-o, --output` | | Write to file instead of stdout |

## Server Mode

Run `webmd serve` to start an HTTP server with a persistent browser instance:

```bash
webmd serve                    # listen on 0.0.0.0:8080
webmd serve --port 9090        # custom port
webmd serve --host 127.0.0.1   # bind to localhost only
```

Convert pages via GET request:

```
GET /?url=https://example.com
GET /?url=https://example.com&article
GET /?url=https://example.com&preview
GET /?url=https://example.com&timeout=30s&wait=2s&user-agent=MyBot
```

Returns plain text markdown by default, or rendered HTML with `&preview`. Query parameters:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `url` | *(required)* | URL to convert |
| `article` | `false` | Extract main article content via readability |
| `mobile` | `false` | Emulate a mobile device (iPhone viewport and user-agent) |
| `images` | `false` | Include images in markdown output |
| `preview` | `false` | Return rendered HTML instead of markdown |
| `timeout` | `15s` | Page load timeout |
| `wait` | `0s` | Extra wait after page load |
| `user-agent` | | Custom User-Agent string |

## Docker

```bash
docker build -t webmd .

# Server mode (default)
docker run --rm -p 8080:8080 webmd

# One-shot CLI
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
