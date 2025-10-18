package history

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/lavumi/crypto-quant/internal/datasource/database"
	"github.com/lavumi/crypto-quant/internal/domain"
)

// CandleRepository handles candle data operations
type CandleRepository struct {
	db *database.DB
}

// NewCandleRepository creates a new candle repository
func NewCandleRepository(db *database.DB) *CandleRepository {
	return &CandleRepository{db: db}
}

// getTableName returns the table name for a given interval
func getTableName(interval string) string {
	return fmt.Sprintf("candles_%s", interval)
}

// Save saves a candle to the database
func (r *CandleRepository) Save(ctx context.Context, candle *domain.Candle, interval string) error {
	tableName := getTableName(interval)

	query := fmt.Sprintf(`
		INSERT INTO %s (symbol, open_time, close_time, open, high, low, close, volume)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol, open_time) DO UPDATE SET
			close_time = excluded.close_time,
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			volume = excluded.volume
	`, tableName)

	_, err := r.db.ExecContext(ctx, query,
		candle.Symbol,
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
	tableName := getTableName(interval)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`
		INSERT INTO %s (symbol, open_time, close_time, open, high, low, close, volume)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol, open_time) DO UPDATE SET
			close_time = excluded.close_time,
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			volume = excluded.volume
	`, tableName)

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, candle := range candles {
		_, err := stmt.ExecContext(ctx,
			candle.Symbol,
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
	tableName := getTableName(interval)

	// Build query with Squirrel
	query := sq.Select("symbol", "open_time", "close_time", "open", "high", "low", "close", "volume").
		From(tableName).
		Where(sq.Eq{"symbol": symbol}).
		Where(sq.GtOrEq{"open_time": start.Unix()}).
		Where(sq.Lt{"open_time": end.Unix()}).
		OrderBy("open_time ASC")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// Query with sqlx
	rows, err := r.db.QueryxContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query candles: %w", err)
	}
	defer rows.Close()

	var candles []*domain.Candle
	for rows.Next() {
		var candle domain.Candle
		var openTime, closeTime int64

		err := rows.Scan(
			&candle.Symbol,
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

// GetFirst retrieves the first (oldest) candle
func (r *CandleRepository) GetFirst(ctx context.Context, symbol, interval string) (*domain.Candle, error) {
	tableName := getTableName(interval)

	// Build query with Squirrel
	query := sq.Select("symbol", "open_time", "close_time", "open", "high", "low", "close", "volume").
		From(tableName).
		Where(sq.Eq{"symbol": symbol}).
		OrderBy("open_time ASC").
		Limit(1)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var candle domain.Candle
	var openTime, closeTime int64

	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&candle.Symbol,
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
		return nil, fmt.Errorf("failed to get first candle: %w", err)
	}

	candle.OpenTime = time.Unix(openTime, 0)
	candle.CloseTime = time.Unix(closeTime, 0)

	return &candle, nil
}

// GetLatest retrieves the most recent candle
func (r *CandleRepository) GetLatest(ctx context.Context, symbol, interval string) (*domain.Candle, error) {
	tableName := getTableName(interval)

	// Build query with Squirrel
	query := sq.Select("symbol", "open_time", "close_time", "open", "high", "low", "close", "volume").
		From(tableName).
		Where(sq.Eq{"symbol": symbol}).
		OrderBy("open_time DESC").
		Limit(1)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var candle domain.Candle
	var openTime, closeTime int64

	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&candle.Symbol,
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
	db *database.DB
}

// NewTradeRepository creates a new trade repository
func NewTradeRepository(db *database.DB) *TradeRepository {
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
