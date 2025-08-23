# Makefile for My IP

# Variables
BINARY_NAME=myip
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
.PHONY: help
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Development

## swagger: Generate Swagger documentation
.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag > /dev/null; then \
		swag init -g main.go -o docs/; \
	else \
		echo "swag not found. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g main.go -o docs/; \
	fi

## build: Build the application
.PHONY: build
build: swagger
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

## run: Run the application
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run . 

## dev: Run in development mode with hot reload
.PHONY: dev
dev:
	@echo "Starting development server..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		go run .; \
	fi

##@ Testing

## test: Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

## test-race: Run tests with race detector
.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	go test -race -v ./...

## test-cover: Run tests with coverage
.PHONY: test-cover
test-cover:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-coverage-ci: Run tests with coverage for CI (with race detector)
.PHONY: test-coverage-ci
test-coverage-ci:
	@echo "Running tests with coverage for CI..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

## bench: Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

##@ Code Quality

## fmt: Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

## vet: Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

## lint: Run golint
.PHONY: lint
lint:
	@echo "Running golint..."
	@if command -v golint > /dev/null; then \
		golint ./...; \
	else \
		echo "golint not found. Install with: go install golang.org/x/lint/golint@latest"; \
	fi

## staticcheck: Run staticcheck
.PHONY: staticcheck
staticcheck:
	@echo "Running staticcheck..."
	@if command -v staticcheck > /dev/null; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not found. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

## check: Run all code quality checks
.PHONY: check
check: fmt vet lint staticcheck test

##@ Build & Release

## build-all: Build for all platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@mkdir -p build
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-windows-amd64.exe .

## release-dry: Dry run GoReleaser
.PHONY: release-dry
release-dry:
	@echo "Running GoReleaser dry run..."
	@if command -v goreleaser > /dev/null; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "GoReleaser not found. Install from: https://goreleaser.com/install/"; \
	fi

## release: Run GoReleaser
.PHONY: release
release:
	@echo "Running GoReleaser..."
	@if command -v goreleaser > /dev/null; then \
		goreleaser release --clean; \
	else \
		echo "GoReleaser not found. Install from: https://goreleaser.com/install/"; \
	fi

##@ Docker

## docker-build: Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

## docker-run: Run Docker container
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME):latest

## docker-build-goreleaser: Build Docker image using GoReleaser Dockerfile
.PHONY: docker-build-goreleaser
docker-build-goreleaser: build
	@echo "Building Docker image with GoReleaser Dockerfile..."
	docker build -f Dockerfile.goreleaser -t $(BINARY_NAME):goreleaser .

## docker-test-build: Build Docker image for testing (no push)
.PHONY: docker-test-build
docker-test-build:
	@echo "Building Docker image for testing..."
	docker build -t $(BINARY_NAME):test .

##@ Maintenance

## clean: Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -rf build/
	@rm -rf dist/
	@rm -f coverage.out coverage.html
	@rm -f *.test
	@rm -f *.log
	@rm -f docs/docs.go docs/swagger.json docs/swagger.yaml

## deps: Download and verify dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

## tidy: Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

## upgrade: Upgrade dependencies
.PHONY: upgrade
upgrade:
	@echo "Upgrading dependencies..."
	go get -u ./...
	go mod tidy

##@ Install

## install: Install the application
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .

## uninstall: Uninstall the application
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)

##@ CI/CD

## ci-setup: Setup CI environment
.PHONY: ci-setup
ci-setup:
	@echo "Setting up CI environment..."
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/lint/golint@latest
	go install github.com/swaggo/swag/cmd/swag@latest

## ci-test: Run CI tests
.PHONY: ci-test
ci-test: deps swagger vet staticcheck test-race

## security: Run security checks
.PHONY: security
security:
	@echo "Running security checks..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

## security-sarif: Run security checks with SARIF output
.PHONY: security-sarif
security-sarif:
	@echo "Running security checks with SARIF output..."
	@if command -v gosec > /dev/null; then \
		gosec -no-fail -fmt sarif -out results.sarif ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

## all: Run all checks and build
.PHONY: all
all: clean deps swagger check build test-cover

.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
