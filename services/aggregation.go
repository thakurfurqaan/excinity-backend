package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"excinity/database"
	"excinity/exchange"
	"excinity/models"
)

type AggregationService struct {
	db             *database.Database
	currentCandles map[string]models.Candle
	mu             sync.RWMutex
}

func NewAggregationService(db *database.Database) *AggregationService {
	return &AggregationService{
		db:             db,
		currentCandles: make(map[string]models.Candle),
	}
}

func (s *AggregationService) StartSymbolStream(client exchange.ExchangeClient, symbol string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickChan, err := client.Connect(ctx, symbol)
	if err != nil {
		log.Printf("Failed to connect to exchange for symbol %s: %v", symbol, err)
		return err
	}

	for tick := range tickChan {
		s.processTickToCandle(tick)
	}

	return nil
}

func (s *AggregationService) processTickToCandle(tick exchange.Tick) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	currentCandle, ok := s.currentCandles[tick.Symbol]

	if !ok || now.Minute() != currentCandle.Timestamp.Minute() {
		if ok {
			s.db.SaveCandle(&currentCandle)
		}
		currentCandle = models.Candle{
			Symbol:    tick.Symbol,
			Timestamp: now.Truncate(time.Minute),
			Open:      tick.Price,
			High:      tick.Price,
			Low:       tick.Price,
			Close:     tick.Price,
		}
	} else {
		currentCandle.Close = tick.Price
		if tick.Price > currentCandle.High {
			currentCandle.High = tick.Price
		}
		if tick.Price < currentCandle.Low {
			currentCandle.Low = tick.Price
		}
	}

	s.currentCandles[tick.Symbol] = currentCandle
}

func (s *AggregationService) GetLatestCandle(symbol string) (models.Candle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	candle, ok := s.currentCandles[symbol]
	if !ok {
		return models.Candle{}, fmt.Errorf("no data for symbol: %s", symbol)
	}
	return candle, nil
}

func (s *AggregationService) GetHistoricalData(symbol string, limit int) ([]models.Candle, error) {
	return s.db.GetHistoricalData(symbol, limit)
}
