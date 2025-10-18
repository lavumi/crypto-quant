package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQL database connection
type DB struct {
	*sqlx.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection with sqlx
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	return &DB{db}, nil
}

// Migrate runs database migrations
func (db *DB) Migrate() error {
	// Interval-specific candle tables
	intervals := []string{"1m", "5m", "15m", "30m", "1h", "4h", "1d"}

	var migrations []string

	// Create separate table for each interval
	for _, interval := range intervals {
		tableName := fmt.Sprintf("candles_%s", interval)
		migrations = append(migrations,
			fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				symbol TEXT NOT NULL,
				open_time INTEGER NOT NULL,
				close_time INTEGER NOT NULL,
				open REAL NOT NULL,
				high REAL NOT NULL,
				low REAL NOT NULL,
				close REAL NOT NULL,
				volume REAL NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(symbol, open_time)
			)`, tableName),
			fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_symbol_time
				ON %s(symbol, open_time DESC)`, tableName, tableName),
		)
	}

	// Trades table
	migrations = append(migrations,
		`CREATE TABLE IF NOT EXISTS trades (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id TEXT NOT NULL,
			symbol TEXT NOT NULL,
			side TEXT NOT NULL,
			price REAL NOT NULL,
			quantity REAL NOT NULL,
			fee REAL NOT NULL,
			fee_asset TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_trades_symbol_timestamp
			ON trades(symbol, timestamp DESC)`,
	)

	// Backtest results table
	migrations = append(migrations,
		`CREATE TABLE IF NOT EXISTS backtest_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			strategy_name TEXT NOT NULL,
			symbol TEXT NOT NULL,
			start_time INTEGER NOT NULL,
			end_time INTEGER NOT NULL,
			initial_balance REAL NOT NULL,
			final_balance REAL NOT NULL,
			total_return REAL NOT NULL,
			sharpe_ratio REAL,
			max_drawdown REAL,
			win_rate REAL,
			total_trades INTEGER NOT NULL,
			config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_backtest_strategy
			ON backtest_results(strategy_name, created_at DESC)`,
	)

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
