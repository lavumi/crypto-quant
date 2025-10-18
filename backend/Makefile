.PHONY: build clean test run-api run-collector run-backtest help

# Build all binaries
build: build-api build-collector build-backtest

# Build individual binaries
build-api:
	@echo "Building API server..."
	@go build -o bin/api cmd/api/main.go

build-collector:
	@echo "Building data collector..."
	@go build -o bin/collector cmd/collector/main.go

build-backtest:
	@echo "Building backtesting engine..."
	@go build -o bin/backtest cmd/backtest/main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@mkdir -p bin/

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run binaries
run-api:
	@./bin/api

run-collector:
	@./bin/collector

run-backtest:
	@./bin/backtest

# Development helpers
dev-api:
	@go run cmd/api/main.go

dev-collector:
	@go run cmd/collector/main.go

dev-backtest:
	@go run cmd/backtest/main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build all binaries"
	@echo "  build-api       - Build API server"
	@echo "  build-collector - Build data collector"
	@echo "  build-backtest  - Build backtesting engine"
	@echo "  clean           - Remove build artifacts"
	@echo "  test            - Run tests"
	@echo "  run-api         - Run API server"
	@echo "  run-collector   - Run data collector"
	@echo "  run-backtest    - Run backtest"
	@echo "  dev-api         - Run API in development mode"
	@echo "  dev-collector   - Run collector in development mode"
	@echo "  dev-backtest    - Run backtest in development mode"
	@echo "  deps            - Install dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  help            - Show this help message"






