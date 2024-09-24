package database

import (
	"excinity/models"

	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

func (s *Database) SaveCandle(candle *models.Candle) error {
	return s.db.Create(candle).Error
}

func (s *Database) GetHistoricalData(symbol string, limit int) ([]models.Candle, error) {
	var candles []models.Candle
	err := s.db.Where("symbol = ?", symbol).Order("timestamp desc").Limit(limit).Find(&candles).Error
	return candles, err
}
