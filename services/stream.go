package services

import (
	"context"
	"log"
	"time"

	"excinity/exchange"
	"excinity/models"
	"excinity/routes"
)

func StartSymbolStream(exchangeClient exchange.ExchangeClient, symbol string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickChan, err := exchangeClient.Connect(ctx, symbol)
	if err != nil {
		return err
	}

	var currentCandle models.Candle
	var candleStartTime time.Time

	for {
		select {
		case tick, ok := <-tickChan:
			if !ok {
				log.Printf("Tick channel closed for symbol %s", symbol)
				return nil
			}
			currentCandle, err = processTickToCandle(tick, currentCandle, &candleStartTime)
			if err != nil {
				log.Printf("Failed to process tick to candle: %v", err)
				continue
			}
			routes.BroadcastData(currentCandle)
		case <-ctx.Done():
			log.Printf("Context cancelled for symbol %s", symbol)
			return nil
		}
	}
}

func processTickToCandle(tick exchange.Tick, currentCandle models.Candle, candleStartTime *time.Time) (models.Candle, error) {
	log.Printf("Processing tick to candle: %+v", tick)
	now := time.Now().UTC()
	if now.Minute() != candleStartTime.Minute() || currentCandle.Open == 0 {
		if currentCandle.Open != 0 {
			routes.BroadcastData(currentCandle)
		}
		*candleStartTime = now.Truncate(time.Minute)
		currentCandle = models.Candle{
			Symbol:    tick.Symbol,
			Timestamp: *candleStartTime,
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

	return currentCandle, nil
}
