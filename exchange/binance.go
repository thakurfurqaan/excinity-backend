package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"excinity/models"

	"github.com/gorilla/websocket"
)

type BinanceClient struct {
	wsUrl string
}

var binanceRawTick struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
}

func NewBinanceClient(config map[string]interface{}) (ExchangeClient, error) {
	wsURL, ok := config["ws_url"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid ws_url in Binance config")
	}
	return &BinanceClient{wsUrl: wsURL}, nil
}

func (b *BinanceClient) Connect(ctx context.Context, symbol string) (<-chan Tick, error) {
	tickChan := make(chan Tick)

	url := fmt.Sprintf("%s/%s@aggTrade", b.wsUrl, symbol)

	log.Println("Starting Binance client for symbol:", url, symbol)

	c, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
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
				log.Printf("Context cancelled for symbol %s", symbol)
				return
			default:
				err := c.SetReadDeadline(time.Now().Add(10 * time.Second))
				if err != nil {
					log.Printf("Error setting read deadline for symbol %s: %v", symbol, err)
					return
				}

				_, message, err := c.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("Unexpected WebSocket close for symbol %s: %v", symbol, err)
					}
					return
				}

				if err := json.Unmarshal(message, &binanceRawTick); err != nil {
					log.Printf("Error unmarshalling message for symbol %s: %v", symbol, err)
					continue
				}

				price, err := strconv.ParseFloat(binanceRawTick.Price, 64)
				if err != nil {
					log.Printf("Error parsing price for symbol %s: %v", binanceRawTick.Symbol, err)
					continue
				}

				tickChan <- Tick{Symbol: binanceRawTick.Symbol, Price: price}
			}
		}
	}()

	return tickChan, nil
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
