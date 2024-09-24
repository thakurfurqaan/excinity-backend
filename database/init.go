package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"excinity/config"
	"excinity/models"
)

func Migrate(db *gorm.DB) error {
	log.Println("Migrating database schema")
	return db.AutoMigrate(&models.Candle{})
}

func InitDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.DSN
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		return nil, err
	}

	err = Migrate(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
