package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/datasource/database"
	binanceExchange "github.com/lavumi/crypto-quant/internal/exchange/binance"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/optimize"
	"github.com/lavumi/crypto-quant/internal/quant/strategy"
	"github.com/lavumi/crypto-quant/internal/repository"
	"github.com/lavumi/crypto-quant/internal/service/market"
	"github.com/lavumi/crypto-quant/pkg/config"
)

type stringListFlag []string

func (s *stringListFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringListFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	// Command line flags
	mode := flag.String("mode", "single", "Execution mode: single, sweep, or walk-forward")
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
	sweepSort := flag.String("sort", "sharpe", "Sweep sort metric: sharpe, return, calmar, profit_factor, win_rate, mdd")
	sweepTop := flag.Int("top", 10, "Number of top sweep results to print")
	trainDays := flag.Int("train-days", 180, "Walk-forward training window in days")
	testDays := flag.Int("test-days", 60, "Walk-forward test window in days")
	stepDays := flag.Int("step-days", 60, "Walk-forward step size in days")
	maxWindows := flag.Int("max-windows", 0, "Maximum walk-forward windows to evaluate (0 = all)")
	var sweepParams stringListFlag
	flag.Var(&sweepParams, "param", "Sweep parameter grid in the form key=v1,v2,v3 (repeatable)")

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

	if *mode == "sweep" {
		grid, err := parseSweepParams(sweepParams)
		if err != nil {
			log.Fatalf("Failed to parse sweep params: %v", err)
		}
		if len(grid) == 0 {
			log.Fatalf("Sweep mode requires at least one --param key=v1,v2 argument")
		}

		results, err := optimize.RunSweep(ctx, candles, optimize.RunnerConfig{
			InitialBalance: *balance,
			Commission:     *commission,
			Persist:        false,
			Symbol:         *symbol,
			Interval:       *interval,
		}, optimize.SweepSpec{
			BaseConfig: strategyConfig,
			Parameters: grid,
			SortBy:     *sweepSort,
			Top:        *sweepTop,
		})
		if err != nil {
			log.Fatalf("Sweep failed: %v", err)
		}

		printSweepResults(results, *sweepSort, *sweepTop, grid)
		return
	}

	if *mode == "walk-forward" {
		grid, err := parseSweepParams(sweepParams)
		if err != nil {
			log.Fatalf("Failed to parse sweep params: %v", err)
		}
		if len(grid) == 0 {
			log.Fatalf("Walk-forward mode requires at least one --param key=v1,v2 argument")
		}

		summary, err := optimize.RunWalkForward(ctx, candles, optimize.RunnerConfig{
			InitialBalance: *balance,
			Commission:     *commission,
			Persist:        false,
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
		})
		if err != nil {
			log.Fatalf("Walk-forward failed: %v", err)
		}

		printWalkForwardSummary(summary, *sweepTop, grid)
		return
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
