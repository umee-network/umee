package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/config"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name      string
		cfg       config.Config
		expectErr bool
	}{
		{
			"valid config",
			config.Config{
				Server: config.Server{
					ListenAddr:     "0.0.0.0:7171",
					VerboseCORS:    false,
					AllowedOrigins: []string{},
				},
				CurrencyPairs: []config.CurrencyPair{
					{Base: "ATOM", Quote: "USDT", Providers: []string{"kraken"}},
				},
				Account: config.Account{
					Address:   "fromaddr",
					Validator: "valaddr",
					ChainID:   "chain-id",
				},
				Keyring: config.Keyring{
					Backend: "test",
					Dir:     "/Users/username/.umee",
				},
				RPC: config.RPC{
					TMRPCEndpoint: "http://localhost:26657",
					GRPCEndpoint:  "localhost:9090",
					RPCTimeout:    "100ms",
				},
				Telemetry: config.Telemetry{
					ServiceName:         "price-feeder",
					Enabled:             true,
					EnableHostname:      true,
					EnableHostnameLabel: true,
					EnableServiceLabel:  true,
					GlobalLabels:        make([][]string, 1),
					Type:                "generic",
				},
				GasAdjustment: 1.5,
			},
			false,
		},
		{
			"empty pairs",
			config.Config{
				Server: config.Server{
					ListenAddr:     "0.0.0.0:7171",
					VerboseCORS:    false,
					AllowedOrigins: []string{},
				},
				CurrencyPairs: []config.CurrencyPair{},
			},
			true,
		},
		{
			"invalid base",
			config.Config{
				Server: config.Server{
					ListenAddr:     "0.0.0.0:7171",
					VerboseCORS:    false,
					AllowedOrigins: []string{},
				},
				CurrencyPairs: []config.CurrencyPair{
					{Base: "", Quote: "USDT", Providers: []string{"kraken"}},
				},
			},
			true,
		},
		{
			"invalid quote",
			config.Config{
				Server: config.Server{
					ListenAddr:     "0.0.0.0:7171",
					VerboseCORS:    false,
					AllowedOrigins: []string{},
				},
				CurrencyPairs: []config.CurrencyPair{
					{Base: "ATOM", Quote: "", Providers: []string{"kraken"}},
				},
			},
			true,
		},
		{
			"empty providers",
			config.Config{
				Server: config.Server{
					ListenAddr:     "0.0.0.0:7171",
					VerboseCORS:    false,
					AllowedOrigins: []string{},
				},
				CurrencyPairs: []config.CurrencyPair{
					{Base: "ATOM", Quote: "USDT", Providers: []string{}},
				},
			},
			true,
		},
		{
			"invalid endpoints",
			config.Config{
				Server: config.Server{
					ListenAddr:     "0.0.0.0:7171",
					VerboseCORS:    false,
					AllowedOrigins: []string{},
				},
				CurrencyPairs: []config.CurrencyPair{
					{Base: "ATOM", Quote: "USDT", Providers: []string{"kraken"}},
				},
				Account: config.Account{
					Address:   "fromaddr",
					Validator: "valaddr",
					ChainID:   "chain-id",
				},
				Keyring: config.Keyring{
					Backend: "test",
					Dir:     "/Users/username/.umee",
				},
				RPC: config.RPC{
					TMRPCEndpoint: "http://localhost:26657",
					GRPCEndpoint:  "localhost:9090",
					RPCTimeout:    "100ms",
				},
				Telemetry: config.Telemetry{
					ServiceName:         "price-feeder",
					Enabled:             true,
					EnableHostname:      true,
					EnableHostnameLabel: true,
					EnableServiceLabel:  true,
					GlobalLabels:        make([][]string, 1),
					Type:                "generic",
				},
				GasAdjustment: 1.5,
				ProviderEndpoints: []config.ProviderEndpoint{
					{
						Name: "binance",
					},
				},
			},
			true,
		},
		{
			"invalid endpoint provider",
			config.Config{
				Server: config.Server{
					ListenAddr:     "0.0.0.0:7171",
					VerboseCORS:    false,
					AllowedOrigins: []string{},
				},
				CurrencyPairs: []config.CurrencyPair{
					{Base: "ATOM", Quote: "USDT", Providers: []string{"kraken"}},
				},
				Account: config.Account{
					Address:   "fromaddr",
					Validator: "valaddr",
					ChainID:   "chain-id",
				},
				Keyring: config.Keyring{
					Backend: "test",
					Dir:     "/Users/username/.umee",
				},
				RPC: config.RPC{
					TMRPCEndpoint: "http://localhost:26657",
					GRPCEndpoint:  "localhost:9090",
					RPCTimeout:    "100ms",
				},
				Telemetry: config.Telemetry{
					ServiceName:         "price-feeder",
					Enabled:             true,
					EnableHostname:      true,
					EnableHostnameLabel: true,
					EnableServiceLabel:  true,
					GlobalLabels:        make([][]string, 1),
					Type:                "generic",
				},
				GasAdjustment: 1.5,
				ProviderEndpoints: []config.ProviderEndpoint{
					{
						Name:      "foo",
						Rest:      "bar",
						Websocket: "baz",
					},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.cfg.Validate() != nil, tc.expectErr)
		})
	}
}

func TestParseConfig_Valid(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "price-feeder.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte(`
gas_adjustment = 1.5

[server]
listen_addr = "0.0.0.0:99999"
read_timeout = "20s"
verbose_cors = true
write_timeout = "20s"

[[currency_pairs]]
base = "ATOM"
quote = "USDT"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[[currency_pairs]]
base = "UMEE"
quote = "USDT"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[[currency_pairs]]
base = "USDT"
quote = "USD"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[account]
address = "umee15nejfgcaanqpw25ru4arvfd0fwy6j8clccvwx4"
validator = "umeevalcons14rjlkfzp56733j5l5nfk6fphjxymgf8mj04d5p"
chain_id = "umee-local-testnet"

[keyring]
backend = "test"
dir = "/Users/username/.umee"
pass = "keyringPassword"

[rpc]
tmrpc_endpoint = "http://localhost:26657"
grpc_endpoint = "localhost:9090"
rpc_timeout = "100ms"

[telemetry]
service_name = "price-feeder"
enabled = true
enable_hostname = true
enable_hostname_label = true
enable_service_label = true
type = "prometheus"
global_labels = [["chain-id", "umee-local-testnet"]]
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	cfg, err := config.ParseConfig(tmpFile.Name())
	require.NoError(t, err)

	require.Equal(t, "0.0.0.0:99999", cfg.Server.ListenAddr)
	require.Equal(t, "20s", cfg.Server.WriteTimeout)
	require.Equal(t, "20s", cfg.Server.ReadTimeout)
	require.True(t, cfg.Server.VerboseCORS)
	require.Len(t, cfg.CurrencyPairs, 3)
	require.Equal(t, "ATOM", cfg.CurrencyPairs[0].Base)
	require.Equal(t, "USDT", cfg.CurrencyPairs[0].Quote)
	require.Len(t, cfg.CurrencyPairs[0].Providers, 3)
	require.Equal(t, "kraken", cfg.CurrencyPairs[0].Providers[0])
	require.Equal(t, "binance", cfg.CurrencyPairs[0].Providers[1])
}

func TestParseConfig_Valid_NoTelemetry(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "price-feeder.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte(`
gas_adjustment = 1.5

[server]
listen_addr = "0.0.0.0:99999"
read_timeout = "20s"
verbose_cors = true
write_timeout = "20s"

[[currency_pairs]]
base = "ATOM"
quote = "USDT"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[[currency_pairs]]
base = "UMEE"
quote = "USDT"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[[currency_pairs]]
base = "USDT"
quote = "USD"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[account]
address = "umee15nejfgcaanqpw25ru4arvfd0fwy6j8clccvwx4"
validator = "umeevalcons14rjlkfzp56733j5l5nfk6fphjxymgf8mj04d5p"
chain_id = "umee-local-testnet"

[keyring]
backend = "test"
dir = "/Users/username/.umee"
pass = "keyringPassword"

[rpc]
tmrpc_endpoint = "http://localhost:26657"
grpc_endpoint = "localhost:9090"
rpc_timeout = "100ms"

[telemetry]
enabled = false
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	cfg, err := config.ParseConfig(tmpFile.Name())
	require.NoError(t, err)

	require.Equal(t, "0.0.0.0:99999", cfg.Server.ListenAddr)
	require.Equal(t, "20s", cfg.Server.WriteTimeout)
	require.Equal(t, "20s", cfg.Server.ReadTimeout)
	require.True(t, cfg.Server.VerboseCORS)
	require.Len(t, cfg.CurrencyPairs, 3)
	require.Equal(t, "ATOM", cfg.CurrencyPairs[0].Base)
	require.Equal(t, "USDT", cfg.CurrencyPairs[0].Quote)
	require.Len(t, cfg.CurrencyPairs[0].Providers, 3)
	require.Equal(t, "kraken", cfg.CurrencyPairs[0].Providers[0])
	require.Equal(t, "binance", cfg.CurrencyPairs[0].Providers[1])
	require.Equal(t, cfg.Telemetry.Enabled, false)
}

func TestParseConfig_InvalidProvider(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "price-feeder.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte(`
listen_addr = ""

[[currency_pairs]]
base = "ATOM"
quote = "USD"
providers = [
	"kraken",
	"binance"
]

[[currency_pairs]]
base = "UMEE"
quote = "USD"
providers = [
	"kraken",
	"foobar"
]
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	_, err = config.ParseConfig(tmpFile.Name())
	require.Error(t, err)
}

func TestParseConfig_NonUSDQuote(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "price-feeder.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte(`
listen_addr = ""

[[currency_pairs]]
base = "ATOM"
quote = "USDT"
providers = [
	"kraken",
	"binance"
]

[[currency_pairs]]
base = "UMEE"
quote = "USDT"
providers = [
	"kraken",
	"binance"
]
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	_, err = config.ParseConfig(tmpFile.Name())
	require.Error(t, err)
}

func TestParseConfig_Valid_Deviations(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "price-feeder.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte(`
gas_adjustment = 1.5

[server]
listen_addr = "0.0.0.0:99999"
read_timeout = "20s"
verbose_cors = true
write_timeout = "20s"

[[deviation_thresholds]]
base = "USDT"
threshold = "2"

[[deviation_thresholds]]
base = "ATOM"
threshold = "1.5"

[[currency_pairs]]
base = "ATOM"
quote = "USDT"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[[currency_pairs]]
base = "UMEE"
quote = "USDT"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[[currency_pairs]]
base = "USDT"
quote = "USD"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[account]
address = "umee15nejfgcaanqpw25ru4arvfd0fwy6j8clccvwx4"
validator = "umeevalcons14rjlkfzp56733j5l5nfk6fphjxymgf8mj04d5p"
chain_id = "umee-local-testnet"

[keyring]
backend = "test"
dir = "/Users/username/.umee"
pass = "keyringPassword"

[rpc]
tmrpc_endpoint = "http://localhost:26657"
grpc_endpoint = "localhost:9090"
rpc_timeout = "100ms"

[telemetry]
service_name = "price-feeder"
enabled = true
enable_hostname = true
enable_hostname_label = true
enable_service_label = true
type = "prometheus"
global_labels = [["chain-id", "umee-local-testnet"]]
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	cfg, err := config.ParseConfig(tmpFile.Name())
	require.NoError(t, err)

	require.Equal(t, "0.0.0.0:99999", cfg.Server.ListenAddr)
	require.Equal(t, "20s", cfg.Server.WriteTimeout)
	require.Equal(t, "20s", cfg.Server.ReadTimeout)
	require.True(t, cfg.Server.VerboseCORS)
	require.Len(t, cfg.CurrencyPairs, 3)
	require.Equal(t, "ATOM", cfg.CurrencyPairs[0].Base)
	require.Equal(t, "USDT", cfg.CurrencyPairs[0].Quote)
	require.Len(t, cfg.CurrencyPairs[0].Providers, 3)
	require.Equal(t, "kraken", cfg.CurrencyPairs[0].Providers[0])
	require.Equal(t, "binance", cfg.CurrencyPairs[0].Providers[1])
	require.Equal(t, "2", cfg.Deviations[0].Threshold)
	require.Equal(t, "USDT", cfg.Deviations[0].Base)
	require.Equal(t, "1.5", cfg.Deviations[1].Threshold)
	require.Equal(t, "ATOM", cfg.Deviations[1].Base)
}

func TestParseConfig_Invalid_Deviations(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "price-feeder.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte(`
gas_adjustment = 1.5

[server]
listen_addr = "0.0.0.0:99999"
read_timeout = "20s"
verbose_cors = true
write_timeout = "20s"

[[deviation_thresholds]]
base = "USDT"
threshold = "4.0"

[[deviation_thresholds]]
base = "ATOM"
threshold = "1.5"

[[currency_pairs]]
base = "ATOM"
quote = "USDT"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[[currency_pairs]]
base = "UMEE"
quote = "USDT"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[[currency_pairs]]
base = "USDT"
quote = "USD"
providers = [
	"kraken",
	"binance",
	"huobi"
]

[account]
address = "umee15nejfgcaanqpw25ru4arvfd0fwy6j8clccvwx4"
validator = "umeevalcons14rjlkfzp56733j5l5nfk6fphjxymgf8mj04d5p"
chain_id = "umee-local-testnet"

[keyring]
backend = "test"
dir = "/Users/username/.umee"
pass = "keyringPassword"

[rpc]
tmrpc_endpoint = "http://localhost:26657"
grpc_endpoint = "localhost:9090"
rpc_timeout = "100ms"

[telemetry]
service_name = "price-feeder"
enabled = true
enable_hostname = true
enable_hostname_label = true
enable_service_label = true
type = "prometheus"
global_labels = [["chain-id", "umee-local-testnet"]]
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	_, err = config.ParseConfig(tmpFile.Name())
	require.Error(t, err)
}
