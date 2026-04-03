# Crypto Quant Trading System

A full-stack quantitative trading platform for cryptocurrency markets.

## Overview

A systematic algorithmic trading system with Go backend and Svelte frontend, targeting cryptocurrency exchanges like Binance.

## Project Structure

```
crypto-quant/
├── backend/          # Go-based trading engine & API
│   ├── cmd/         # Command-line applications (API, collector, backtest)
│   ├── internal/    # Internal packages (domain, services, handlers)
│   └── pkg/         # Shared packages
└── frontend/        # Svelte + TypeScript web dashboard
    └── src/         # Frontend source code
```

## Development Roadmap

### Phase 1: Virtual Trading System
- [x] Real-time price data collection
- [x] Virtual balance management
- [x] Manual trading simulation
- [x] Portfolio tracking

### Phase 2: Backtesting Engine
- [x] Historical data collection and storage
- [x] Backtesting framework
- [x] Performance metrics (Sharpe Ratio, MDD, etc.)
- [x] Strategy interface and MA Crossover example
- [ ] Visualization tools
- [ ] Advanced strategies (RSI, MACD, Bollinger Bands)
- [ ] Parameter optimization

### Phase 3: Quantitative Trading Logic
- [ ] Strategy interface design
- [ ] Technical indicators implementation
- [ ] Strategy parameter optimization
- [ ] Risk management system

### Phase 4: Live Trading System
- [ ] Binance API integration
- [ ] Automated trading system
- [ ] Monitoring and alerts
- [ ] Logging and audit trail

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Exchange**: Binance API
- **Database**: SQLite (development)
- **Framework**: Gin (REST API)

### Frontend
- **Framework**: Svelte 5 + SvelteKit
- **Language**: TypeScript
- **Build Tool**: Vite
- **Package Manager**: PNPM

## 📖 Documentation

- **[⚡ 빠른 시작 (Quick Start)](QUICKSTART.md)** - 5분 안에 시작하기
- **[📘 사용 가이드 (Usage Guide)](docs/USAGE_GUIDE.md)** - 데이터 수집, 백테스트 실행 등 모든 사용법
- **[🚀 배포 가이드 (Deployment Guide)](docs/DEPLOYMENT.md)** - 프로덕션 빌드 및 배포 방법
- [📊 백테스팅 가이드](docs/BACKTEST.md) - 백테스팅 상세 설명
- [🔌 백테스트 API 가이드](docs/API_BACKTEST.md) - REST API로 백테스트 사용하기
- [🎯 Phase 2 완료 요약](docs/PHASE2_SUMMARY.md) - Phase 2 구현 내역
- [📝 변경사항 로그](CHANGELOG.md) - 전체 변경 이력

## Getting Started

### Prerequisites

- **Go 1.21+** - Backend development
- **Node.js 22+** - Frontend development
- **PNPM** - Frontend package manager

### Development Setup

#### Backend Only

```bash
cd backend

# Install dependencies
go mod download

# Build all binaries
make build
# or build individually:
go build -o bin/api cmd/api/main.go
go build -o bin/collector cmd/collector/main.go
go build -o bin/backtest cmd/backtest/main.go
```

#### Frontend Development

```bash
cd frontend

# Install dependencies
pnpm install

# Start development server (with hot reload)
pnpm dev
```

### Production Build (with Embedded Frontend)

Build a single binary that includes the frontend:

```bash
# Option 1: Using Makefile (recommended)
cd backend
make build-full

# Option 2: Using build script
./scripts/build.sh

# Run the all-in-one binary
cd backend
./bin/api

# Access at:
# - Frontend: http://localhost:8080
# - API: http://localhost:8080/api/v1
# - Swagger: http://localhost:8080/swagger/index.html
```

> 💡 **Note**: The production binary embeds the frontend using Go's `embed` package, creating a single deployable binary.
> 
> For detailed build and deployment instructions, see the **[Deployment Guide](docs/DEPLOYMENT.md)**.

### Collecting Data

```bash
# Collect historical data (required before backtesting)
./server collect --symbol BTCUSDT --interval 1h --days 90 --db data/trading.db
./server collect --symbol BTCUSDT --interval 1d --days 365 --db data/trading.db
```

### Running Backtest

#### CLI 사용

```bash
# Run backtest with default settings
./server backtest --symbol BTCUSDT --interval 1h --db data/trading.db

# Run backtest with custom parameters
./server backtest \
  --symbol ETHUSDT \
  --interval 4h \
  --start 2025-07-01 \
  --end 2025-10-17 \
  --balance 10000 \
  --commission 0.001 \
  --fast 10 \
  --slow 30 \
  --db data/trading.db
```

#### API 사용 (NEW!)

```bash
cd backend

# Start API server
./bin/api --port 8080

# Run backtest via API (from another terminal)
curl -X POST http://localhost:8080/api/v1/backtest/run \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "interval": "1h",
    "start_date": "2025-07-19",
    "end_date": "2025-10-17",
    "strategy": "ma_cross",
    "fast_period": 10,
    "slow_period": 30
  }'
```

> 💡 **Tip**: See [API Backtest Guide](docs/API_BACKTEST.md) for detailed API usage and examples.


## Features

### Current Features
- ✅ Real-time price data collection
- ✅ Virtual balance management
- ✅ Comprehensive backtesting framework
- ✅ REST API for backtesting
- ✅ Multiple strategy support (MA Cross, RSI, BB+RSI)
- ✅ Performance metrics (Sharpe Ratio, MDD, Win Rate, etc.)

### Planned Features
- 🔄 Web-based dashboard (in progress)
- 📊 Interactive charts and visualization
- 🎯 Risk management tools
- 🤖 Live trading capabilities
- 📱 Real-time monitoring and alerts

## License

TBD

## Contributing

TBD
