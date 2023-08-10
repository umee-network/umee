package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
