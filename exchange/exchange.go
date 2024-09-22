package main

import (
	"context"
	"time"
)

type Tick struct {
	Symbol string
	Price  float64
}

type ExchangeClient interface {
	Connect(ctx context.Context, symbol string) (<-chan Tick, error)
	GetAvailableSymbols() ([]string, error)
	GetHistoricalData(symbol string, limit int) ([]Candle, error)
}

type Candle struct {
	Symbol    string
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
}
