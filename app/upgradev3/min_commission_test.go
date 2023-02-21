package upgradev3

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"
)

func TestUpdateMinimumCommissionRateParam(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	_, sk := GenerateRandomTestCase()

	// check default min commission rate
	oldParams := sk.GetParams(ctx)
	assert.Equal(t, types.DefaultMinCommissionRate, oldParams.MinCommissionRate)

	//  update the min commission rate
	_, err := UpdateMinimumCommissionRateParam(ctx, &sk)
	assert.NilError(t, err)

	// get the updated params
	updatedParams := sk.GetParams(ctx)
	assert.Equal(t, minCommissionRate, updatedParams.MinCommissionRate)
}

func TestSetMinimumCommissionRateToValidators(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	valAddrs, sk := GenerateRandomTestCase()

	// update the min commission rate
	minCommissionRate, err := UpdateMinimumCommissionRateParam(ctx, &sk)
	assert.NilError(t, err)

	// update min commisson rate to all validators
	err = SetMinimumCommissionRateToValidators(ctx, &sk, minCommissionRate)
	assert.NilError(t, err)

	// get the validator
	validator, found := sk.GetValidator(ctx, valAddrs[0])
	assert.Equal(t, true, found)
	assert.Equal(t, true, minCommissionRate.Equal(validator.Commission.Rate))

	validator2, found := sk.GetValidator(ctx, valAddrs[1])
	assert.Equal(t, true, found)
	// validator2 commission rate should be greater than minCommissionRate
	assert.Equal(t, true, minCommissionRate.LT(validator2.Commission.Rate))
}
