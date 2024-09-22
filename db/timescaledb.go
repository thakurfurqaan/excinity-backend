package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type DBHandler struct {
	db *sql.DB
}

func NewDBHandler(connectionString string) (*DBHandler, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DBHandler{db: db}, nil
}

func (h *DBHandler) InsertCandle(symbol string, timestamp time.Time, open, high, low, close float64) error {
	_, err := h.db.Exec(`
		INSERT INTO candles (symbol, timestamp, open, high, low, close)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (symbol, timestamp) 
		DO UPDATE SET open = $3, high = $4, low = $5, close = $6
	`, symbol, timestamp, open, high, low, close)

	if err != nil {
		return fmt.Errorf("failed to insert candle: %w", err)
	}

	return nil
}

func (h *DBHandler) GetCandles(symbol string, start, end time.Time) ([]Candle, error) {
	rows, err := h.db.Query(`
		SELECT timestamp, open, high, low, close
		FROM candles
		WHERE symbol = $1 AND timestamp >= $2 AND timestamp < $3
		ORDER BY timestamp
	`, symbol, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query candles: %w", err)
	}
	defer rows.Close()

	var candles []Candle
	for rows.Next() {
		var c Candle
		if err := rows.Scan(&c.Timestamp, &c.Open, &c.High, &c.Low, &c.Close); err != nil {
			return nil, fmt.Errorf("failed to scan candle: %w", err)
		}
		c.Symbol = symbol
		candles = append(candles, c)
	}

	return candles, nil
}
