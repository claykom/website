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
	@sed -n 's/^##//p' $(MAKEFILE_LIST)

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
	$(GOTEST) ./internal/config ./internal/handlers ./internal/middleware

## coverage: Run tests with coverage report
coverage:
	$(GOTEST) -cover ./internal/config ./internal/handlers ./internal/middleware

## coverage-html: Generate HTML coverage report
coverage-html:
	$(GOTEST) -coverprofile=middleware_coverage.out ./internal/middleware
	go tool cover -html=middleware_coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-verbose: Run tests with verbose output
test-verbose:
	$(GOTEST) -v ./internal/config ./internal/handlers ./internal/middleware

## test-race: Run tests with race detection
test-race:
	$(GOTEST) -race ./internal/config ./internal/handlers ./internal/middleware

## bench: Run benchmarks
bench:
	$(GOTEST) -bench=. ./internal/middleware

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

## ci: Run CI pipeline (fmt, vet, test, build)
ci: fmt vet test build
	@echo "CI pipeline completed successfully"