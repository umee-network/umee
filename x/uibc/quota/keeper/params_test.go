package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/v6/x/uibc"
)

func TestUnitParams(t *testing.T) {
	require := require.New(t)
	k := initKeeperSimpleMock(t).Keeper

	// unit test doesn't setup params, so we should get zeroParams at the beginning
	params := k.GetParams()
	zeroParams := uibc.Params{}
	require.Equal(zeroParams, params)
	// update params
	params.IbcStatus = uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_DISABLED
	params.TokenQuota = sdk.MustNewDecFromStr("12.23")
	params.TotalQuota = sdk.MustNewDecFromStr("3.4321")
	err := k.SetParams(params)
	require.NoError(err)
	// check the updated params
	params2 := k.GetParams()
	require.Equal(params, params2)
}

func TestValidateEmergencyQuotaParamsUpdate(t *testing.T) {
	mkParams := func(total, token int64, duration time.Duration) uibc.Params {
		return uibc.Params{
			TotalQuota:    sdk.NewDec(total),
			TokenQuota:    sdk.NewDec(token),
			QuotaDuration: duration,
		}
	}

	p := mkParams(100, 10, 50)
	tcs := []struct {
		name   string
		p      uibc.Params
		errMsg string
	}{
		{"no change", p, ""},
		{"valid total quota update", mkParams(99, 10, 50), ""},
		{"valid update", mkParams(0, 0, 50), ""},

		{"invalid  update", mkParams(201, 11, 50), "can't increase"},
		{"invalid total quota update", mkParams(101, 10, 50), "can't increase"},
		{"invalid token quota update", mkParams(10, 12, 50), "can't increase"},
		{"invalid quota duration update1", mkParams(100, 10, 51), "can't change QuotaDuration"},
		{"invalid quota duration update2", mkParams(100, 10, 49), "can't change QuotaDuration"},
	}

	assert := assert.New(t)
	for _, tc := range tcs {
		err := validateEmergencyQuotaParamsUpdate(p, tc.p)
		if tc.errMsg == "" {
			assert.NoError(err, tc.name)
		} else {
			assert.ErrorContains(err, tc.errMsg, tc.name)
		}
	}
}
