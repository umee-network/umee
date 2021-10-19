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
				ListenAddr: "0.0.0.0:7171",
				CurrencyPairs: []config.CurrencyPair{
					{Base: "ATOM", Quote: "USDT", Providers: []string{"kraken"}},
				},
			},
			false,
		},
		{
			"empty pairs",
			config.Config{
				ListenAddr:    "0.0.0.0:7171",
				CurrencyPairs: []config.CurrencyPair{},
			},
			true,
		},
		{
			"invalid base",
			config.Config{
				ListenAddr: "0.0.0.0:7171",
				CurrencyPairs: []config.CurrencyPair{
					{Base: "", Quote: "USDT", Providers: []string{"kraken"}},
				},
			},
			true,
		},
		{
			"invalid quote",
			config.Config{
				ListenAddr: "0.0.0.0:7171",
				CurrencyPairs: []config.CurrencyPair{
					{Base: "ATOM", Quote: "", Providers: []string{"kraken"}},
				},
			},
			true,
		},
		{
			"empty providers",
			config.Config{
				ListenAddr: "0.0.0.0:7171",
				CurrencyPairs: []config.CurrencyPair{
					{Base: "ATOM", Quote: "USDT", Providers: []string{}},
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
listen_addr = ""

[[currency_pairs]]
base = "ATOM"
quote = "USDT"
providers = [
	"kraken",
	"bitfinex"
]

[[currency_pairs]]
base = "UMEE"
quote = "USDT"
providers = [
	"kraken",
	"bitfinex"
]
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	cfg, err := config.ParseConfig(tmpFile.Name())
	require.NoError(t, err)

	require.Equal(t, "0.0.0.0:7171", cfg.ListenAddr)
	require.Len(t, cfg.CurrencyPairs, 2)
	require.Equal(t, "ATOM", cfg.CurrencyPairs[0].Base)
	require.Equal(t, "USDT", cfg.CurrencyPairs[0].Quote)
	require.Len(t, cfg.CurrencyPairs[0].Providers, 2)
	require.Equal(t, "kraken", cfg.CurrencyPairs[0].Providers[0])
	require.Equal(t, "bitfinex", cfg.CurrencyPairs[0].Providers[1])
}

func TestParseConfig_InvalidProvider(t *testing.T) {
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
	"bitfinex"
]

[[currency_pairs]]
base = "UMEE"
quote = "USDT"
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
