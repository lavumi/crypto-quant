package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/api"
	"github.com/lavumi/crypto-quant/internal/api/handler"
	"github.com/lavumi/crypto-quant/internal/datasource/database"
	"github.com/lavumi/crypto-quant/internal/datasource/exchange"
	"github.com/lavumi/crypto-quant/internal/datasource/market/history"
	"github.com/lavumi/crypto-quant/internal/datasource/market/price"
	"github.com/lavumi/crypto-quant/internal/portfolio"
	"github.com/lavumi/crypto-quant/internal/portfolio/wallet"

	_ "github.com/lavumi/crypto-quant/docs"
)

// @title Crypto Quant API
// @version 1.0
// @description Cryptocurrency quantitative trading API server
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
╔══════════════════════════════════════════════════════════════╗
║          Crypto Quant - Trading Analysis Platform            ║
╚══════════════════════════════════════════════════════════════╝

USAGE:
  ./server [OPTIONS]

MODES:
  Default      Start API server with web interface
  --collect    Run historical data collector

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
API SERVER OPTIONS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`)
		fmt.Fprintf(os.Stderr, "  --port          Server port (default: 8080)\n")
		fmt.Fprintf(os.Stderr, "  --db            SQLite database path (default: data/trading.db)\n")
		fmt.Fprintf(os.Stderr, "  --api-key       Binance API key (optional)\n")
		fmt.Fprintf(os.Stderr, "  --secret-key    Binance secret key (optional)\n")
		fmt.Fprintf(os.Stderr, "  --testnet       Use Binance testnet (default: false)\n")

		fmt.Fprintf(os.Stderr, `
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DATA COLLECTOR OPTIONS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`)
		fmt.Fprintf(os.Stderr, "  --collect       Enable collector mode\n")
		fmt.Fprintf(os.Stderr, "  --symbol        Trading pair symbol (default: BTCUSDT)\n")
		fmt.Fprintf(os.Stderr, "  --interval      Candle interval: 1m, 5m, 15m, 1h, 4h, 1d (default: 1h)\n")
		fmt.Fprintf(os.Stderr, "  --days          Number of days to collect from today (default: 30)\n")
		fmt.Fprintf(os.Stderr, "  --start         Start date in YYYY-MM-DD format\n")
		fmt.Fprintf(os.Stderr, "  --end           End date in YYYY-MM-DD format\n")

		fmt.Fprintf(os.Stderr, `
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EXAMPLES:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  # Start API server on default port 8080
  ./server

  # Start API server on custom port
  ./server --port 3000

  # Collect last 7 days of BTC 5-minute candles
  ./server --collect --symbol BTCUSDT --interval 5m --days 7

  # Collect specific date range
  ./server --collect --symbol ETHUSDT --interval 1h --start 2024-01-01 --end 2024-01-31

  # Collect with custom database
  ./server --collect --db ./custom.db --symbol BNBUSDT --days 30

`)
	}

	// Parse command line flags
	port := flag.String("port", "8080", "API server port")
	dbPath := flag.String("db", "data/trading.db", "Path to SQLite database file")
	apiKey := flag.String("api-key", "", "Binance API key (optional for public data)")
	secretKey := flag.String("secret-key", "", "Binance secret key (optional for public data)")
	useTestnet := flag.Bool("testnet", false, "Use Binance testnet")

	// Collector flags
	collect := flag.Bool("collect", false, "Run data collector instead of API server")
	symbol := flag.String("symbol", "BTCUSDT", "Trading pair symbol for collector (e.g., BTCUSDT)")
	interval := flag.String("interval", "1h", "Candle interval for collector (e.g., 1m, 5m, 1h, 1d)")
	days := flag.Int("days", 0, "Number of days to collect (from today backwards)")
	startDate := flag.String("start", "", "Start date for collector (YYYY-MM-DD format)")
	endDate := flag.String("end", "", "End date for collector (YYYY-MM-DD format)")

	flag.Parse()

	// If collect flag is set, run collector instead of API server
	if *collect {
		runCollector(*dbPath, *symbol, *interval, *days, *startDate, *endDate)
		return
	}

	log.Printf("=== Crypto Quant API Server ===")
	log.Printf("Port: %s", *port)
	log.Printf("Database: %s", *dbPath)
	log.Printf("Testnet: %t", *useTestnet)

	// Initialize database
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	candleRepo := history.NewCandleRepository(db)
	tradeRepo := history.NewTradeRepository(db)

	// Initialize Binance client
	var binanceExchange *exchange.BinanceExchange
	if *apiKey != "" && *secretKey != "" {
		binanceExchange, err = exchange.NewBinanceExchange(*apiKey, *secretKey, *useTestnet)
		if err != nil {
			log.Fatalf("Failed to initialize Binance exchange: %v", err)
		}
	} else {
		// Use public API (no authentication)
		if *useTestnet {
			binance.UseTestnet = true
		}
		binanceExchange = &exchange.BinanceExchange{}
		binanceExchange.SetClient(binance.NewClient("", ""))
	}

	// Initialize Binance REST client for data collection
	var binanceClient *binance.Client
	if *useTestnet {
		binance.UseTestnet = true
	}
	binanceClient = binance.NewClient(*apiKey, *secretKey)

	// Initialize wallet (virtual trading)
	initialBalances := map[string]float64{
		"USDT": 10000.0, // Start with $10,000 USDT
	}
	walletManager := wallet.NewManager(initialBalances)

	// Initialize portfolio (position tracking)
	portfolioManager := portfolio.NewManager()

	// Initialize services
	marketService := price.NewService(binanceExchange)
	dataService := history.NewService(candleRepo, tradeRepo, binanceClient)
	walletService := wallet.NewService(walletManager)
	portfolioService := portfolio.NewService(portfolioManager, binanceExchange)

	// Initialize handlers
	marketHandler := handler.NewMarketHandler(marketService)
	dataHandler := handler.NewDataHandler(dataService)
	walletHandler := handler.NewWalletHandler(walletService)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)
	backtestHandler := handler.NewBacktestHandler(dataService)

	// Setup router
	r := api.SetupRouter(marketHandler, dataHandler, walletHandler, portfolioHandler, backtestHandler)

	// Start server
	log.Printf("API server starting on port %s", *port)
	log.Printf("--------------------------------")
	log.Printf("Frontend: http://localhost:%s", *port)
	log.Printf("Health check: http://localhost:%s/health", *port)
	log.Printf("Swagger docs: http://localhost:%s/swagger/index.html", *port)
	log.Printf("API base URL: http://localhost:%s/api/v1", *port)
	log.Printf("--------------------------------")

	// Graceful shutdown
	go func() {
		if err := r.Run(":" + *port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}

// runCollector runs the data collector functionality
func runCollector(dbPath, symbol, interval string, days int, startDate, endDate string) {
	log.Printf("=== Historical Data Collector ===")
	log.Printf("Symbol: %s", symbol)
	log.Printf("Interval: %s", interval)
	log.Printf("Database: %s", dbPath)

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Binance client (no API key needed for public data)
	client := binance.NewClient("", "")

	// Initialize collector
	candleRepo := history.NewCandleRepository(db)
	col := history.NewCollector(client, candleRepo)

	// Calculate time range
	var startTime, endTime time.Time

	// Priority: start/end dates > days
	if startDate != "" && endDate != "" {
		// Use start/end dates
		startTime, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			log.Fatalf("Invalid start date format. Use YYYY-MM-DD: %v", err)
		}
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)

		endTime, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			log.Fatalf("Invalid end date format. Use YYYY-MM-DD: %v", err)
		}
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, time.UTC)

		log.Printf("Start Date: %s", startDate)
		log.Printf("End Date: %s", endDate)
	} else if days > 0 {
		// Use days
		endTime = time.Now()
		startTime = endTime.AddDate(0, 0, -days)
		log.Printf("Days: %d", days)
	} else {
		// Default: 30 days
		endTime = time.Now()
		startTime = endTime.AddDate(0, 0, -30)
		log.Printf("Days: 30 (default)")
	}

	// Collect historical data
	ctx := context.Background()
	if err := col.CollectHistorical(ctx, symbol, interval, startTime, endTime); err != nil {
		log.Fatalf("Failed to collect historical data: %v", err)
	}

	log.Println("Historical data collection completed successfully!")
}
