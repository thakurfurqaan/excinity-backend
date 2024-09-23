package services

import (
	"context"
	"log"
	"time"

	"excinity/exchange"
	"excinity/models"
	"excinity/routes"
)

func StartSymbolStream(exchangeClient exchange.ExchangeClient, symbol string) {
	ctx := context.Background()
	tickChan, err := exchangeClient.Connect(ctx, symbol)
	if err != nil {
		log.Printf("Failed to connect to %s stream: %v", symbol, err)
		return
	}

	var currentCandle models.Candle
	var candleStartTime time.Time

	for tick := range tickChan {
		currentCandle, err = processTickToCandle(tick, currentCandle, candleStartTime)
		if err != nil {
			log.Printf("Failed to process tick to candle: %v", err)
			continue
		}
		routes.BroadcastData(currentCandle)
	}
}

func processTickToCandle(tick exchange.Tick, currentCandle models.Candle, candleStartTime time.Time) (models.Candle, error) {
	log.Println("Processing tick to candle:", tick)
	now := time.Now().UTC()
	if now.Minute() != candleStartTime.Minute() || currentCandle.Open == 0 {
		if currentCandle.Open != 0 {
			routes.BroadcastData(currentCandle)
		}
		candleStartTime = now.Truncate(time.Minute)
		currentCandle = models.Candle{
			Symbol:    tick.Symbol,
			Timestamp: candleStartTime,
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
