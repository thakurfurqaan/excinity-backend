package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"excinity/models"
	"excinity/routes"

	"github.com/gorilla/websocket"
)

type BinanceTick struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
}

func startBinanceClient(symbol string) {

	log.Println("Starting Binance client for symbol:", symbol)

	url := "wss://stream.binance.com:9443/ws/" + symbol + "@aggTrade"
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	defer func() {
		log.Println("Closing Binance WebSocket connection for symbol:", symbol)
		err := c.Close()
		if err != nil {
			log.Printf("Error closing Binance WebSocket connection for symbol %s: %v", symbol, err)
		} else {
			log.Println("Successfully closed Binance WebSocket connection for symbol:", symbol)
		}
	}()

	log.Println("Connected to Binance WebSocket for symbol:", symbol)

	var currentCandle models.Candle
	var candleStartTime time.Time

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		var tick BinanceTick
		err = json.Unmarshal(message, &tick)
		if err != nil {
			log.Println("unmarshal:", err)
			continue
		}

		price, err := strconv.ParseFloat(tick.Price, 64)
		if err != nil {
			log.Println("parse price:", err)
			continue
		}

		now := time.Now().UTC()
		if now.Minute() != candleStartTime.Minute() || currentCandle.Open == 0 {
			if currentCandle.Open != 0 {
				routes.BroadcastData(currentCandle)
			}
			candleStartTime = now.Truncate(time.Minute)
			currentCandle = models.Candle{
				Symbol:    symbol,
				Timestamp: candleStartTime,
				Open:      price,
				High:      price,
				Low:       price,
				Close:     price,
			}
		} else {
			currentCandle.Close = price
			if price > currentCandle.High {
				currentCandle.High = price
			}
			if price < currentCandle.Low {
				currentCandle.Low = price
			}
		}

		routes.BroadcastData(currentCandle)
	}
}
