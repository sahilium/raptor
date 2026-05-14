# Raptor Makefile

# Variables
BINARY_NAME=raptor
BUILD_DIR=bin
CMD_PATH=./cmd/raptor
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

# .PHONY ensures that these targets are always run, even if a file with the same name exists
.PHONY: all build clean test lint vet check help

all: check build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //g' -e 's/: /: /'

## build: Build the raptor binary
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ${CMD_PATH}

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf ${BUILD_DIR}
	rm -f coverage.out

## test: Run unit tests
test:
	@echo "Running tests..."
	go test -v -race ./...

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## check: Run all checks (vet + test)
check: vet test

## run: Build and run the binary locally
run: build
	./${BUILD_DIR}/${BINARY_NAME}
