package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	core "github.com/umee-network/umee/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestExchangeRate(t *testing.T) {
	input := CreateTestInput(t)

	cnyExchangeRate := sdk.NewDecWithPrec(839, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)
	gbpExchangeRate := sdk.NewDecWithPrec(4995, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)
	krwExchangeRate := sdk.NewDecWithPrec(2838, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)
	lunaExchangeRate := sdk.NewDecWithPrec(3282384, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)

	// Set & get rates
	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroCNYDenom, cnyExchangeRate)
	rate, err := input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroCNYDenom)
	require.NoError(t, err)
	require.Equal(t, cnyExchangeRate, rate)

	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroGBPDenom, gbpExchangeRate)
	rate, err = input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroGBPDenom)
	require.NoError(t, err)
	require.Equal(t, gbpExchangeRate, rate)

	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroKRWDenom, krwExchangeRate)
	rate, err = input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroKRWDenom)
	require.NoError(t, err)
	require.Equal(t, krwExchangeRate, rate)

	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroLunaDenom, lunaExchangeRate)
	rate, _ = input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroLunaDenom)
	require.Equal(t, sdk.OneDec(), rate)

	input.OracleKeeper.DeleteLunaExchangeRate(input.Ctx, core.MicroKRWDenom)
	_, err = input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroKRWDenom)
	require.Error(t, err)

	numExchangeRates := 0
	handler := func(denom string, exchangeRate sdk.Dec) (stop bool) {
		numExchangeRates = numExchangeRates + 1
		return false
	}
	input.OracleKeeper.IterateLunaExchangeRates(input.Ctx, handler)

	require.True(t, numExchangeRates == 3)
}
