# Variables
BINARY_NAME=asmfmt
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "devel")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.gitHash=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)"

.PHONY: all build test test-update vet fmt install clean help

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/asmfmt

test:
	go test ./...

test-update:
	go test -run TestRewrite -update

vet:
	go vet ./...

fmt:
	gofmt -w .

install:
	go install $(LDFLAGS) ./cmd/asmfmt

clean:
	rm -f $(BINARY_NAME)
	find . -name "*.asmfmt" -type f -delete

help:
	@echo "Available targets:"
	@echo "  build        - Build the $(BINARY_NAME) binary with version info"
	@echo "  test         - Run all unit and integration tests"
	@echo "  test-update  - Run TestRewrite and update test golden fixtures"
	@echo "  vet          - Run go vet static checks"
	@echo "  fmt          - Format all Go source files"
	@echo "  install      - Install the binary into GOPATH/bin"
	@echo "  clean        - Remove build artifacts and diagnostic files"
