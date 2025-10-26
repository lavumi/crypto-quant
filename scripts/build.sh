#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Building Crypto Quant ===${NC}"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${YELLOW}Project root: ${PROJECT_ROOT}${NC}"

# Step 1: Build Frontend
echo -e "\n${GREEN}[1/3] Building frontend...${NC}"
cd "${PROJECT_ROOT}/frontend"

if ! command -v pnpm &> /dev/null; then
    echo -e "${RED}Error: pnpm is not installed${NC}"
    exit 1
fi

pnpm build

if [ ! -d "build" ]; then
    echo -e "${RED}Error: Frontend build directory not found${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Frontend built successfully${NC}"

# Step 2: Copy frontend build to backend for embedding
echo -e "\n${GREEN}[2/3] Copying frontend build to backend...${NC}"
EMBED_DIR="${PROJECT_ROOT}/backend/internal/api/frontend/build"
rm -rf "${EMBED_DIR}"
mkdir -p "${EMBED_DIR}"
cp -r build/* "${EMBED_DIR}/"

echo -e "${GREEN}✓ Frontend copied to: ${EMBED_DIR}${NC}"

# Step 3: Build Backend
echo -e "\n${GREEN}[3/3] Building backend...${NC}"
cd "${PROJECT_ROOT}/backend"

# Build API server
echo "Building api..."
go build -o bin/api ./cmd/api

# Build other binaries
echo "Building collector..."
go build -o bin/collector ./cmd/collector

echo "Building backtest..."
go build -o bin/backtest ./cmd/backtest

echo -e "${GREEN}✓ Backend built successfully${NC}"

# Summary
echo -e "\n${GREEN}=== Build Complete ===${NC}"
echo -e "${GREEN}Binaries:${NC}"
echo -e "  • ${PROJECT_ROOT}/backend/bin/api"
echo -e "  • ${PROJECT_ROOT}/backend/bin/collector"
echo -e "  • ${PROJECT_ROOT}/backend/bin/backtest"
echo -e "\n${YELLOW}To run the API server with embedded frontend:${NC}"
echo -e "  cd ${PROJECT_ROOT}/backend"
echo -e "  ./bin/api"
echo -e "\n${YELLOW}Then open: http://localhost:8080${NC}"




