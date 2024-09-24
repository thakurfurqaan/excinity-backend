package models

import (
	"time"

	"gorm.io/gorm"
)

type Candle struct {
	gorm.Model
	Symbol    string    `json:"symbol" gorm:"index:idx_symbol_timestamp"`
	Timestamp time.Time `json:"timestamp" gorm:"index:idx_symbol_timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
}
