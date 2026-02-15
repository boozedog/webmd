VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test lint docker clean

build:
	go build -ldflags="$(LDFLAGS)" -o bin/webmd .

test:
	go test ./...

lint:
	go vet ./...

docker:
	docker build --build-arg VERSION=$(VERSION) -t webmd .

clean:
	rm -rf bin/
