package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/api"
	"github.com/lavumi/crypto-quant/internal/api/handler"
	"github.com/lavumi/crypto-quant/internal/datasource/database"
	binanceExchange "github.com/lavumi/crypto-quant/internal/exchange/binance"
	"github.com/lavumi/crypto-quant/internal/portfolio"
	"github.com/lavumi/crypto-quant/internal/portfolio/wallet"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/optimize"
	"github.com/lavumi/crypto-quant/internal/quant/strategy"
	"github.com/lavumi/crypto-quant/internal/repository"
	"github.com/lavumi/crypto-quant/internal/service/market"
	"github.com/lavumi/crypto-quant/pkg/config"

	_ "github.com/lavumi/crypto-quant/docs"
)

const defaultDBPath = "data/trading.db"

type stringListFlag []string

func (s *stringListFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringListFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "api":
			runAPISubcommand(os.Args[2:])
			return
		case "collect":
			runCollectSubcommand(os.Args[2:])
			return
		case "backtest":
			runBacktestSubcommand("single", os.Args[2:])
			return
		case "sweep":
			runBacktestSubcommand("sweep", os.Args[2:])
			return
		case "walk-forward":
			runBacktestSubcommand("walk-forward", os.Args[2:])
			return
		case "help", "--help", "-h":
			printRootUsage()
			return
		}
	}

	runLegacyMode(os.Args[1:])
}

func printRootUsage() {
	fmt.Fprintf(os.Stderr, `
╔══════════════════════════════════════════════════════════════╗
║          Crypto Quant - Trading Analysis Platform            ║
╚══════════════════════════════════════════════════════════════╝

USAGE:
  ./server <command> [options]
  ./server [legacy-options]

COMMANDS:
  api            Start API server with web interface
  collect        Collect historical candle data
  backtest       Run a single backtest and persist the result
  sweep          Run a parameter sweep and persist the result
  walk-forward   Run walk-forward validation and persist the result

LEGACY MODE:
  ./server                Start API server
  ./server --collect ...  Run collector

EXAMPLES:
  ./server api --port 8080 --db data/trading.db
  ./server collect --symbol BTCUSDT --interval 1h --days 365 --db data/trading.db
  ./server backtest --strategy golden_rsi_bb --symbol BTCUSDT --start 2025-01-01 --end 2025-12-31 --db data/trading.db
  ./server sweep --strategy golden_rsi_bb --param golden_fast=5,10 --param golden_slow=20,30 --param tp=0.04,0.06 --param sl=0.02,0.03 --db data/trading.db
  ./server walk-forward --strategy golden_rsi_bb --param golden_fast=5,10 --param golden_slow=20,30 --param tp=0.04,0.06 --param sl=0.02,0.03 --train-days 180 --test-days 60 --step-days 60 --db data/trading.db
`)
}

func runLegacyMode(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	fs.Usage = printRootUsage

	port := fs.String("port", "8080", "API server port")
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	apiKey := fs.String("api-key", "", "Binance API key (optional for public data)")
	secretKey := fs.String("secret-key", "", "Binance secret key (optional for public data)")
	useTestnet := fs.Bool("testnet", false, "Use Binance testnet")

	collect := fs.Bool("collect", false, "Run data collector instead of API server")
	symbol := fs.String("symbol", "BTCUSDT", "Trading pair symbol for collector (e.g., BTCUSDT)")
	interval := fs.String("interval", "1h", "Candle interval for collector (e.g., 1m, 5m, 1h, 1d)")
	days := fs.Int("days", 0, "Number of days to collect (from today backwards)")
	startDate := fs.String("start", "", "Start date for collector (YYYY-MM-DD format)")
	endDate := fs.String("end", "", "End date for collector (YYYY-MM-DD format)")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	if *collect {
		runCollector(*dbPath, *symbol, *interval, *days, *startDate, *endDate)
		return
	}

	runAPIServer(*port, *dbPath, *apiKey, *secretKey, *useTestnet)
}

func runAPISubcommand(args []string) {
	fs := flag.NewFlagSet("api", flag.ExitOnError)
	port := fs.String("port", "8080", "API server port")
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	apiKey := fs.String("api-key", "", "Binance API key (optional for public data)")
	secretKey := fs.String("secret-key", "", "Binance secret key (optional for public data)")
	useTestnet := fs.Bool("testnet", false, "Use Binance testnet")
	if err := fs.Parse(args); err != nil {
		log.Fatalf("Failed to parse api flags: %v", err)
	}

	runAPIServer(*port, *dbPath, *apiKey, *secretKey, *useTestnet)
}

func runCollectSubcommand(args []string) {
	fs := flag.NewFlagSet("collect", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	symbol := fs.String("symbol", "BTCUSDT", "Trading pair symbol")
	interval := fs.String("interval", "1h", "Candle interval")
	days := fs.Int("days", 0, "Number of days to collect from today backwards")
	startDate := fs.String("start", "", "Start date for collector (YYYY-MM-DD format)")
	endDate := fs.String("end", "", "End date for collector (YYYY-MM-DD format)")
	if err := fs.Parse(args); err != nil {
		log.Fatalf("Failed to parse collect flags: %v", err)
	}

	runCollector(*dbPath, *symbol, *interval, *days, *startDate, *endDate)
}

func runBacktestSubcommand(mode string, args []string) {
	fs := flag.NewFlagSet(mode, flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	strategyName := fs.String("strategy", strategy.NameGoldenRSIBB, "Strategy to run")
	symbol := fs.String("symbol", "BTCUSDT", "Trading symbol")
	interval := fs.String("interval", "1h", "Candle interval (1m, 5m, 15m, 1h, 4h, 1d)")
	startDate := fs.String("start", "", "Start date (YYYY-MM-DD)")
	endDate := fs.String("end", "", "End date (YYYY-MM-DD)")
	balance := fs.Float64("balance", 10000.0, "Initial balance")
	commission := fs.Float64("commission", 0.001, "Commission rate (default: 0.1%)")
	fastMA := fs.Int("fast", 5, "Fast MA period")
	slowMA := fs.Int("slow", 20, "Slow MA period")
	rsiPeriod := fs.Int("rsi", 14, "RSI period")
	rsiLower := fs.Float64("rsi-lower", 40, "RSI lower bound")
	rsiUpper := fs.Float64("rsi-upper", 70, "RSI upper bound")
	bbPeriod := fs.Int("bb", 20, "Bollinger Bands period")
	bbMult := fs.Float64("bb-mult", 2.0, "Bollinger Bands multiplier")
	dcaPeriod := fs.String("dca-period", "24h", "DCA purchase interval")
	dcaAmount := fs.Float64("dca-amount", 100, "DCA purchase amount in USDT")
	volThresh := fs.Float64("vol-threshold", 1.3, "Volume spike threshold")
	tp := fs.Float64("tp", 0.06, "Take profit percent")
	sl := fs.Float64("sl", 0.03, "Stop loss percent")
	position := fs.Float64("position", 1.0, "Position size as fraction of balance")
	sweepSort := fs.String("sort", "sharpe", "Sort metric: sharpe, return, calmar, profit_factor, win_rate, mdd")
	sweepTop := fs.Int("top", 10, "Number of top results to print")
	trainDays := fs.Int("train-days", 180, "Walk-forward training window in days")
	testDays := fs.Int("test-days", 60, "Walk-forward test window in days")
	stepDays := fs.Int("step-days", 60, "Walk-forward step size in days")
	maxWindows := fs.Int("max-windows", 0, "Maximum walk-forward windows to evaluate (0 = all)")
	var sweepParams stringListFlag
	fs.Var(&sweepParams, "param", "Sweep parameter grid in the form key=v1,v2,v3 (repeatable)")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Failed to parse %s flags: %v", mode, err)
	}

	var startTime, endTime time.Time
	var err error
	if *startDate != "" {
		startTime, err = time.Parse("2006-01-02", *startDate)
		if err != nil {
			log.Fatalf("Invalid start date: %v", err)
		}
	} else {
		startTime = time.Now().AddDate(0, -3, 0)
	}

	if *endDate != "" {
		endTime, err = time.Parse("2006-01-02", *endDate)
		if err != nil {
			log.Fatalf("Invalid end date: %v", err)
		}
	} else {
		endTime = time.Now()
	}

	cfg, err := config.Load(resolveConfigPath())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	binanceClient := binance.NewClient(cfg.Exchange.Binance.APIKey, cfg.Exchange.Binance.SecretKey)
	candleRepo := repository.NewCandleRepository(db)
	tradeRepo := repository.NewTradeRepository(db)
	collector := binanceExchange.NewCollector(binanceClient, candleRepo)
	historyService := market.NewHistoryService(candleRepo, tradeRepo, collector)

	ctx := context.Background()

	log.Printf("Checking for historical data...")
	latestCandle, err := historyService.GetLatestCandle(ctx, *symbol, *interval)
	if err != nil || latestCandle == nil || latestCandle.OpenTime.Before(startTime) {
		log.Printf("Collecting historical data for %s (%s) from %s to %s",
			*symbol, *interval, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
		if err := historyService.CollectHistoricalData(ctx, *symbol, *interval, startTime, endTime); err != nil {
			log.Fatalf("Failed to collect historical data: %v", err)
		}
	}

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

	runnerCfg := optimize.RunnerConfig{
		InitialBalance: *balance,
		Commission:     *commission,
		Persist:        false,
		Symbol:         *symbol,
		Interval:       *interval,
	}

	switch mode {
	case "sweep":
		grid, err := parseSweepParams(sweepParams)
		if err != nil {
			log.Fatalf("Failed to parse sweep params: %v", err)
		}
		if len(grid) == 0 {
			log.Fatalf("Sweep mode requires at least one --param key=v1,v2 argument")
		}

		results, err := optimize.RunSweep(ctx, candles, runnerCfg, optimize.SweepSpec{
			BaseConfig: strategyConfig,
			Parameters: grid,
			SortBy:     *sweepSort,
			Top:        *sweepTop,
		})
		if err != nil {
			log.Fatalf("Sweep failed: %v", err)
		}

		sweepID, err := optimize.PersistSweep(ctx, db, optimize.RunnerConfig{
			InitialBalance: *balance,
			Commission:     *commission,
			Symbol:         *symbol,
			Interval:       *interval,
		}, optimize.SweepSpec{
			BaseConfig: strategyConfig,
			Parameters: grid,
			SortBy:     *sweepSort,
		}, results)
		if err != nil {
			log.Printf("Failed to persist sweep results: %v", err)
		} else {
			log.Printf("Persisted sweep run: id=%d", sweepID)
		}

		printSweepResults(results, *sweepSort, *sweepTop, grid)
		return
	case "walk-forward":
		grid, err := parseSweepParams(sweepParams)
		if err != nil {
			log.Fatalf("Failed to parse sweep params: %v", err)
		}
		if len(grid) == 0 {
			log.Fatalf("Walk-forward mode requires at least one --param key=v1,v2 argument")
		}

		summary, err := optimize.RunWalkForward(ctx, candles, runnerCfg, optimize.WalkForwardSpec{
			SweepSpec: optimize.SweepSpec{
				BaseConfig: strategyConfig,
				Parameters: grid,
				SortBy:     *sweepSort,
			},
			TrainDuration:   time.Duration(*trainDays) * 24 * time.Hour,
			TestDuration:    time.Duration(*testDays) * 24 * time.Hour,
			StepDuration:    time.Duration(*stepDays) * 24 * time.Hour,
			MaxWindows:      *maxWindows,
			SelectionMetric: *sweepSort,
		})
		if err != nil {
			log.Fatalf("Walk-forward failed: %v", err)
		}

		walkForwardID, err := optimize.PersistWalkForward(ctx, db, optimize.RunnerConfig{
			InitialBalance: *balance,
			Commission:     *commission,
			Symbol:         *symbol,
			Interval:       *interval,
		}, optimize.WalkForwardSpec{
			SweepSpec: optimize.SweepSpec{
				BaseConfig: strategyConfig,
				Parameters: grid,
				SortBy:     *sweepSort,
			},
			TrainDuration:   time.Duration(*trainDays) * 24 * time.Hour,
			TestDuration:    time.Duration(*testDays) * 24 * time.Hour,
			StepDuration:    time.Duration(*stepDays) * 24 * time.Hour,
			MaxWindows:      *maxWindows,
			SelectionMetric: *sweepSort,
		}, summary)
		if err != nil {
			log.Printf("Failed to persist walk-forward results: %v", err)
		} else {
			log.Printf("Persisted walk-forward run: id=%d", walkForwardID)
		}

		gridForSummary, _ := parseSweepParams(sweepParams)
		printWalkForwardSummary(summary, *sweepTop, gridForSummary)
		return
	default:
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

		result.Print()

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
}

func runAPIServer(port, dbPath, apiKey, secretKey string, useTestnet bool) {
	log.Printf("=== Crypto Quant API Server ===")
	log.Printf("Port: %s", port)
	log.Printf("Database: %s", dbPath)
	log.Printf("Testnet: %t", useTestnet)

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	candleRepo := repository.NewCandleRepository(db)
	tradeRepo := repository.NewTradeRepository(db)

	var exchange *binanceExchange.Exchange
	if apiKey != "" && secretKey != "" {
		exchange, err = binanceExchange.New(apiKey, secretKey, useTestnet)
		if err != nil {
			log.Fatalf("Failed to initialize Binance exchange: %v", err)
		}
	} else {
		if useTestnet {
			binance.UseTestnet = true
		}
		exchange, _ = binanceExchange.New("", "", useTestnet)
		exchange.SetClient(binance.NewClient("", ""))
	}

	exchange.SetCollector(candleRepo)

	initialBalances := map[string]float64{
		"USDT": 10000.0,
	}
	walletManager := wallet.NewManager(initialBalances)
	portfolioManager := portfolio.NewManager()

	priceService := market.NewPriceService(exchange)
	historyService := market.NewHistoryService(candleRepo, tradeRepo, exchange)
	walletService := wallet.NewService(walletManager)
	portfolioService := portfolio.NewService(portfolioManager, exchange)

	marketHandler := handler.NewMarketHandler(priceService)
	dataHandler := handler.NewDataHandler(historyService)
	walletHandler := handler.NewWalletHandler(walletService)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)
	backtestHandler := handler.NewBacktestHandler(historyService)
	experimentsHandler := handler.NewExperimentsHandler(db)

	r := api.SetupRouter(marketHandler, dataHandler, walletHandler, portfolioHandler, backtestHandler, experimentsHandler)

	log.Printf("API server starting on port %s", port)
	log.Printf("--------------------------------")
	log.Printf("Frontend: http://localhost:%s", port)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Printf("Swagger docs: http://localhost:%s/swagger/index.html", port)
	log.Printf("API base URL: http://localhost:%s/api/v1", port)
	log.Printf("--------------------------------")

	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}

func runCollector(dbPath, symbol, interval string, days int, startDate, endDate string) {
	log.Printf("=== Historical Data Collector ===")
	log.Printf("Symbol: %s", symbol)
	log.Printf("Interval: %s", interval)
	log.Printf("Database: %s", dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	client := binance.NewClient("", "")
	candleRepo := repository.NewCandleRepository(db)
	col := binanceExchange.NewCollector(client, candleRepo)

	var startTime, endTime time.Time
	if startDate != "" && endDate != "" {
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
		endTime = time.Now()
		startTime = endTime.AddDate(0, 0, -days)
		log.Printf("Days: %d", days)
	} else {
		endTime = time.Now()
		startTime = endTime.AddDate(0, 0, -30)
		log.Printf("Days: 30 (default)")
	}

	ctx := context.Background()
	if err := col.CollectHistorical(ctx, symbol, interval, startTime, endTime); err != nil {
		log.Fatalf("Failed to collect historical data: %v", err)
	}

	log.Println("Historical data collection completed successfully!")
}

func resolveConfigPath() string {
	candidates := []string{
		"backend/configs/config.yaml",
		"configs/config.yaml",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return candidates[0]
}

func parseSweepParams(raw []string) (map[string][]string, error) {
	params := make(map[string][]string)
	for _, item := range raw {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid sweep param %q, expected key=v1,v2", item)
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("invalid empty sweep param key in %q", item)
		}

		values := strings.Split(parts[1], ",")
		clean := make([]string, 0, len(values))
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value != "" {
				clean = append(clean, value)
			}
		}
		if len(clean) == 0 {
			return nil, fmt.Errorf("sweep param %q has no values", item)
		}
		params[key] = clean
	}
	return params, nil
}

func printSweepResults(results []optimize.SweepResult, sortBy string, top int, grid map[string][]string) {
	summary := optimize.Summarize(results, sortBy)
	log.Printf("Sweep complete: %d candidates (%d success, %d failed) ranked by %s",
		summary.TotalCandidates, summary.SuccessfulRuns, summary.FailedRuns, summary.SortBy)

	if summary.Best != nil {
		log.Printf("Best   : %s", formatSweepSummaryLine(*summary.Best))
	}
	if summary.Median != nil {
		log.Printf("Median : %s", formatSweepSummaryLine(*summary.Median))
	}
	if summary.Worst != nil {
		log.Printf("Worst  : %s", formatSweepSummaryLine(*summary.Worst))
	}

	log.Printf("Top %d candidates:", min(top, len(results)))
	log.Printf(" rank | params | return | sharpe | calmar | mdd | pf | trades ")

	for idx, item := range limitResults(results, top) {
		if item.Err != nil {
			log.Printf(" %4d | ERROR | %v", idx+1, item.Err)
			continue
		}

		log.Printf(
			" %4d | %s | %7.2f%% | %6.2f | %6.2f | %6.2f%% | %4.2f | %6d",
			idx+1,
			formatSweepParams(item.Config, grid),
			item.Result.TotalReturn*100,
			item.Result.SharpeRatio,
			item.Result.CalmarRatio,
			item.Result.MaxDrawdownPct*100,
			item.Result.ProfitFactor,
			item.Result.TotalTrades,
		)
	}
}

func formatSweepSummaryLine(item optimize.SweepResult) string {
	if item.Result == nil {
		return "n/a"
	}

	return fmt.Sprintf(
		"return=%.2f%% sharpe=%.2f calmar=%.2f mdd=%.2f%% pf=%.2f trades=%d",
		item.Result.TotalReturn*100,
		item.Result.SharpeRatio,
		item.Result.CalmarRatio,
		item.Result.MaxDrawdownPct*100,
		item.Result.ProfitFactor,
		item.Result.TotalTrades,
	)
}

func formatSweepParams(cfg strategy.Config, grid map[string][]string) string {
	keys := make([]string, 0, len(grid))
	for key := range grid {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		if value, ok := sweepParamValue(cfg, key); ok {
			parts = append(parts, key+"="+value)
		}
	}

	if len(parts) == 0 {
		return cfg.Name
	}
	return strings.Join(parts, ",")
}

func sweepParamValue(cfg strategy.Config, key string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "fast", "fast_period":
		return strconv.Itoa(cfg.FastPeriod), true
	case "slow", "slow_period":
		return strconv.Itoa(cfg.SlowPeriod), true
	case "rsi", "rsi_period":
		return strconv.Itoa(cfg.RSIPeriod), true
	case "rsi_lower", "rsi_oversold":
		return formatSweepFloat(cfg.RSIOversold), true
	case "rsi_upper", "rsi_overbought":
		return formatSweepFloat(cfg.RSIOverbought), true
	case "bb", "bb_period":
		return strconv.Itoa(cfg.BBPeriod), true
	case "bb_mult", "bb_multiplier":
		return formatSweepFloat(cfg.BBMultiplier), true
	case "dca_period":
		return cfg.DCAPeriod, true
	case "dca_amount", "dca_amount_usdt":
		return formatSweepFloat(cfg.DCAAmountUSDT), true
	case "golden_fast", "golden_fast_period":
		return strconv.Itoa(cfg.GoldenFastPeriod), true
	case "golden_slow", "golden_slow_period":
		return strconv.Itoa(cfg.GoldenSlowPeriod), true
	case "golden_rsi", "golden_rsi_period":
		return strconv.Itoa(cfg.GoldenRSIPeriod), true
	case "golden_rsi_lower", "golden_rsi_lower_bound":
		return formatSweepFloat(cfg.GoldenRSILowerBound), true
	case "golden_rsi_upper", "golden_rsi_upper_bound":
		return formatSweepFloat(cfg.GoldenRSIUpperBound), true
	case "golden_bb", "golden_bb_period":
		return strconv.Itoa(cfg.GoldenBBPeriod), true
	case "golden_bb_mult", "golden_bb_multiplier":
		return formatSweepFloat(cfg.GoldenBBMultiplier), true
	case "vol_threshold", "golden_volume_threshold":
		return formatSweepFloat(cfg.GoldenVolumeThreshold), true
	case "tp", "golden_take_profit_pct":
		return formatSweepFloat(cfg.GoldenTakeProfitPct), true
	case "sl", "golden_stop_loss_pct":
		return formatSweepFloat(cfg.GoldenStopLossPct), true
	case "position", "position_size":
		return formatSweepFloat(cfg.PositionSize), true
	default:
		return "", false
	}
}

func formatSweepFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func limitResults(results []optimize.SweepResult, top int) []optimize.SweepResult {
	if top <= 0 || top >= len(results) {
		return results
	}
	return results[:top]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func printWalkForwardSummary(summary *optimize.WalkForwardSummary, top int, grid map[string][]string) {
	if summary == nil {
		log.Printf("Walk-forward summary unavailable")
		return
	}

	log.Printf("Walk-forward complete: %d windows evaluated using %s", summary.Completed, summary.SelectionMetric)
	log.Printf(" win | train range | test range | selected params | train sharpe | test return | test sharpe | test mdd ")

	for idx, window := range limitWalkForward(summary.Windows, top) {
		selectedParams := "n/a"
		trainSharpe := 0.0
		testReturn := 0.0
		testSharpe := 0.0
		testMDD := 0.0

		if window.Selected != nil {
			selectedParams = formatSweepParams(window.Selected.Config, grid)
			if window.Selected.Result != nil {
				trainSharpe = window.Selected.Result.SharpeRatio
			}
		}

		if window.OutOfSample != nil && window.OutOfSample.Result != nil {
			testReturn = window.OutOfSample.Result.TotalReturn * 100
			testSharpe = window.OutOfSample.Result.SharpeRatio
			testMDD = window.OutOfSample.Result.MaxDrawdownPct * 100
		}

		log.Printf(
			" %3d | %s..%s | %s..%s | %s | %11.2f | %10.2f%% | %11.2f | %8.2f%%",
			idx+1,
			window.Window.TrainStart.Format("2006-01-02"),
			window.Window.TrainEnd.Format("2006-01-02"),
			window.Window.TestStart.Format("2006-01-02"),
			window.Window.TestEnd.Format("2006-01-02"),
			selectedParams,
			trainSharpe,
			testReturn,
			testSharpe,
			testMDD,
		)
	}
}

func limitWalkForward(windows []optimize.WalkForwardResult, top int) []optimize.WalkForwardResult {
	if top <= 0 || top >= len(windows) {
		return windows
	}
	return windows[:top]
}
