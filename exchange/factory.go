package exchange

import (
	"fmt"

	"excinity/config"
)

type ExchangeClientCreator func(config map[string]interface{}) (ExchangeClient, error)

type ExchangeFactory struct {
	creators map[string]ExchangeClientCreator
	cfg      *config.Config
}

func NewExchangeFactory(cfg *config.Config) *ExchangeFactory {
	return &ExchangeFactory{
		creators: map[string]ExchangeClientCreator{
			"binance":  NewBinanceClient,
			"coinbase": NewCoinbaseClient,
		},
		cfg: cfg,
	}
}

func getExchangeConfig(name string, cfg *config.Config) (map[string]interface{}, error) {
	for _, e := range cfg.Exchanges {
		if e.Name == name {
			return e.Config, nil
		}
	}
	return nil, fmt.Errorf("exchange config not found: %s", name)
}

func (f *ExchangeFactory) Create(name string) (ExchangeClient, error) {
	creator, ok := f.creators[name]

	if !ok {
		return nil, fmt.Errorf("unsupported exchange: %s", name)
	}
	exchangeConfig, err := getExchangeConfig(name, f.cfg)
	if err != nil {
		return nil, err
	}
	return creator(exchangeConfig)
}
