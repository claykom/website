# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=website

# Build targets
.PHONY: all build clean test coverage lint fmt vet deps help

## help: Show this help message
help:
	@echo "Available targets:"
	@echo "  all              Run fmt, vet, lint, test, and build"
	@echo "  build            Build the binary"
	@echo "  clean            Clean build artifacts"
	@echo "  test             Run all tests"
	@echo "  coverage         Run tests with coverage report"
	@echo "  coverage-html    Generate HTML coverage report"
	@echo "  test-verbose     Run tests with verbose output"
	@echo "  test-race        Run tests with race detection"
	@echo "  bench            Run benchmarks"
	@echo "  fmt              Format Go code"
	@echo "  vet              Run go vet"
	@echo "  lint             Run golint (requires golint to be installed)"
	@echo "  deps             Download and tidy dependencies"
	@echo "  run              Build and run the application"
	@echo "  dev              Run in development mode"
	@echo "  docker-build     Build Docker image"
	@echo "  docker-run       Run Docker container"
	@echo "  security         Run security checks"
	@echo "  help             Show this help message"

## all: Run fmt, vet, lint, test, and build
all: fmt vet test build

## build: Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -ldflags="-w -s" -v ./

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -f *_coverage.out
	rm -f coverage.html

## test: Run all tests
test:
	$(GOTEST) ./...

## coverage: Run tests with coverage report
coverage:
	$(GOTEST) -cover ./...

## coverage-html: Generate HTML coverage report
coverage-html:
	$(GOTEST) -coverprofile=middleware_coverage.out ./...
	go tool cover -html=middleware_coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-verbose: Run tests with verbose output
test-verbose:
	$(GOTEST) -v ./...

## test-race: Run tests with race detection
test-race:
	$(GOTEST) -race ./...

## bench: Run benchmarks
bench:
	$(GOTEST) -bench=. ./...

## fmt: Format Go code
fmt:
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	$(GOCMD) vet ./...

## lint: Run golint (requires golint to be installed)
lint:
	golint ./...

## deps: Download and tidy dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## run: Build and run the application
run: build
	./$(BINARY_NAME)

## dev: Run in development mode
dev:
	$(GOCMD) run main.go

## docker-build: Build Docker image
docker-build:
	docker build -t $(BINARY_NAME) .

## docker-run: Run Docker container
docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

## security: Run security checks
security:
	$(GOCMD) vet ./...
	@echo "Run 'gosec ./...' for additional security checks (requires gosec)"