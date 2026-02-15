FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o /webmd .

FROM chromedp/headless-shell:latest
COPY --from=builder /webmd /usr/local/bin/webmd
ENV WEBMD_BROWSER_PATH=/headless-shell/headless-shell
ENTRYPOINT ["webmd"]
