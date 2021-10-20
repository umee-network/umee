package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate = validator.New()

	// ErrEmptyConfigPath defines a sentinel error for an empty config path.
	ErrEmptyConfigPath = errors.New("empty configuration file path")
)

var (
	defaultListenAddr      = "0.0.0.0:7171"
	defaultSrvWriteTimeout = 15 * time.Second
	defaultSrvReadTimeout  = 15 * time.Second

	supportedProviders = map[string]struct{}{
		"kraken":   {},
		"bitfinex": {},
	}
)

type (
	// Config defines all necessary price-feeder configuration parameters.
	Config struct {
		ListenAddr         string         `toml:"listen_addr"`
		ServerWriteTimeout string         `toml:"server_write_timeout"`
		ServerReadTimeout  string         `toml:"server_read_timeout"`
		CurrencyPairs      []CurrencyPair `toml:"currency_pairs" validate:"required,gt=0,dive,required"`
	}

	// CurrencyPair defines a price quote of the exchange rate for two different
	// currencies and the supported providers for getting the exchange rate.
	CurrencyPair struct {
		Base      string   `toml:"base" validate:"required"`
		Quote     string   `toml:"quote" validate:"required"`
		Providers []string `toml:"providers" validate:"required,gt=0,dive,required"`
	}
)

// Validate returns an error if the Config object is invalid.
func (c Config) Validate() error {
	return validate.Struct(c)
}

// ParseConfig attempts to read and parse configuration from the given file path.
// An error is returned if reading or parsing the config fails.
func ParseConfig(configPath string) (Config, error) {
	var cfg Config

	if configPath == "" {
		return cfg, ErrEmptyConfigPath
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}

	if _, err := toml.Decode(string(configData), &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = defaultListenAddr
	}
	if len(cfg.ServerWriteTimeout) == 0 {
		cfg.ServerWriteTimeout = defaultSrvWriteTimeout.String()
	}
	if len(cfg.ServerReadTimeout) == 0 {
		cfg.ServerReadTimeout = defaultSrvReadTimeout.String()
	}

	for _, cp := range cfg.CurrencyPairs {
		for _, provider := range cp.Providers {
			if _, ok := supportedProviders[provider]; !ok {
				return cfg, fmt.Errorf("unsupported provider: %s", provider)
			}
		}
	}

	return cfg, cfg.Validate()
}
