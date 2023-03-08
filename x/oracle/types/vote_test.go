package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
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
	exchangeRateTuple := NewExchangeRateTuple(UmeeDenom, sdk.OneDec())
	assert.Equal(t, exchangeRateTuple.String(), "denom: uumee\nexchange_rate: \"1.000000000000000000\"\n")

	exchangeRateTuples := ExchangeRateTuples{
		exchangeRateTuple,
		NewExchangeRateTuple(IbcDenomAtom, sdk.SmallestDec()),
	}
	assert.Equal(t, "- denom: uumee\n  exchange_rate: \"1.000000000000000000\"\n- denom: ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2\n  exchange_rate: \"0.000000000000000001\"\n", exchangeRateTuples.String())
}

func TestParseExchangeRateTuples(t *testing.T) {
	errMsg := "invalid oracle price"

	valid := "uumee:123.0,uatom:123.123"
	_, err := ParseExchangeRateTuples(valid)
	assert.NilError(t, err)

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
	assert.ErrorContains(t, err, errMsg)

	negativeCoinsWithValid := "uumee:-1234.5,uatom:123.1"
	_, err = ParseExchangeRateTuples(negativeCoinsWithValid)
	assert.ErrorContains(t, err, errMsg)

	multiplePricesPerRate := "uumee:123: uumee:456,uusdc:789"
	_, err = ParseExchangeRateTuples(multiplePricesPerRate)
	assert.ErrorContains(t, err, "invalid exchange rate")

	_, err = ParseExchangeRateTuples("")
	assert.NilError(t, err)
}
