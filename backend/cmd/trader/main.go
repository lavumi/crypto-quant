package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lavumi/crypto-quant/internal/datasource/exchange"
	"github.com/lavumi/crypto-quant/internal/domain"
	"github.com/lavumi/crypto-quant/pkg/config"
)

func main() {
	fmt.Println("=== Crypto Quant Trading System ===")
	log.Println("Starting trader application...")

	// Load configuration
	cfg, err := config.LoadOrDefault("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Configuration loaded: %s exchange", cfg.Exchange.Type)

	// Initialize exchange
	var ex domain.Exchange
	var err2 error

	switch cfg.Exchange.Type {
	case "virtual":
		ex = exchange.NewVirtualExchange(cfg.Exchange.InitialPrices)
		log.Println("Virtual exchange initialized")
	case "binance":
		// API keys are optional for public data (price queries)
		// Only required for private operations (trading, balance checks)
		ex, err2 = exchange.NewBinanceExchange(
			cfg.Exchange.Binance.APIKey,
			cfg.Exchange.Binance.SecretKey,
			cfg.Exchange.Binance.UseTestnet,
		)
		if err2 != nil {
			log.Fatalf("Failed to initialize Binance exchange: %v", err2)
		}
		if cfg.Exchange.Binance.UseTestnet {
			log.Println("Binance TESTNET exchange initialized")
		} else {
			log.Println("Binance exchange initialized (public data access)")
		}
	default:
		log.Fatalf("Unsupported exchange type: %s", cfg.Exchange.Type)
	}
	defer ex.Close()

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start periodic price display
	ticker := time.NewTicker(time.Duration(cfg.Trading.UpdateIntervalSec) * time.Second)
	defer ticker.Stop()

	log.Println("\nSystem running. Press Ctrl+C to exit.")
	log.Printf("Monitoring %d symbols: %v\n", len(cfg.Trading.Symbols), cfg.Trading.Symbols)

	// Display prices immediately
	displayPrices(ctx, ex, cfg.Trading.Symbols)

	for {
		select {
		case <-ticker.C:
			displayPrices(ctx, ex, cfg.Trading.Symbols)
		case <-sigCh:
			log.Println("\nShutdown signal received. Closing...")
			cancel()
			time.Sleep(500 * time.Millisecond)
			fmt.Println("\nThank you for using Crypto Quant Trading System!")
			return
		}
	}
}

func displayPrices(ctx context.Context, ex domain.Exchange, symbols []string) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Crypto Prices - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))

	type priceInfo struct {
		symbol string
		price  float64
		err    error
	}

	results := make(chan priceInfo, len(symbols))

	// Fetch all prices concurrently
	for _, symbol := range symbols {
		go func(sym string) {
			price, err := ex.GetCurrentPrice(ctx, sym)
			results <- priceInfo{symbol: sym, price: price, err: err}
		}(symbol)
	}

	// Collect and display results
	prices := make(map[string]float64)
	for i := 0; i < len(symbols); i++ {
		result := <-results
		if result.err != nil {
			log.Printf("Failed to get price for %s: %v", result.symbol, result.err)
		} else {
			prices[result.symbol] = result.price
		}
	}

	// Display in order
	for _, symbol := range symbols {
		if price, ok := prices[symbol]; ok {
			// Extract coin name (e.g., BTC from BTCUSDT)
			coinName := symbol[:len(symbol)-4]
			fmt.Printf("  %-10s $%12.2f\n", coinName, price)
		}
	}

	fmt.Println(strings.Repeat("=", 80))
}
