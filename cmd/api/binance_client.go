package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type BinanceTick struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
}

type Candle struct {
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
}

func startBinanceClient(symbol string) {

	log.Println("Starting Binance client for symbol:", symbol)

	url := "wss://stream.binance.com:9443/ws/" + symbol + "@aggTrade"
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	log.Println("Connected to Binance WebSocket for symbol:", symbol)

	var currentCandle Candle
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
		if now.Sub(candleStartTime) >= time.Minute || currentCandle.Open == 0 {
			if currentCandle.Open != 0 {
				broadcastCandle(currentCandle)
			}
			candleStartTime = now.Truncate(time.Minute)
			currentCandle = Candle{
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

		broadcastCandle(currentCandle)
	}
}
