package db

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"excinity/models"
)

type InMemoryDB struct {
	data     sync.Map
	filename string
}

type CandleKey struct {
	Symbol string
	Time   time.Time
}

func NewInMemoryDB(filename string) (*InMemoryDB, error) {
	db := &InMemoryDB{filename: filename}
	if err := db.load(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *InMemoryDB) InsertCandle(symbol string, timestamp time.Time, open, high, low, close float64) error {
	key := CandleKey{Symbol: symbol, Time: timestamp.Truncate(time.Minute)}
	candle := models.Candle{
		Symbol:    symbol,
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
	}
	db.data.Store(key, candle)
	return db.save()
}

func (db *InMemoryDB) GetCandles(symbol string, start, end time.Time) ([]models.Candle, error) {
	var candles []models.Candle
	db.data.Range(func(key, value interface{}) bool {
		k := key.(CandleKey)
		if k.Symbol == symbol && !k.Time.Before(start) && k.Time.Before(end) {
			candles = append(candles, value.(models.Candle))
		}
		return true
	})
	return candles, nil
}

func (db *InMemoryDB) save() error {
	file, err := os.Create(db.filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	var data []models.Candle
	db.data.Range(func(_, value interface{}) bool {
		data = append(data, value.(models.Candle))
		return true
	})

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}
	return nil
}

func (db *InMemoryDB) load() error {
	file, err := os.Open(db.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // It's okay if the file doesn't exist yet
		}
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var data []models.Candle
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	for _, candle := range data {
		key := CandleKey{Symbol: candle.Symbol, Time: candle.Timestamp}
		db.data.Store(key, candle)
	}
	return nil
}
