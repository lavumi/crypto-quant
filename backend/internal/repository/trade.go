package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/lavumi/crypto-quant/internal/datasource/database"
	"github.com/lavumi/crypto-quant/internal/domain"
)

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
