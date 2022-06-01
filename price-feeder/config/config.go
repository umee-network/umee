package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

const (
	DenomUSD = "USD"

	defaultListenAddr      = "0.0.0.0:7171"
	defaultSrvWriteTimeout = 15 * time.Second
	defaultSrvReadTimeout  = 15 * time.Second
	defaultProviderTimeout = 100 * time.Millisecond

	ProviderKraken   = "kraken"
	ProviderBinance  = "binance"
	ProviderOsmosis  = "osmosis"
	ProviderHuobi    = "huobi"
	ProviderOkx      = "okx"
	ProviderGate     = "gate"
	ProviderCoinbase = "coinbase"
	ProviderMock     = "mock"
)

var (
	validate *validator.Validate = validator.New()

	// ErrEmptyConfigPath defines a sentinel error for an empty config path.
	ErrEmptyConfigPath = errors.New("empty configuration file path")

	// SupportedProviders defines a lookup table of all the supported currency API
	// providers.
	SupportedProviders = map[string]struct{}{
		ProviderKraken:   {},
		ProviderBinance:  {},
		ProviderOsmosis:  {},
		ProviderOkx:      {},
		ProviderHuobi:    {},
		ProviderGate:     {},
		ProviderCoinbase: {},
		ProviderMock:     {},
	}
)

type (
	// Config defines all necessary price-feeder configuration parameters.
	Config struct {
		Server          Server         `toml:"server"`
		CurrencyPairs   []CurrencyPair `toml:"currency_pairs" validate:"required,gt=0,dive,required"`
		Deviations      []Deviation    `toml:"deviation_thresholds"`
		Account         Account        `toml:"account" validate:"required,gt=0,dive,required"`
		Keyring         Keyring        `toml:"keyring" validate:"required,gt=0,dive,required"`
		RPC             RPC            `toml:"rpc" validate:"required,gt=0,dive,required"`
		Telemetry       Telemetry      `toml:"telemetry"`
		GasAdjustment   float64        `toml:"gas_adjustment" validate:"required"`
		ProviderTimeout string         `toml:"provider_timeout"`
	}

	// Server defines the API server configuration.
	Server struct {
		ListenAddr     string   `toml:"listen_addr"`
		WriteTimeout   string   `toml:"write_timeout"`
		ReadTimeout    string   `toml:"read_timeout"`
		VerboseCORS    bool     `toml:"verbose_cors"`
		AllowedOrigins []string `toml:"allowed_origins"`
	}

	// CurrencyPair defines a price quote of the exchange rate for two different
	// currencies and the supported providers for getting the exchange rate.
	CurrencyPair struct {
		Base      string   `toml:"base" validate:"required"`
		Quote     string   `toml:"quote" validate:"required"`
		Providers []string `toml:"providers" validate:"required,gt=0,dive,required"`
	}

	// Deviation defines a maximum amount of standard deviations that a given asset can
	// be from the median without being filtered out before voting.
	Deviation struct {
		Base      string `toml:"base" validate:"required"`
		Threshold string `toml:"threshold" validate:"required"`
	}

	// Account defines account related configuration that is related to the Umee
	// network and transaction signing functionality.
	Account struct {
		ChainID   string `toml:"chain_id" validate:"required"`
		Address   string `toml:"address" validate:"required"`
		Validator string `toml:"validator" validate:"required"`
	}

	// Keyring defines the required Umee keyring configuration.
	Keyring struct {
		Backend string `toml:"backend" validate:"required"`
		Dir     string `toml:"dir" validate:"required"`
	}

	// RPC defines RPC configuration of both the Umee gRPC and Tendermint nodes.
	RPC struct {
		TMRPCEndpoint string `toml:"tmrpc_endpoint" validate:"required"`
		GRPCEndpoint  string `toml:"grpc_endpoint" validate:"required"`
		RPCTimeout    string `toml:"rpc_timeout" validate:"required"`
	}

	// Telemetry defines the configuration options for application telemetry.
	Telemetry struct {
		// Prefixed with keys to separate services
		ServiceName string `toml:"service_name"`

		// Enabled enables the application telemetry functionality. When enabled,
		// an in-memory sink is also enabled by default. Operators may also enabled
		// other sinks such as Prometheus.
		Enabled bool `toml:"enabled"`

		// Enable prefixing gauge values with hostname
		EnableHostname bool `toml:"enable_hostname"`

		// Enable adding hostname to labels
		EnableHostnameLabel bool `toml:"enable_hostname_label"`

		// Enable adding service to labels
		EnableServiceLabel bool `toml:"enable_service_label"`

		// GlobalLabels defines a global set of name/value label tuples applied to all
		// metrics emitted using the wrapper functions defined in telemetry package.
		//
		// Example:
		// [["chain_id", "cosmoshub-1"]]
		GlobalLabels [][]string `toml:"global_labels""`

		// Type determines which type of telemetry to use
		// Valid values are "prometheus" or "generic"
		Type string `toml:"type"`
	}
)

// telemetryValidation is custom validation for the Telemetry struct
func telemetryValidation(sl validator.StructLevel) {
	tel := sl.Current().Interface().(Telemetry)

	if tel.Enabled && (len(tel.GlobalLabels) == 0 || len(tel.ServiceName) == 0 ||
		len(tel.Type) == 0) {
		sl.ReportError(tel.Enabled, "enabled", "Enabled", "enabledNoOptions", "")
	} else if tel.Enabled &&
		(len(tel.Type) == 0 ||
			(tel.Type != "prometheus" && tel.Type != "generic")) {
		sl.ReportError(tel.Enabled, "type", "Type", "unsupportedTelemetryType", "")
	}
}

// Validate returns an error if the Config object is invalid.
func (c Config) Validate() error {
	validate.RegisterStructValidation(telemetryValidation, Telemetry{})
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

	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = defaultListenAddr
	}
	if len(cfg.Server.WriteTimeout) == 0 {
		cfg.Server.WriteTimeout = defaultSrvWriteTimeout.String()
	}
	if len(cfg.Server.ReadTimeout) == 0 {
		cfg.Server.ReadTimeout = defaultSrvReadTimeout.String()
	}
	if len(cfg.ProviderTimeout) == 0 {
		cfg.ProviderTimeout = defaultProviderTimeout.String()
	}

	pairs := make(map[string]map[string]struct{})
	coinQuotes := make(map[string]struct{})
	for _, cp := range cfg.CurrencyPairs {
		if _, ok := pairs[cp.Base]; !ok {
			pairs[cp.Base] = make(map[string]struct{})
		}
		if strings.ToUpper(cp.Quote) != DenomUSD {
			coinQuotes[cp.Quote] = struct{}{}
		}

		for _, provider := range cp.Providers {
			if _, ok := SupportedProviders[provider]; !ok {
				return cfg, fmt.Errorf("unsupported provider: %s", provider)
			}
			pairs[cp.Base][provider] = struct{}{}
		}
	}

	// Use coinQuotes to ensure that any quotes can be converted to USD.
	for quote := range coinQuotes {
		for index, pair := range cfg.CurrencyPairs {
			if pair.Base == quote && pair.Quote == DenomUSD {
				break
			}
			if index == len(cfg.CurrencyPairs)-1 {
				return cfg, fmt.Errorf("all non-usd quotes require a conversion rate feed")
			}
		}
	}

	for base, providers := range pairs {
		if _, ok := pairs[base]["mock"]; !ok && len(providers) < 3 {
			return cfg, fmt.Errorf("must have at least three providers for %s", base)
		}
	}

	gatePairs := []string{}
	for base, providers := range pairs {
		if _, ok := providers["gate"]; ok {
			gatePairs = append(gatePairs, base)
		}
	}
	if len(gatePairs) > 1 {
		return cfg, fmt.Errorf("gate provider does not support multiple pairs: %v", gatePairs)
	}

	return cfg, cfg.Validate()
}
