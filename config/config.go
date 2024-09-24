package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const STREAM_UPDATE_INTERVAL = 1 * time.Second

type Server struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

type Database struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type Exchange struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:"config"`
}

type Symbol struct {
	Name     string `yaml:"name"`
	Symbol   string `yaml:"symbol"`
	Exchange string `yaml:"exchange"`
	Active   bool   `yaml:"active"`
}

type Config struct {
	Server    Server     `yaml:"server"`
	Database  Database   `yaml:"database"`
	Exchanges []Exchange `yaml:"exchanges"`
	Symbols   []Symbol   `yaml:"symbols"`
}

func LoadFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
