package types

import (
	"encoding/json"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewDenomExchangeRate creates a DenomExchangeRate instance
func NewDenomExchangeRate(denom string, exchangeRate sdk.Dec, t time.Time) DenomExchangeRate {
	return DenomExchangeRate{
		Denom:     denom,
		Rate:      exchangeRate,
		Timestamp: t,
	}
}

func (v DenomExchangeRate) String() string {
	bz, _ := json.Marshal(v)
	return string(bz)
}

// ExchangeRate is type for storing rate and timestamp of denom into store without denom.
type ExchangeRate struct {
	Rate      sdk.Dec   `json:"rate"`
	Timestamp time.Time `json:"timestamp"`
}

// Marshal implements store.Marshalable.
func (d *ExchangeRate) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

func (d *ExchangeRate) String() string {
	out, _ := json.Marshal(d)
	return string(out)
}

// MarshalTo implements store.Marshalable.
func (ExchangeRate) MarshalTo(_ []byte) (int, error) {
	panic("unimplemented")
}

// Unmarshal implements store.Marshalable.
func (d *ExchangeRate) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, d)
	return err
}
