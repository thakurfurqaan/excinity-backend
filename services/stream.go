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

	for {
		select {
		case tick, ok := <-tickChan:
			if !ok {
				log.Printf("Tick channel closed for symbol %s", symbol)
				return nil
			}
			currentCandle, err = processTickToCandle(tick, currentCandle)
			if err != nil {
				log.Printf("Failed to process tick to candle: %v", err)
				continue
			}
			go routes.BroadcastData(currentCandle)
		case <-ctx.Done():
			log.Printf("Context cancelled for symbol %s", symbol)
			return nil
		default:
			time.Sleep(100 * time.Millisecond)
		}

	}
}

func processTickToCandle(tick exchange.Tick, currentCandle models.Candle) (models.Candle, error) {
	now := time.Now().UTC()
	if now.Minute() != currentCandle.Timestamp.Minute() || currentCandle.Open == 0 {
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
	return currentCandle, nil
}
