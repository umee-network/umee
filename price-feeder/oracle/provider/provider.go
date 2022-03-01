package provider

import (
	"fmt"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	defaultTimeout           = 10 * time.Second
	defaultReadNewWSMessage  = 50 * time.Millisecond
	defaultMaxConnectionTime = time.Hour * 23 // should be < 24h
	defaultReconnectTime     = time.Minute * 20
	maxReconnectionTries     = 3
)

var ping = []byte("ping")

// Provider defines an interface an exchange price provider must implement.
type Provider interface {
	GetTickerPrices(...types.CurrencyPair) (map[string]TickerPrice, error)
}

// TickerPrice defines price and volume information for a symbol or ticker
// exchange rate.
type TickerPrice struct {
	Price  sdk.Dec // last trade price
	Volume sdk.Dec // 24h volume
}

// preventRedirect avoid any redirect in the http.Client the request call
// will not return an error, but a valid response with redirect response code.
func preventRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func newDefaultHTTPClient() *http.Client {
	return newHTTPClientWithTimeout(defaultTimeout)
}

func newHTTPClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: preventRedirect,
	}
}

func newTickerPrice(provider, symbol, lastPrice, volume string) (TickerPrice, error) {
	price, err := sdk.NewDecFromStr(lastPrice)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse %s price (%s) for %s", provider, lastPrice, symbol)
	}

	volumeDec, err := sdk.NewDecFromStr(volume)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse %s volume (%s) for %s", provider, volume, symbol)
	}

	return TickerPrice{Price: price, Volume: volumeDec}, nil
}
