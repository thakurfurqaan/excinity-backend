package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"excinity/models"

	"github.com/gorilla/websocket"
)

type BinanceClient struct {
	// Add any Binance-specific fields here
}

func NewBinanceClient() *BinanceClient {
	return &BinanceClient{}
}

func (b *BinanceClient) Connect(ctx context.Context, symbol string) (<-chan Tick, error) {
	tickChan := make(chan Tick)

	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@aggTrade", symbol)

	log.Println("Starting Binance client for symbol:", symbol)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	log.Println("Connected to Binance WebSocket for symbol:", symbol)

	go func() {
		defer close(tickChan)
		defer closeWebsocketConn(c, symbol)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					// Handle error (log, retry, etc.)
					continue
				}

				var rawTick struct {
					Symbol string `json:"s"`
					Price  string `json:"p"`
				}

				if err := json.Unmarshal(message, &rawTick); err != nil {
					// Handle error
					continue
				}

				price, err := strconv.ParseFloat(rawTick.Price, 64)
				if err != nil {
					log.Printf("Error parsing price for symbol %s: %v", rawTick.Symbol, err)
					continue
				}

				tickChan <- Tick{Symbol: rawTick.Symbol, Price: price}
			}
		}
	}()

	return tickChan, nil
}

func (b *BinanceClient) GetAvailableSymbols() ([]string, error) {
	// For simplicity, we're returning a static list here
	return []string{"BTCUSDT", "ETHUSDT", "PEPEUSDT"}, nil
}

func (b *BinanceClient) GetHistoricalData(symbol string, limit int) ([]models.Candle, error) {
	// This is a placeholder implementation
	candles := make([]models.Candle, limit)
	return candles, nil
}

func closeWebsocketConn(c *websocket.Conn, symbol string) {
	log.Println("Closing Binance WebSocket connection for symbol:", symbol)
	err := c.Close()
	if err != nil {
		log.Printf("Error closing Binance WebSocket connection for symbol %s: %v", symbol, err)
	} else {
		log.Println("Successfully closed Binance WebSocket connection for symbol:", symbol)
	}
}
