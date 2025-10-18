package history

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/domain"
)

// Collector collects historical candle data
type Collector struct {
	client     *binance.Client
	candleRepo *CandleRepository
}

// NewCollector creates a new historical data collector
func NewCollector(client *binance.Client, candleRepo *CandleRepository) *Collector {
	return &Collector{
		client:     client,
		candleRepo: candleRepo,
	}
}

// CollectHistorical collects historical candle data for a symbol
func (c *Collector) CollectHistorical(ctx context.Context, symbol, interval string, startTime, endTime time.Time) error {
	log.Printf("Collecting historical data for %s (%s) from %s to %s",
		symbol, interval, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	// Binance allows max 1000 candles per request
	const maxLimit = 1000

	currentStart := startTime
	totalCandles := 0

	for currentStart.Before(endTime) {
		// Calculate end time for this batch
		var limit int
		intervalDuration := parseInterval(interval)
		batchEnd := currentStart.Add(intervalDuration * maxLimit)
		if batchEnd.After(endTime) {
			batchEnd = endTime
			// Calculate actual limit needed
			duration := batchEnd.Sub(currentStart)
			limit = int(duration / intervalDuration)
			if limit > maxLimit {
				limit = maxLimit
			}
			// Ensure limit is at least 1
			if limit < 1 {
				limit = 1
			}
		} else {
			limit = maxLimit
		}

		// Fetch klines from Binance
		klines, err := c.client.NewKlinesService().
			Symbol(symbol).
			Interval(interval).
			StartTime(currentStart.UnixMilli()).
			EndTime(batchEnd.UnixMilli()).
			Limit(limit).
			Do(ctx)

		if err != nil {
			return fmt.Errorf("failed to fetch klines: %w", err)
		}

		if len(klines) == 0 {
			break
		}

		// Convert to candles
		candles := make([]*domain.Candle, 0, len(klines))
		for _, k := range klines {
			open, _ := strconv.ParseFloat(k.Open, 64)
			high, _ := strconv.ParseFloat(k.High, 64)
			low, _ := strconv.ParseFloat(k.Low, 64)
			close, _ := strconv.ParseFloat(k.Close, 64)
			volume, _ := strconv.ParseFloat(k.Volume, 64)

			candle := &domain.Candle{
				Symbol:    symbol,
				OpenTime:  time.Unix(k.OpenTime/1000, 0),
				CloseTime: time.Unix(k.CloseTime/1000, 0),
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    volume,
			}
			candles = append(candles, candle)
		}

		// Save to database
		if err := c.candleRepo.SaveBatch(ctx, candles, interval); err != nil {
			return fmt.Errorf("failed to save candles: %w", err)
		}

		totalCandles += len(candles)
		log.Printf("Saved %d candles (total: %d)", len(candles), totalCandles)

		// Move to next batch
		if len(klines) > 0 {
			lastCandle := klines[len(klines)-1]
			currentStart = time.Unix(lastCandle.CloseTime/1000, 0).Add(1 * time.Millisecond)
		} else {
			break
		}

		// Rate limiting - be nice to Binance API
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Historical data collection complete: %d total candles collected for %s", totalCandles, symbol)
	return nil
}

// parseInterval converts interval string to duration
func parseInterval(interval string) time.Duration {
	switch interval {
	case "1m":
		return 1 * time.Minute
	case "3m":
		return 3 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "2h":
		return 2 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "6h":
		return 6 * time.Hour
	case "8h":
		return 8 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "1d":
		return 24 * time.Hour
	case "3d":
		return 3 * 24 * time.Hour
	case "1w":
		return 7 * 24 * time.Hour
	default:
		return 1 * time.Minute
	}
}
