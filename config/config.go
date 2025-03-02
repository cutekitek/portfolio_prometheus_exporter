package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ExchangeConfig struct {
	MetricsEndpoint  string            `yaml:"metrics_endpoint"`
	ScrapingInterval time.Duration     `yaml:"interval"`
	Params           map[string]string `yaml:"params"`
}

type Config struct {
	Host      string           `yaml:"host"`
	Port      int              `yaml:"port"`
	Exchanges map[string]ExchangeConfig `yaml:"exchanges"`

}

func NewDefaultConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 80,
	}
}

func ConfigFromYaml(filename string) (*Config, error) {
	cfg := NewDefaultConfig()
	file, err := os.Open(filename)
	if err != nil{
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil{
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return cfg, nil
}

func (c *Config) ListenAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}