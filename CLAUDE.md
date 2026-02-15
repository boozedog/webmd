# webmd

## MANDATORY: Use td for Task Management

Run td usage --new-session at conversation start (or after /clear). This tells you what to work on next.

Sessions are automatic (based on terminal/agent context). Optional:
- td session "name" to label the current session
- td session --new to force a new session in the same context

Use td usage -q after first read.

URL to agent-friendly markdown CLI using headless Chrome.

## Build & Run

```bash
make build           # builds bin/webmd
make test            # go test ./...
make lint            # go vet ./...
make docker          # docker build
make clean           # rm bin/
```

## Quick Test

```bash
bin/webmd https://example.com              # readability mode
bin/webmd --full https://example.com       # full page mode
bin/webmd --no-download https://example.com # system Chrome only
```

## Module

- Path: `github.com/boozedog/webmd`
- Go 1.24+
- Key deps: go-rod/rod, mackee/go-readability, JohannesKaufmann/html-to-markdown/v2, spf13/cobra
