# webmd

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
