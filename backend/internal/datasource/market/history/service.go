package history

import (
	"context"
	"fmt"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/domain"
)

// Service handles historical data operations
type Service struct {
	candleRepo *CandleRepository
	tradeRepo  *TradeRepository
	collector  *Collector
}

// NewService creates a new data service
func NewService(candleRepo *CandleRepository, tradeRepo *TradeRepository, binanceClient *binance.Client) *Service {
	return &Service{
		candleRepo: candleRepo,
		tradeRepo:  tradeRepo,
		collector:  NewCollector(binanceClient, candleRepo),
	}
}

// CollectHistoricalData collects historical candle data
func (s *Service) CollectHistoricalData(ctx context.Context, symbol, interval string, startTime, endTime time.Time) error {
	if err := s.collector.CollectHistorical(ctx, symbol, interval, startTime, endTime); err != nil {
		return fmt.Errorf("failed to collect historical data: %w", err)
	}
	return nil
}

// GetCandles retrieves candles within a time range
func (s *Service) GetCandles(ctx context.Context, symbol, interval string, start, end time.Time) ([]*domain.Candle, error) {
	candles, err := s.candleRepo.GetRange(ctx, symbol, interval, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get candles: %w", err)
	}
	return candles, nil
}

// GetLatestCandle retrieves the most recent candle
func (s *Service) GetLatestCandle(ctx context.Context, symbol, interval string) (*domain.Candle, error) {
	candle, err := s.candleRepo.GetLatest(ctx, symbol, interval)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest candle: %w", err)
	}
	if candle == nil {
		return nil, fmt.Errorf("no candle data found for %s %s", symbol, interval)
	}
	return candle, nil
}

// SaveTrade saves a trade to the database
func (s *Service) SaveTrade(ctx context.Context, trade *domain.Trade) error {
	if err := s.tradeRepo.Save(ctx, trade); err != nil {
		return fmt.Errorf("failed to save trade: %w", err)
	}
	return nil
}

// GetTradeHistory retrieves all trades for a symbol
func (s *Service) GetTradeHistory(ctx context.Context, symbol string) ([]*domain.Trade, error) {
	trades, err := s.tradeRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade history: %w", err)
	}
	return trades, nil
}

// DataValidationResult represents the result of data validation
type DataValidationResult struct {
	HasData       bool       `json:"has_data"`
	AvailableFrom *time.Time `json:"available_from,omitempty"`
	AvailableTo   *time.Time `json:"available_to,omitempty"`
	RequestedFrom time.Time  `json:"requested_from"`
	RequestedTo   time.Time  `json:"requested_to"`
	CandleCount   int        `json:"candle_count"`
	IsComplete    bool       `json:"is_complete"`
	Message       string     `json:"message"`
}

// ValidateDataAvailability checks if data is available for the requested period
func (s *Service) ValidateDataAvailability(ctx context.Context, symbol, interval string, startTime, endTime time.Time) (*DataValidationResult, error) {
	result := &DataValidationResult{
		RequestedFrom: startTime,
		RequestedTo:   endTime,
		HasData:       false,
		IsComplete:    false,
	}

	// Get candles in the requested range
	candles, err := s.candleRepo.GetRange(ctx, symbol, interval, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check data availability: %w", err)
	}

	result.CandleCount = len(candles)

	if len(candles) == 0 {
		result.Message = fmt.Sprintf("No data available for %s %s in the requested period", symbol, interval)
		return result, nil
	}

	result.HasData = true
	firstCandle := candles[0].OpenTime
	lastCandle := candles[len(candles)-1].OpenTime
	result.AvailableFrom = &firstCandle
	result.AvailableTo = &lastCandle

	// Compare dates only (ignore time component)
	// Truncate to day boundary for accurate comparison
	firstCandleDate := time.Date(firstCandle.Year(), firstCandle.Month(), firstCandle.Day(), 0, 0, 0, 0, time.UTC)
	lastCandleDate := time.Date(lastCandle.Year(), lastCandle.Month(), lastCandle.Day(), 0, 0, 0, 0, time.UTC)
	requestedStartDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
	requestedEndDate := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, time.UTC)

	// Check if the available data covers the requested period
	// firstCandleDate should be <= requestedStartDate
	// lastCandleDate should be >= requestedEndDate
	if firstCandleDate.After(requestedStartDate) || lastCandleDate.Before(requestedEndDate) {
		result.IsComplete = false
		result.Message = fmt.Sprintf("Data is incomplete. Available: %s to %s, Requested: %s to %s (Missing: %d candles)",
			firstCandleDate.Format("2006-01-02"),
			lastCandleDate.Format("2006-01-02"),
			requestedStartDate.Format("2006-01-02"),
			requestedEndDate.Format("2006-01-02"),
			len(candles))
	} else {
		result.IsComplete = true
		result.Message = fmt.Sprintf("Data is complete with %d candles", len(candles))
	}

	return result, nil
}

// GetDataRange returns the first and last candle times for a symbol/interval
func (s *Service) GetDataRange(ctx context.Context, symbol, interval string) (*time.Time, *time.Time, error) {
	// Get the first candle
	firstCandle, err := s.candleRepo.GetFirst(ctx, symbol, interval)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get first candle: %w", err)
	}
	if firstCandle == nil {
		return nil, nil, nil // No data
	}

	// Get the last candle
	lastCandle, err := s.candleRepo.GetLatest(ctx, symbol, interval)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get last candle: %w", err)
	}
	if lastCandle == nil {
		return nil, nil, nil // No data
	}

	return &firstCandle.OpenTime, &lastCandle.OpenTime, nil
}
