package main

import (
	"log"

	"excinity/config"
	"excinity/database"
	"excinity/exchange"
	"excinity/routes"
	"excinity/services"
)

func startSymbolStream(s config.Symbol, exchangeFactory *exchange.ExchangeFactory, aggregationService *services.AggregationService) {
	exchange := s.Exchange

	exchangeClient, err := exchangeFactory.Create(exchange)
	if err != nil {
		log.Fatal(err)
	}
	err = aggregationService.StartSymbolStream(exchangeClient, s.Symbol)
	if err != nil {
		log.Printf("Error starting stream for symbol %s: %v", s.Symbol, err)
	}
}

func initDb(cfg *config.Config) *database.Database {
	db, err := database.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	return database.NewDatabase(db)
}

func loadConfig(path string) *config.Config {
	cfg, err := config.LoadFile(path)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}

const CONFIG_PATH = "config.yml"

func main() {
	cfg := loadConfig(CONFIG_PATH)
	database := initDb(cfg)
	exchangeFactory := exchange.NewExchangeFactory(cfg)
	aggregationService := services.NewAggregationService(database)

	for _, symbol := range cfg.Symbols {
		go startSymbolStream(symbol, exchangeFactory, aggregationService)
	}

	router := routes.SetupRoutes(aggregationService)
	StartServer(cfg, router)
}
