package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestAggregateExchangeRatePrevoteString(t *testing.T) {
	addr := sdk.ValAddress(sdk.AccAddress([]byte("addr1_______________")))
	aggregateVoteHash := GetAggregateVoteHash("salt", "UMEE:100,ATOM:100", addr)
	aggregateExchangeRatePreVote := NewAggregateExchangeRatePrevote(
		aggregateVoteHash,
		addr,
		100,
	)

	assert.Equal(t, "hash: 19c30cf9ea8aa0e0b03904162cadec0f2024a76d\nvoter: umeevaloper1v9jxgu33ta047h6lta047h6lta047h6l5ltnvg\nsubmit_block: 100\n", aggregateExchangeRatePreVote.String())
}

func TestAggregateExchangeRateVoteString(t *testing.T) {
	aggregateExchangeRatePreVote := NewAggregateExchangeRateVote(
		ExchangeRateTuples{
			NewExchangeRateTuple(UmeeDenom, sdk.OneDec()),
		},
		sdk.ValAddress(sdk.AccAddress([]byte("addr1_______________"))),
	)

	assert.Equal(t, "exchange_rate_tuples:\n    - denom: uumee\n      exchange_rate: \"1.000000000000000000\"\nvoter: umeevaloper1v9jxgu33ta047h6lta047h6lta047h6l5ltnvg\n", aggregateExchangeRatePreVote.String())
}

func TestExchangeRateTuplesString(t *testing.T) {
	t.Parallel()
	exchangeRateTuple := NewExchangeRateTuple(UmeeDenom, sdk.OneDec())
	assert.Equal(t, exchangeRateTuple.String(), `{"denom":"uumee", "exchange_rate":"1"}`)

	exchangeRateTuples := ExchangeRateTuples{
		exchangeRateTuple,
		NewExchangeRateTuple(IbcDenomAtom, sdk.SmallestDec()),
	}
	assert.Equal(t, `[{"denom":"uumee","exchange_rate":"1"},{"denom":"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2","exchange_rate":"0.000000000000000001"}]`, exchangeRateTuples.String())
}

func TestParseExchangeRateTuples(t *testing.T) {
	valid := "uumee:123.0,uatom:123.123"
	_, err := ParseExchangeRateTuples(valid)
	assert.NoError(t, err)

	duplicatedDenom := "uumee:100.0,uatom:123.123,uatom:121233.123"
	_, err = ParseExchangeRateTuples(duplicatedDenom)
	assert.ErrorContains(t, err, "duplicated denom UATOM: invalid coins")

	invalidCoins := "123.123"
	_, err = ParseExchangeRateTuples(invalidCoins)
	assert.ErrorContains(t, err, "invalid exchange rate")

	invalidCoinsWithValid := "uumee:123.0,123.1"
	_, err = ParseExchangeRateTuples(invalidCoinsWithValid)
	assert.ErrorContains(t, err, "invalid exchange rate")

	zeroCoinsWithValid := "uumee:0.0,uatom:123.1"
	_, err = ParseExchangeRateTuples(zeroCoinsWithValid)
	assert.ErrorContains(t, err, "can't be negative")

	negativeCoinsWithValid := "uumee:-1234.5,uatom:123.1"
	_, err = ParseExchangeRateTuples(negativeCoinsWithValid)
	assert.ErrorContains(t, err, "can't be negative")

	multiplePricesPerRate := "uumee:123: uumee:456,uusdc:789"
	_, err = ParseExchangeRateTuples(multiplePricesPerRate)
	assert.ErrorContains(t, err, "invalid exchange rate")

	_, err = ParseExchangeRateTuples("")
	assert.NoError(t, err)
}

func TestExchangeRateString(t *testing.T) {
	t1 := time.Date(2022, 9, 18, 15, 55, 01, 0, time.UTC)
	er := ExchangeRate{Rate: sdk.MustNewDecFromStr("1.5"), Timestamp: t1}
	assert.Equal(t, `{"rate":"1.500000000000000000","timestamp":"2022-09-18T15:55:01Z"}`, er.String())
}
