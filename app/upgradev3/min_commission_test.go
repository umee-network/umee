package upgradev3

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestUpdateMinimumCommissionRateParam(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	_, sk := GenerateRandomTestCase()

	// check default min commission rate
	oldParams := sk.GetParams(ctx)
	require.Equal(t, types.DefaultMinCommissionRate, oldParams.MinCommissionRate)

	//  update the min commission rate
	_, err := UpdateMinimumCommissionRateParam(ctx, &sk)
	require.NoError(t, err)

	// get the updated params
	updatedParams := sk.GetParams(ctx)
	require.Equal(t, minCommissionRate, updatedParams.MinCommissionRate)
}

func TestSetMinimumCommissionRateToValidators(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	valAddrs, sk := GenerateRandomTestCase()

	// update the min commission rate
	minCommissionRate, err := UpdateMinimumCommissionRateParam(ctx, &sk)
	require.NoError(t, err)
	require.NotNil(t, minCommissionRate)

	// update min commisson rate to all validators
	err = SetMinimumCommissionRateToValidators(ctx, &sk, minCommissionRate)
	require.NoError(t, err)

	// get the validator
	validator, found := sk.GetValidator(ctx, valAddrs[0])
	require.True(t, found)
	require.True(t, minCommissionRate.Equal(validator.Commission.Rate))

	validator2, found := sk.GetValidator(ctx, valAddrs[1])
	require.True(t, found)
	// validator2 commission rate should be greater than minCommissionRate
	require.True(t, minCommissionRate.LT(validator2.Commission.Rate))
}
