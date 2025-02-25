.PHONY: all build test clean install deps lint release snapshot

BINARY_NAME=git-branch-delete
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X github.com/bral/git-branch-delete-go/cmd.Version=${VERSION} -X github.com/bral/git-branch-delete-go/cmd.CommitSHA=${COMMIT_SHA} -X github.com/bral/git-branch-delete-go/cmd.BuildTime=${BUILD_TIME}"

# Default target
all: deps build test

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build the binary
build: deps
	@echo "Building ${BINARY_NAME} version ${VERSION}..."
	mkdir -p bin
	CGO_ENABLED=0 go build ${LDFLAGS} -o bin/${BINARY_NAME}

# Run tests
test: deps
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -testcache

# Install to $GOPATH/bin
install: build
	@echo "Installing to ${GOPATH}/bin..."
	mv bin/${BINARY_NAME} ${GOPATH}/bin/${BINARY_NAME}

# Create a new release
release:
	@echo "Creating release ${VERSION}..."
	goreleaser release --clean

# Create a snapshot release
snapshot:
	@echo "Creating snapshot..."
	goreleaser release --snapshot --clean

# Run the program
run: build
	@echo "Running ${BINARY_NAME}..."
	./bin/${BINARY_NAME}

# Show version info
version:
	@echo "Version: ${VERSION}"
	@echo "Commit: ${COMMIT_SHA}"
	@echo "Build Time: ${BUILD_TIME}"

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Install dependencies, build, and test"
	@echo "  deps          - Install dependencies"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  release       - Create a new release"
	@echo "  snapshot      - Create a snapshot release"
	@echo "  run           - Run the program"
	@echo "  version       - Show version info"
	@echo "  help          - Show this help"
