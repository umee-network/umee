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
	assert.Equal(t, uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_ENABLED, defaultGs.Params.IbcStatus)
	app.UIbcQuotaKeeper.InitGenesis(ctx, *defaultGs)
	params := app.UIbcQuotaKeeper.GetParams(ctx)
	assert.Equal(t, params.IbcStatus, defaultGs.Params.IbcStatus)
}

func TestExportGenesis(t *testing.T) {
	s := initKeeperTestSuite(t)
	app, ctx := s.app, s.ctx
	defaultGs := uibc.DefaultGenesisState()
	defaultGs.Outflows = append(defaultGs.Outflows, sampleOutflow)
	app.UIbcQuotaKeeper.InitGenesis(ctx, *defaultGs)
	eGs := app.UIbcQuotaKeeper.ExportGenesis(ctx)
	assert.DeepEqual(t, eGs, defaultGs)
	// update params in genesis state
	defaultGs.Params.IbcStatus = uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_DISABLED
	app.UIbcQuotaKeeper.InitGenesis(ctx, *defaultGs)
	eGs = app.UIbcQuotaKeeper.ExportGenesis(ctx)
	assert.Equal(t, eGs.Params.IbcStatus, uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_DISABLED)
}
