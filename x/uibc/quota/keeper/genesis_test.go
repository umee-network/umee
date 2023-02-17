package keeper_test

import (
	"testing"

	"github.com/umee-network/umee/v4/x/uibc"
	"gotest.tools/v3/assert"
)

func TestInitGenesis(t *testing.T) {
	s := initKeeperTestSuite(t)
	app, ctx := s.app, s.ctx

	defaultGs := uibc.DefaultGenesisState()
	app.UIbcQuotaKeeper.InitGenesis(ctx, *defaultGs)
	params := app.UIbcQuotaKeeper.GetParams(ctx)
	assert.Equal(t, params.IbcPause, defaultGs.Params.IbcPause)
}

func TestExportGenesis(t *testing.T) {
	s := initKeeperTestSuite(t)
	app, ctx := s.app, s.ctx
	defaultGs := uibc.DefaultGenesisState()
	app.UIbcQuotaKeeper.InitGenesis(ctx, *defaultGs)
	eGs := app.UIbcQuotaKeeper.ExportGenesis(ctx)
	assert.DeepEqual(t, eGs, defaultGs)
	// update params in genesis state
	defaultGs.Params.IbcPause = uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_ENABLED
	app.UIbcQuotaKeeper.InitGenesis(ctx, *defaultGs)
	eGs = app.UIbcQuotaKeeper.ExportGenesis(ctx)
	assert.Equal(t, eGs.Params.IbcPause, uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_ENABLED)
}
