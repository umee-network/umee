package types

import (
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestExchangeRateMarshalAndUnmarshal(t *testing.T) {
	der := ExchangeRate{Rate: sdk.NewDec(1), Timestamp: time.Now()}

	// Marshal the exchange rate
	md, err := der.Marshal()
	assert.NilError(t, err)

	// Unmarshal the exchange rate
	var newExgRate ExchangeRate
	err = newExgRate.Unmarshal(md)
	assert.NilError(t, err)
	assert.DeepEqual(t, der, newExgRate)

	// error expected
	der = ExchangeRate{Timestamp: time.Now()}

	// Marshal the exchange rate
	_, err = der.Marshal()
	assert.ErrorContains(t, err, "rate should not be nil")

	// error expected
	der = ExchangeRate{Rate: sdk.NewDec(1)}

	// Marshal the exchange rate
	_, err = der.Marshal()
	assert.ErrorContains(t, err, "timestamp in exchange_rate should not be nil")
}
