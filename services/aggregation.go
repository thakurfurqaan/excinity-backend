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

func (s *AggregationService) isNewCandlePeriod(now time.Time, candle models.Candle) bool {
	return now.Minute() != candle.Timestamp.Minute()
}

func (s *AggregationService) handleNewCandle(tick exchange.Tick, now time.Time, currentCandle models.Candle, exists bool) models.Candle {
	if exists {
		s.db.SaveCandle(&currentCandle)
	}
	return models.Candle{
		Symbol:    tick.Symbol,
		Timestamp: now.Truncate(time.Minute),
		Open:      tick.Price,
		High:      tick.Price,
		Low:       tick.Price,
		Close:     tick.Price,
	}
}

func (s *AggregationService) updateExistingCandle(candle models.Candle, tick exchange.Tick) models.Candle {
	candle.Close = tick.Price
	candle.High = max(candle.High, tick.Price)
	candle.Low = min(candle.Low, tick.Price)
	return candle
}

func (s *AggregationService) processTickToCandle(tick exchange.Tick) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	currentCandle, ok := s.currentCandles[tick.Symbol]

	if !ok || s.isNewCandlePeriod(now, currentCandle) {
		currentCandle = s.handleNewCandle(tick, now, currentCandle, ok)
	} else {
		currentCandle = s.updateExistingCandle(currentCandle, tick)
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
