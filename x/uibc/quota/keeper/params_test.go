package keeper_test

import (
	"testing"

	"github.com/umee-network/umee/v4/x/uibc"
	"gotest.tools/v3/assert"
)

func TestParams(t *testing.T) {
	s := initKeeperTestSuite(t)
	app, ctx := s.app, s.ctx
	params := app.UIbcQuotaKeeper.GetParams(ctx)
	defaultParams := uibc.DefaultParams()
	assert.DeepEqual(t, params, defaultParams)
	// update params
	params.IbcPause = uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_ENABLED
	err := app.UIbcQuotaKeeper.SetParams(ctx, params)
	assert.NilError(t, err)
	// check the update param
	params = app.UIbcQuotaKeeper.GetParams(ctx)
	assert.Equal(t, params.IbcPause, uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_ENABLED)
}
