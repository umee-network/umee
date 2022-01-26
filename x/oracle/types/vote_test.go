package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseExchangeRateTuples(t *testing.T) {
	valid := "uumee:123.0,uatom:123.123"
	_, err := ParseExchangeRateTuples(valid)
	require.NoError(t, err)

	duplicatedDenom := "uumee:100.0,uatom:123.123,uatom:121233.123"
	_, err = ParseExchangeRateTuples(duplicatedDenom)
	require.Error(t, err)

	invalidCoins := "123.123"
	_, err = ParseExchangeRateTuples(invalidCoins)
	require.Error(t, err)

	invalidCoinsWithValid := "uumee:123.0,123.1"
	_, err = ParseExchangeRateTuples(invalidCoinsWithValid)
	require.Error(t, err)

	zeroCoinsWithValid := "uumee:0.0,uatom:123.1"
	_, err = ParseExchangeRateTuples(zeroCoinsWithValid)
	require.Error(t, err)

	negativeCoinsWithValid := "uumee:-1234.5,uatom:123.1"
	_, err = ParseExchangeRateTuples(negativeCoinsWithValid)
	require.Error(t, err)

	multiplePricesPerRate := "uumee:123: uumee:456,uusdc:789"
	_, err = ParseExchangeRateTuples(multiplePricesPerRate)
	require.Error(t, err)
}
