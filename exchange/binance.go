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

const (
	maxRetries        = 5
	reconnectInterval = 5 * time.Second
)

type BinanceClient struct {
	wsUrl string
}

type BinanceTick struct {
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

func (b *BinanceClient) dialWebSocket(ctx context.Context, symbol string) (*websocket.Conn, error) {
	url := fmt.Sprintf("%s/%s@aggTrade", b.wsUrl, symbol)
	log.Printf("Connecting to Binance WebSocket for symbol: %s", symbol)

	c, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	log.Printf("Connected to Binance WebSocket for symbol: %s", symbol)
	return c, nil
}

func (b *BinanceClient) handleWebSocketConnection(ctx context.Context, c *websocket.Conn, symbol string, tickChan chan<- Tick) error {
	defer closeWebsocketConn(c, symbol)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled for symbol %s", symbol)
			return nil
		default:
			if err := b.readAndProcessMessage(c, tickChan); err != nil {
				log.Printf("Error processing message for symbol %s: %v", symbol, err)
				return err
			}
		}
	}
}

func (b *BinanceClient) readAndProcessMessage(c *websocket.Conn, tickChan chan<- Tick) error {
	if err := c.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return fmt.Errorf("error setting read deadline: %w", err)
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return fmt.Errorf("unexpected WebSocket close: %w", err)
		}
		return err
	}

	tick, err := b.parseMessage(message)
	if err != nil {
		return err
	}

	tickChan <- tick
	return nil
}

func (b *BinanceClient) parseMessage(message []byte) (Tick, error) {
	var binanceRawTick BinanceTick

	if err := json.Unmarshal(message, &binanceRawTick); err != nil {
		return Tick{}, fmt.Errorf("error unmarshalling message: %w", err)
	}

	price, err := strconv.ParseFloat(binanceRawTick.Price, 64)
	if err != nil {
		return Tick{}, fmt.Errorf("error parsing price: %w", err)
	}

	tick := Tick{Symbol: binanceRawTick.Symbol, Price: price}
	return tick, nil
}

func (b *BinanceClient) Connect(ctx context.Context, symbol string) (<-chan Tick, error) {
	tickChan := make(chan Tick)

	go b.manageConnection(ctx, symbol, tickChan)

	return tickChan, nil
}

func (b *BinanceClient) manageConnection(ctx context.Context, symbol string, tickChan chan<- Tick) {
	defer close(tickChan)

	retries := 0
	for {
		if err := b.connectAndHandle(ctx, symbol, tickChan); err != nil {
			if !b.shouldRetry(ctx, symbol, &retries) {
				return
			}
			continue
		}
		retries = 0
	}
}

func (b *BinanceClient) connectAndHandle(ctx context.Context, symbol string, tickChan chan<- Tick) error {
	c, err := b.dialWebSocket(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer closeWebsocketConn(c, symbol)

	return b.handleWebSocketConnection(ctx, c, symbol, tickChan)
}

func (b *BinanceClient) shouldRetry(ctx context.Context, symbol string, retries *int) bool {
	*retries++
	if *retries >= maxRetries {
		log.Printf("Max retries reached for symbol %s. Stopping reconnection attempts.", symbol)
		return false
	}

	log.Printf("Attempting to reconnect for symbol %s (Retry %d/%d)", symbol, *retries, maxRetries)

	select {
	case <-ctx.Done():
		return false
	case <-time.After(reconnectInterval):
		return true
	}
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
