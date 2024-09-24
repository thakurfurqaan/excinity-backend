package exchange

import (
	"context"
	"excinity/models"
)

type Tick struct {
	Symbol string
	Price  float64
}

type ExchangeClient interface {
	Connect(ctx context.Context, symbol string) (<-chan Tick, error)
	GetHistoricalData(symbol string, limit int) ([]models.Candle, error)
}
