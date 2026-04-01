package main

import (
	"context"
	"flag"
	"log"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/datasource/database"
	binanceExchange "github.com/lavumi/crypto-quant/internal/exchange/binance"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/strategy"
	"github.com/lavumi/crypto-quant/internal/repository"
	"github.com/lavumi/crypto-quant/internal/service/market"
	"github.com/lavumi/crypto-quant/pkg/config"
)

func main() {
	// Command line flags
	strategyName := flag.String("strategy", strategy.NameGoldenRSIBB, "Strategy to run")
	symbol := flag.String("symbol", "BTCUSDT", "Trading symbol")
	interval := flag.String("interval", "1h", "Candle interval (1m, 5m, 15m, 1h, 4h, 1d)")
	startDate := flag.String("start", "", "Start date (YYYY-MM-DD)")
	endDate := flag.String("end", "", "End date (YYYY-MM-DD)")
	balance := flag.Float64("balance", 10000.0, "Initial balance")
	commission := flag.Float64("commission", 0.001, "Commission rate (default: 0.1%)")

	// Strategy parameters (GoldenRSIBB)
	fastMA := flag.Int("fast", 5, "Fast MA period (e.g., 5)")
	slowMA := flag.Int("slow", 20, "Slow MA period (e.g., 20)")
	rsiPeriod := flag.Int("rsi", 14, "RSI period")
	rsiLower := flag.Float64("rsi-lower", 40, "RSI lower bound")
	rsiUpper := flag.Float64("rsi-upper", 70, "RSI upper bound")
	bbPeriod := flag.Int("bb", 20, "Bollinger Bands period")
	bbMult := flag.Float64("bb-mult", 2.0, "Bollinger Bands multiplier")
	dcaPeriod := flag.String("dca-period", "24h", "DCA purchase interval")
	dcaAmount := flag.Float64("dca-amount", 100, "DCA purchase amount in USDT")
	volThresh := flag.Float64("vol-threshold", 1.3, "Volume spike threshold (x average)")
	tp := flag.Float64("tp", 0.06, "Take profit percent (e.g., 0.06 = 6%)")
	sl := flag.Float64("sl", 0.03, "Stop loss percent (e.g., 0.03 = 3%)")
	position := flag.Float64("position", 1.0, "Position size as fraction of balance (0.0-1.0)")

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

	// Initialize repositories
	candleRepo := repository.NewCandleRepository(db)
	tradeRepo := repository.NewTradeRepository(db)

	// Initialize collector and exchange
	collector := binanceExchange.NewCollector(binanceClient, candleRepo)

	// Initialize service
	historyService := market.NewHistoryService(candleRepo, tradeRepo, collector)

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

	strategyConfig := strategy.Config{
		Name:                  *strategyName,
		FastPeriod:            *fastMA,
		SlowPeriod:            *slowMA,
		RSIPeriod:             *rsiPeriod,
		RSIOversold:           *rsiLower,
		RSIOverbought:         *rsiUpper,
		BBPeriod:              *bbPeriod,
		BBMultiplier:          *bbMult,
		DCAPeriod:             *dcaPeriod,
		DCAAmountUSDT:         *dcaAmount,
		GoldenFastPeriod:      *fastMA,
		GoldenSlowPeriod:      *slowMA,
		GoldenRSIPeriod:       *rsiPeriod,
		GoldenRSILowerBound:   *rsiLower,
		GoldenRSIUpperBound:   *rsiUpper,
		GoldenBBPeriod:        *bbPeriod,
		GoldenBBMultiplier:    *bbMult,
		GoldenVolumeThreshold: *volThresh,
		GoldenTakeProfitPct:   *tp,
		GoldenStopLossPct:     *sl,
		PositionSize:          *position,
	}

	strat, configJSON, err := strategy.Build(strategyConfig)
	if err != nil {
		log.Fatalf("Failed to build strategy: %v", err)
	}

	// Create and run backtest engine
	engine := backtest.NewEngine(&backtest.Config{
		InitialBalance: *balance,
		Commission:     *commission,
		Strategy:       strat,
		Persist:        true,
		DB:             db,
		Symbol:         *symbol,
		Interval:       *interval,
		ConfigJSON:     configJSON,
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

	// Persistence is handled by engine when configured
}
