package main

import (
	"context"
	"flag"
	"log"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/datasource/database"
	"github.com/lavumi/crypto-quant/internal/datasource/market/history"
	"github.com/lavumi/crypto-quant/internal/quant/strategy"
	"github.com/lavumi/crypto-quant/pkg/config"
)

func main() {
	// Command line flags
	symbol := flag.String("symbol", "BTCUSDT", "Trading symbol")
	interval := flag.String("interval", "1h", "Candle interval (1m, 5m, 15m, 1h, 4h, 1d)")
	startDate := flag.String("start", "", "Start date (YYYY-MM-DD)")
	endDate := flag.String("end", "", "End date (YYYY-MM-DD)")
	balance := flag.Float64("balance", 10000.0, "Initial balance")
	commission := flag.Float64("commission", 0.001, "Commission rate (default: 0.1%)")

	// Strategy parameters
	fastMA := flag.Int("fast", 10, "Fast MA period")
	slowMA := flag.Int("slow", 30, "Slow MA period")

	flag.Parse()

	// Parse dates
	var startTime, endTime time.Time
	var err error

	if *startDate != "" {
		startTime, err = time.Parse("2006-01-02", *startDate)
		if err != nil {
			log.Fatalf("Invalid start date: %v", err)
		}
	} else {
		// Default: 3 months ago
		startTime = time.Now().AddDate(0, -3, 0)
	}

	if *endDate != "" {
		endTime, err = time.Parse("2006-01-02", *endDate)
		if err != nil {
			log.Fatalf("Invalid end date: %v", err)
		}
	} else {
		// Default: now
		endTime = time.Now()
	}

	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New("data/trading.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Binance client
	binanceClient := binance.NewClient(cfg.Exchange.Binance.APIKey, cfg.Exchange.Binance.SecretKey)

	// Initialize repositories and services
	candleRepo := history.NewCandleRepository(db)
	tradeRepo := history.NewTradeRepository(db)
	historyService := history.NewService(candleRepo, tradeRepo, binanceClient)

	ctx := context.Background()

	// Check if historical data exists, if not, collect it
	log.Printf("Checking for historical data...")
	latestCandle, err := historyService.GetLatestCandle(ctx, *symbol, *interval)
	if err != nil || latestCandle == nil || latestCandle.OpenTime.Before(startTime) {
		log.Printf("Collecting historical data for %s (%s) from %s to %s",
			*symbol, *interval, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

		if err := historyService.CollectHistoricalData(ctx, *symbol, *interval, startTime, endTime); err != nil {
			log.Fatalf("Failed to collect historical data: %v", err)
		}
	}

	// Load candles for backtesting
	log.Printf("Loading candles for backtesting...")
	candles, err := historyService.GetCandles(ctx, *symbol, *interval, startTime, endTime)
	if err != nil {
		log.Fatalf("Failed to load candles: %v", err)
	}

	if len(candles) == 0 {
		log.Fatalf("No candles found for the specified period")
	}

	log.Printf("Loaded %d candles from %s to %s",
		len(candles),
		candles[0].OpenTime.Format("2006-01-02"),
		candles[len(candles)-1].OpenTime.Format("2006-01-02"))

	// Create strategy
	strat := strategy.NewMACrossStrategy(*fastMA, *slowMA)

	// Create and run backtest engine
	engine := backtest.NewEngine(&backtest.Config{
		InitialBalance: *balance,
		Commission:     *commission,
		Strategy:       strat,
	})

	log.Printf("Running backtest with strategy: %s", strat.Name())
	result, err := engine.Run(ctx, candles)
	if err != nil {
		log.Fatalf("Backtest failed: %v", err)
	}

	// Print results
	result.Print()

	// Print some trades
	if len(result.Trades) > 0 {
		log.Printf("\nFirst 10 trades:")
		for i, trade := range result.Trades {
			if i >= 10 {
				break
			}
			log.Printf("  %s: %s %.8f @ %.2f (Fee: %.2f) - %s",
				trade.Timestamp.Format("2006-01-02 15:04"),
				trade.Side,
				trade.Quantity,
				trade.Price,
				trade.Fee,
				trade.Reason,
			)
		}

		if len(result.Trades) > 10 {
			log.Printf("  ... and %d more trades", len(result.Trades)-10)
		}
	}
}
