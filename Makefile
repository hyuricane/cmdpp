BINARY_NAME := cmdpp
PREFIX ?= $(HOME)/.local
INSTALL_DIR ?= $(PREFIX)/bin
GO := go

.PHONY: all build build-all install uninstall test test-coverage fmt vet tidy clean run help

# Default target
all: build

## build: Compile the cmdpp binary
build:
	@echo "==> Building $(BINARY_NAME)..."
	$(GO) build -ldflags "-s -w" -o $(BINARY_NAME) main.go
	@echo "✓ Binary built: ./$(BINARY_NAME)"

## build-all: Cross-compile cmdpp for supported platforms into dist/
build-all:
	@echo "==> Building for all platforms..."
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags "-s -w" -o dist/$(BINARY_NAME)_linux_amd64 main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -trimpath -ldflags "-s -w" -o dist/$(BINARY_NAME)_linux_arm64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags "-s -w" -o dist/$(BINARY_NAME)_darwin_amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -ldflags "-s -w" -o dist/$(BINARY_NAME)_darwin_arm64 main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags "-s -w" -o dist/$(BINARY_NAME)_windows_amd64.exe main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GO) build -trimpath -ldflags "-s -w" -o dist/$(BINARY_NAME)_windows_arm64.exe main.go
	@echo "✓ Multi-platform binaries built in dist/"

## install: Install cmdpp to $(INSTALL_DIR)
install: build
	@echo "==> Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Installed to $(INSTALL_DIR)/$(BINARY_NAME)"

## uninstall: Remove cmdpp from $(INSTALL_DIR)
uninstall:
	@echo "==> Removing $(INSTALL_DIR)/$(BINARY_NAME)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Uninstalled"

## test: Run all unit tests
test:
	@echo "==> Running tests..."
	$(GO) test -v ./...

## test-coverage: Run tests with code coverage output
test-coverage:
	@echo "==> Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

## fmt: Format Go source code
fmt:
	@echo "==> Formatting code..."
	$(GO) fmt ./...

## vet: Run go vet linter
vet:
	@echo "==> Running go vet..."
	$(GO) vet ./...

## tidy: Tidy go.mod dependencies
tidy:
	@echo "==> Tidying dependencies..."
	$(GO) mod tidy

## run: Run cmdpp directly from source
run:
	$(GO) run main.go $(ARGS)

## clean: Remove build artifacts and temporary files
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -f $(BINARY_NAME) coverage.out coverage.html *.test
	@rm -rf bin/ dist/
	@echo "✓ Clean complete"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
