.PHONY: build build-full build-frontend clean clean-frontend test run-api run-collector run-backtest run-sweep run-walk-forward help

# Build all binaries (without frontend)
build: build-api build-collector build-backtest

# Build everything including frontend (for production)
build-full: build-frontend build

# Build frontend and embed it
build-frontend:
	@echo "Building frontend..."
	@cd ../frontend && pnpm build
	@echo "Copying frontend to backend for embedding..."
	@rm -rf internal/api/frontend/build
	@mkdir -p internal/api/frontend/build
	@cp -r ../frontend/build/* internal/api/frontend/build/
	@echo "✓ Frontend ready for embedding"

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

# Clean frontend build
clean-frontend:
	@echo "Cleaning frontend build..."
	@rm -rf internal/api/frontend/build
	@cd ../frontend && rm -rf build .svelte-kit

# Clean everything
clean-all: clean clean-frontend

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run binaries
run-api:
	@./bin/api api

run-collector:
	@./bin/api collect

run-backtest:
	@./bin/api backtest

run-sweep:
	@./bin/api sweep

run-walk-forward:
	@./bin/api walk-forward

# Development helpers
dev-api:
	@go run cmd/api/main.go

dev-collector:
	@go run cmd/api/main.go collect

dev-backtest:
	@go run cmd/api/main.go backtest

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
	@echo "  build           - Build all binaries (without frontend)"
	@echo "  build-full      - Build everything including frontend (production)"
	@echo "  build-frontend  - Build frontend and prepare for embedding"
	@echo "  build-api       - Build API server"
	@echo "  build-collector - Build data collector"
	@echo "  build-backtest  - Build backtesting engine"
	@echo "  clean           - Remove build artifacts"
	@echo "  clean-frontend  - Remove frontend build"
	@echo "  clean-all       - Remove all build artifacts"
	@echo "  test            - Run tests"
	@echo "  run-api         - Run API server"
	@echo "  run-collector   - Run data collector"
	@echo "  run-backtest    - Run single backtest"
	@echo "  run-sweep       - Run parameter sweep"
	@echo "  run-walk-forward - Run walk-forward validation"
	@echo "  dev-api         - Run API in development mode"
	@echo "  dev-collector   - Run collector in development mode"
	@echo "  dev-backtest    - Run backtest in development mode"
	@echo "  deps            - Install dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  help            - Show this help message"





