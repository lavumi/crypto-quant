.PHONY: build build-full build-frontend clean clean-frontend clean-all test run dev help

# Default target
all: build

# Build server binary (API + Collector + Backtest in one)
build:
	@echo "🔨 Building server..."
	@cd backend && go build -o ../server ./cmd/api
	@echo "✓ Server built successfully: ./server"

# Build everything including frontend (for production)
build-full: build-frontend build
	@echo "✓ Full build complete!"

# Build frontend and embed it
build-frontend:
	@echo "🎨 Building frontend..."
	@cd frontend && pnpm build
	@echo "📦 Copying frontend to backend for embedding..."
	@rm -rf backend/internal/api/frontend/build
	@mkdir -p backend/internal/api/frontend/build
	@cp -r frontend/build/* backend/internal/api/frontend/build/
	@echo "✓ Frontend ready for embedding"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f server
	@rm -rf backend/bin/

# Clean frontend build
clean-frontend:
	@echo "🧹 Cleaning frontend build..."
	@rm -rf backend/internal/api/frontend/build
	@cd frontend && rm -rf build .svelte-kit

# Clean everything
clean-all: clean clean-frontend
	@echo "✓ All clean!"

# Run tests
test:
	@echo "🧪 Running tests..."
	@cd backend && go test -v ./...

# Run server in production mode
run:
	@./server

# Run server in development mode (no build)
dev:
	@cd backend && go run ./cmd/api

# Run with collector mode
collect:
	@./server --collect

# Install dependencies
deps:
	@echo "📥 Installing Go dependencies..."
	@cd backend && go mod download && go mod tidy
	@echo "📥 Installing frontend dependencies..."
	@cd frontend && pnpm install

# Format code
fmt:
	@echo "✨ Formatting code..."
	@cd backend && go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	@cd backend && golangci-lint run

# Show help
help:
	@echo "╔══════════════════════════════════════════════════════════╗"
	@echo "║          Crypto Quant - Build System                     ║"
	@echo "╚══════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "Main targets:"
	@echo "  make              - Build server binary (default)"
	@echo "  make build        - Build server binary"
	@echo "  make build-full   - Build frontend + server (production)"
	@echo ""
	@echo "Frontend:"
	@echo "  make build-frontend - Build frontend and embed"
	@echo ""
	@echo "Clean:"
	@echo "  make clean        - Remove server binary"
	@echo "  make clean-frontend - Remove frontend build"
	@echo "  make clean-all    - Remove all build artifacts"
	@echo ""
	@echo "Run:"
	@echo "  make run          - Run server (./server)"
	@echo "  make dev          - Run in development mode (no build)"
	@echo "  make collect      - Run in collector mode"
	@echo ""
	@echo "Development:"
	@echo "  make deps         - Install dependencies"
	@echo "  make test         - Run tests"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Lint code"
	@echo ""
	@echo "Usage examples:"
	@echo "  ./server                                    # Start API server"
	@echo "  ./server --help                             # Show all options"
	@echo "  ./server --collect --symbol BTCUSDT --days 7  # Collect data"


