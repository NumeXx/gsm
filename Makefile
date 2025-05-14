# Go parameters
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(shell go env GOPATH)
GOBIN := $(GOBASE)/bin

# Target binary
BINARY_NAME = gsm
CMD_PATH = ./cmd/gsm/

# Build flags
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT_HASH ?= $(shell git rev-parse --short HEAD)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT_HASH} -X main.date=${BUILD_DATE} -s -w"
BUILD_TAGS = netgo osusergo
CGO_ENABLED = 0

.PHONY: all build clean fmt vet test install uninstall run dev_setup

all: build

build: fmt vet
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${GOBIN}
	CGO_ENABLED=${CGO_ENABLED} go build -tags '${BUILD_TAGS}' ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME} ${CMD_PATH}
	@echo "${BINARY_NAME} (version ${VERSION}) built in ${GOBIN}/"

# Statically linked binary (experimental, may need more specific ldflags for different OS)
build-static:
	@echo "Building statically linked ${BINARY_NAME}..."
	@mkdir -p ${GOBIN}
	CGO_ENABLED=0 go build -tags '${BUILD_TAGS} static_build' ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME}-static ${CMD_PATH}
	@echo "${BINARY_NAME}-static built in ${GOBIN}/"

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Vetting code..."
	@go vet ./...

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning..."
	@rm -f ${GOBIN}/${BINARY_NAME} ${GOBIN}/${BINARY_NAME}-static
	@go clean

install:
	@echo "Installing ${BINARY_NAME} to /usr/local/bin/ (requires sudo)..."
	@sudo cp ${GOBIN}/${BINARY_NAME} /usr/local/bin/${BINARY_NAME}
	@echo "${BINARY_NAME} installed. Run 'gsm' from anywhere."

uninstall:
	@echo "Uninstalling ${BINARY_NAME} from /usr/local/bin/ (requires sudo)..."
	@sudo rm -f /usr/local/bin/${BINARY_NAME}
	@echo "${BINARY_NAME} uninstalled."

run: build
	@echo "Running ${BINARY_NAME}..."
	@${GOBIN}/${BINARY_NAME}

# Setup for dev.sh - installs inotify-tools if not present
dev_setup:
	@if ! command -v inotifywait > /dev/null; then \
		echo "inotifywait not found. Attempting to install inotify-tools..."; \
		if [ -f /etc/debian_version ]; then sudo apt-get update && sudo apt-get install -y inotify-tools; \
		elif [ -f /etc/redhat-release ]; then sudo yum install -y inotify-tools; \
		elif [ -f /etc/arch-release ]; then sudo pacman -Syu --noconfirm inotify-tools; \
		else echo "Unsupported Linux distribution for automatic inotify-tools installation. Please install it manually."; exit 1; fi; \
	else \
		echo "inotifywait found."; \
	fi

# Usage: make help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all                Build the application (default)."
	@echo "  build              Build the application."
	@echo "  build-static       Build a statically linked version of the application (experimental)."
	@echo "  fmt                Format Go source files."
	@echo "  vet                Run go vet against the codebase."
	@echo "  test               Run tests."
	@echo "  clean              Remove build artifacts."
	@echo "  install            Install the binary to /usr/local/bin (requires sudo)."
	@echo "  uninstall          Remove the binary from /usr/local/bin (requires sudo)."
	@echo "  run                Build and run the application."
	@echo "  dev_setup          Install inotify-tools for live development (Linux only)."
	@echo "  help               Show this help message." 