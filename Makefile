.PHONY: all build clean test install uninstall

# Build configuration
BINARY_NAME = psiphond-ng
BUILD_DIR = build
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0
LDFLAGS ?= -s -w
TAGS ?=

VERSION ?= 1.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS += -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

all: build

build:
	@echo "Building $(BINARY_NAME) $(VERSION)"
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) \
		go build -trimpath -ldflags="$(LDFLAGS)" $(TAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/psiphond-ng
	@echo "Binary created: $(BUILD_DIR)/$(BINARY_NAME)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

build-static:
	@$(MAKE) build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 LDFLAGS="$(LDFLAGS) -extldflags '-static'"

build-tun:
	@$(MAKE) build TAGS="tun"

build-all:
	@echo "Building for multiple architectures..."
	@for arch in amd64 arm64; do \
		echo "Building for linux/$$arch..."; \
		GOOS=linux GOARCH=$$arch $(MAKE) build; \
		mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-linux-$$arch; \
	done

test:
	go test ./... -v

test-race:
	go test ./... -race

vet:
	go vet ./...

fmt:
	go fmt ./...

lint:
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

clean:
	rm -rf $(BUILD_DIR)
	go clean

install:
	sudo ./scripts/install.sh

uninstall:
	sudo ./scripts/uninstall.sh

run:
	./$(BUILD_DIR)/$(BINARY_NAME) -config psiphond-dev.conf

run-debug:
	LOG_LEVEL=debug ./$(BUILD_DIR)/$(BINARY_NAME) -config psiphond-dev.conf

deps:
	go mod download
	go mod verify
	go mod tidy

install-deps-ubuntu:
	sudo apt-get update
	sudo apt-get install -y golang-go build-essential

install-deps-fedora:
	sudo dnf install -y golang gcc

install-deps-arch:
	sudo pacman -S --needed go base-devel

release: clean build-static
	@echo "Creating release..."
	@mkdir -p release
	@cp $(BUILD_DIR)/$(BINARY_NAME) release/$(BINARY_NAME)-$(GOOS)-$(GOARCH)
	@if command -v upx &> /dev/null; then \
		echo "Compressing with upx..."; \
		upx --best --lzma release/$(BINARY_NAME)-$(GOOS)-$(GOARCH); \
	fi
	@tar czf release/$(BINARY_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz -C release $(BINARY_NAME)-$(GOOS)-$(GOARCH)
	@echo "Release created: release/$(BINARY_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz"

.PHONY: help
help:
	@echo "PsiphonNGLinux Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all              - Alias for build"
	@echo "  build            - Build binary (default GoOS/GoArch)"
	@echo "  build-static     - Build static binary"
	@echo "  build-tun        - Build with TUN support (adds -tags tun)"
	@echo "  build-all        - Build for all supported architectures"
	@echo "  test             - Run tests"
	@echo "  test-race        - Run tests with race detector"
	@echo "  vet              - Run go vet"
	@echo "  fmt              - Format source code"
	@echo "  lint             - Run linter (requires golangci-lint)"
	@echo "  clean            - Clean build artifacts"
	@echo "  install          - Install system-wide (requires sudo)"
	@echo "  uninstall        - Uninstall from system"
	@echo "  run              - Build and run in foreground"
	@echo "  run-debug        - Build and run with debug logging"
	@echo "  deps             - Download and verify dependencies"
	@echo "  release          - Create compressed release binary"
	@echo ""
	@echo "Variables:"
	@echo "  GOOS=$(GOOS)         - Target OS (linux, windows, darwin)"
	@echo "  GOARCH=$(GOARCH)     - Target architecture (amd64, arm64, 386)"
	@echo "  VERSION=$(VERSION)   - Version string"
	@echo "  COMMIT=$(COMMIT)     - Git commit short SHA"
	@echo ""
	@echo "Examples:"
	@echo "  make build GOARCH=arm64"
	@echo "  make release VERSION=1.0.0"
	@echo "  sudo make install"
