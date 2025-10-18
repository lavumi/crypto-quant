package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/api"
	"github.com/lavumi/crypto-quant/internal/api/handler"
	"github.com/lavumi/crypto-quant/internal/datasource/database"
	"github.com/lavumi/crypto-quant/internal/datasource/exchange"
	"github.com/lavumi/crypto-quant/internal/datasource/market/history"
	"github.com/lavumi/crypto-quant/internal/datasource/market/price"
	"github.com/lavumi/crypto-quant/internal/trading"
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
	// Parse command line flags
	port := flag.String("port", "8080", "API server port")
	dbPath := flag.String("db", "data/trading.db", "Path to SQLite database file")
	apiKey := flag.String("api-key", "", "Binance API key (optional for public data)")
	secretKey := flag.String("secret-key", "", "Binance secret key (optional for public data)")
	useTestnet := flag.Bool("testnet", false, "Use Binance testnet")

	flag.Parse()

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
	orderService := order.NewService(walletManager, portfolioManager, binanceExchange)

	// Initialize handlers
	marketHandler := handler.NewMarketHandler(marketService)
	dataHandler := handler.NewDataHandler(dataService)
	walletHandler := handler.NewWalletHandler(walletService)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)
	orderHandler := handler.NewOrderHandler(orderService)
	backtestHandler := handler.NewBacktestHandler(dataService)

	// Setup router
	r := api.SetupRouter(marketHandler, dataHandler, walletHandler, portfolioHandler, orderHandler, backtestHandler)

	// Start server
	log.Printf("API server starting on port %s", *port)
	log.Printf("Health check: http://localhost:%s/health", *port)
	log.Printf("Swagger docs: http://localhost:%s/swagger/index.html", *port)
	log.Printf("API base URL: http://localhost:%s/api/v1", *port)

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
