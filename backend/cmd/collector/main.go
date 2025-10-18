package main

import (
	"context"
	"flag"
	"log"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/datasource/database"
	"github.com/lavumi/crypto-quant/internal/datasource/market/history"
)

func main() {
	// Parse command line flags
	symbol := flag.String("symbol", "BTCUSDT", "Trading pair symbol (e.g., BTCUSDT)")
	interval := flag.String("interval", "1h", "Candle interval (e.g., 1m, 5m, 1h, 1d)")
	days := flag.Int("days", 30, "Number of days to collect (from today backwards)")
	dbPath := flag.String("db", "data/trading.db", "Path to SQLite database file")

	flag.Parse()

	log.Printf("=== Historical Data Collector ===")
	log.Printf("Symbol: %s", *symbol)
	log.Printf("Interval: %s", *interval)
	log.Printf("Days: %d", *days)
	log.Printf("Database: %s", *dbPath)

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

	// Initialize Binance client (no API key needed for public data)
	client := binance.NewClient("", "")

	// Initialize collector
	candleRepo := history.NewCandleRepository(db)
	col := history.NewCollector(client, candleRepo)

	// Calculate time range
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -*days)

	// Collect historical data
	ctx := context.Background()
	if err := col.CollectHistorical(ctx, *symbol, *interval, startTime, endTime); err != nil {
		log.Fatalf("Failed to collect historical data: %v", err)
	}

	log.Println("Historical data collection completed successfully!")
}
