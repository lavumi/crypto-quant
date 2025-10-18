package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// CandleRepository handles candle data operations
type CandleRepository struct {
	db *DB
}

// NewCandleRepository creates a new candle repository
func NewCandleRepository(db *DB) *CandleRepository {
	return &CandleRepository{db: db}
}

// Save saves a candle to the database
func (r *CandleRepository) Save(ctx context.Context, candle *domain.Candle, interval string) error {
	query := `
		INSERT INTO candles (symbol, interval, open_time, close_time, open, high, low, close, volume)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol, interval, open_time) DO UPDATE SET
			close_time = excluded.close_time,
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			volume = excluded.volume
	`

	_, err := r.db.ExecContext(ctx, query,
		candle.Symbol,
		interval,
		candle.OpenTime.Unix(),
		candle.CloseTime.Unix(),
		candle.Open,
		candle.High,
		candle.Low,
		candle.Close,
		candle.Volume,
	)

	if err != nil {
		return fmt.Errorf("failed to save candle: %w", err)
	}

	return nil
}

// SaveBatch saves multiple candles in a transaction
func (r *CandleRepository) SaveBatch(ctx context.Context, candles []*domain.Candle, interval string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO candles (symbol, interval, open_time, close_time, open, high, low, close, volume)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol, interval, open_time) DO UPDATE SET
			close_time = excluded.close_time,
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			volume = excluded.volume
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, candle := range candles {
		_, err := stmt.ExecContext(ctx,
			candle.Symbol,
			interval,
			candle.OpenTime.Unix(),
			candle.CloseTime.Unix(),
			candle.Open,
			candle.High,
			candle.Low,
			candle.Close,
			candle.Volume,
		)
		if err != nil {
			return fmt.Errorf("failed to insert candle: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetRange retrieves candles within a time range
func (r *CandleRepository) GetRange(ctx context.Context, symbol, interval string, start, end time.Time) ([]*domain.Candle, error) {
	query := `
		SELECT symbol, interval, open_time, close_time, open, high, low, close, volume
		FROM candles
		WHERE symbol = ? AND interval = ? AND open_time >= ? AND open_time < ?
		ORDER BY open_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query, symbol, interval, start.Unix(), end.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to query candles: %w", err)
	}
	defer rows.Close()

	var candles []*domain.Candle
	for rows.Next() {
		var candle domain.Candle
		var intervalStr string
		var openTime, closeTime int64

		err := rows.Scan(
			&candle.Symbol,
			&intervalStr,
			&openTime,
			&closeTime,
			&candle.Open,
			&candle.High,
			&candle.Low,
			&candle.Close,
			&candle.Volume,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan candle: %w", err)
		}

		candle.OpenTime = time.Unix(openTime, 0)
		candle.CloseTime = time.Unix(closeTime, 0)
		candles = append(candles, &candle)
	}

	return candles, nil
}

// GetLatest retrieves the most recent candle
func (r *CandleRepository) GetLatest(ctx context.Context, symbol, interval string) (*domain.Candle, error) {
	query := `
		SELECT symbol, interval, open_time, close_time, open, high, low, close, volume
		FROM candles
		WHERE symbol = ? AND interval = ?
		ORDER BY open_time DESC
		LIMIT 1
	`

	var candle domain.Candle
	var intervalStr string
	var openTime, closeTime int64

	err := r.db.QueryRowContext(ctx, query, symbol, interval).Scan(
		&candle.Symbol,
		&intervalStr,
		&openTime,
		&closeTime,
		&candle.Open,
		&candle.High,
		&candle.Low,
		&candle.Close,
		&candle.Volume,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest candle: %w", err)
	}

	candle.OpenTime = time.Unix(openTime, 0)
	candle.CloseTime = time.Unix(closeTime, 0)

	return &candle, nil
}

// TradeRepository handles trade data operations
type TradeRepository struct {
	db *DB
}

// NewTradeRepository creates a new trade repository
func NewTradeRepository(db *DB) *TradeRepository {
	return &TradeRepository{db: db}
}

// Save saves a trade to the database
func (r *TradeRepository) Save(ctx context.Context, trade *domain.Trade) error {
	query := `
		INSERT INTO trades (order_id, symbol, side, price, quantity, fee, fee_asset, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		trade.OrderID,
		trade.Symbol,
		string(trade.Side),
		trade.Price,
		trade.Quantity,
		trade.Fee,
		trade.FeeAsset,
		trade.Timestamp.Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to save trade: %w", err)
	}

	return nil
}

// GetBySymbol retrieves all trades for a symbol
func (r *TradeRepository) GetBySymbol(ctx context.Context, symbol string) ([]*domain.Trade, error) {
	query := `
		SELECT id, order_id, symbol, side, price, quantity, fee, fee_asset, timestamp
		FROM trades
		WHERE symbol = ?
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to query trades: %w", err)
	}
	defer rows.Close()

	var trades []*domain.Trade
	for rows.Next() {
		var trade domain.Trade
		var id int64
		var sideStr string
		var timestamp int64

		err := rows.Scan(
			&id,
			&trade.OrderID,
			&trade.Symbol,
			&sideStr,
			&trade.Price,
			&trade.Quantity,
			&trade.Fee,
			&trade.FeeAsset,
			&timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}

		trade.ID = fmt.Sprintf("%d", id)
		trade.Side = domain.OrderSide(sideStr)
		trade.Timestamp = time.Unix(timestamp, 0)
		trades = append(trades, &trade)
	}

	return trades, nil
}
