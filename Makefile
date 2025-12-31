# Sway Layout Manager Makefile

BINARY_NAME = sway-layout-manager
PACKAGE_PATH = github.com/lidsol/sway-layout-manager
CMD_PATH = ./cmd/swaylayoutmgr
VERSION = 0.1.0-dev
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# build flags
BUILD_FLAGS = -trimpath

# directories
PREFIX = /usr/local
BINDIR = $(PREFIX)/bin
MANDIR = $(PREFIX)/share/man/man1

# default target
.PHONY: all
all: build

# build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) $(CMD_PATH)

# build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-linux-arm64 $(CMD_PATH)

# install the binary
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	install -d $(BINDIR)
	install -m 755 $(BINARY_NAME) $(BINDIR)/$(BINARY_NAME)

# uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "Removing $(BINARY_NAME) from $(BINDIR)..."
	rm -f $(BINDIR)/$(BINARY_NAME)

# run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# test with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# lint the code
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# tidy up dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out coverage.html

# development build (with debug info)
.PHONY: dev
dev:
	@echo "Building development version..."
	go build -o $(BINARY_NAME) $(CMD_PATH)

# quick test of basic functionality
.PHONY: test-cli
test-cli: build
	@echo "Testing CLI functionality..."
	./$(BINARY_NAME) --version
	./$(BINARY_NAME) --help

# run the application
.PHONY: run
run:
	go run $(CMD_PATH) $(ARGS)

# show help
.PHONY: help
help:
	@cat makefile-help.txt