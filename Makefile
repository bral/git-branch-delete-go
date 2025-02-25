.PHONY: all setup build test test-race test-cover lint lint-fix clean generate-mocks release

# Variables
BINARY_NAME=git-branch-delete
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X github.com/bral/git-branch-delete-go/cmd.Version=${VERSION} -X github.com/bral/git-branch-delete-go/cmd.BuildTime=${BUILD_TIME} -X github.com/bral/git-branch-delete-go/cmd.GitCommit=${COMMIT_HASH}"

all: setup test lint build

setup:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest
	go mod download
	go mod tidy

build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME}

test:
	@echo "Running tests..."
	go test -v ./...

test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

test-cover:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	@echo "Running linter..."
	golangci-lint run

lint-fix:
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix

generate-mocks:
	@echo "Generating mocks..."
	mockgen -source=pkg/git/git.go -destination=test/mocks/mock_git.go
	mockgen -source=internal/ui/ui.go -destination=test/mocks/mock_ui.go

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -testcache

release:
	@echo "Building release version..."
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}_darwin_arm64
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}_linux_amd64
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}_windows_amd64.exe
	@echo "Creating checksums..."
	cd bin && sha256sum ${BINARY_NAME}* > checksums.txt
