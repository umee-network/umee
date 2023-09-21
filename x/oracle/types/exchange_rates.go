package types

import (
	"encoding/json"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewExchangeRate creates a ExchangeRate instance
func NewExchangeRate(denom string, exchangeRate sdk.Dec, t time.Time) ExchangeRate {
	return ExchangeRate{
		Denom:     denom,
		Rate:      exchangeRate,
		Timestamp: t,
	}
}

// DenomExchangeRate is type for storing rate and timestamp of denom into store without denom.
type DenomExchangeRate struct {
	Rate      string    `json:"rate"`
	Timestamp time.Time `json:"timestamp"`
}

// Marshal implements store.Marshalable.
func (d *DenomExchangeRate) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

// MarshalTo implements store.Marshalable.
func (DenomExchangeRate) MarshalTo(_ []byte) (int, error) {
	panic("unimplemented")
}

// Unmarshal implements store.Marshalable.
func (d *DenomExchangeRate) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, d)
	return err
}
