//go:build experimental
// +build experimental

package keeper_test

import "github.com/umee-network/umee/v4/x/uibc"

func (s *IntegrationTestSuite) TestInitGenesis() {
	app, ctx := s.app, s.ctx

	defaultGs := uibc.DefaultGenesisState()

	app.UIbcQuotaKeeper.InitGenesis(ctx, *defaultGs)
	params := app.UIbcQuotaKeeper.GetParams(ctx)
	s.Require().Equal(params.IbcPause, defaultGs.Params.IbcPause)
}

func (s *IntegrationTestSuite) TestExportGenesis() {
	app, ctx := s.app, s.ctx

	defaultGs := uibc.DefaultGenesisState()

	app.UIbcQuotaKeeper.InitGenesis(ctx, *defaultGs)

	eGs := app.UIbcQuotaKeeper.ExportGenesis(ctx)
	s.Require().Equal(eGs, defaultGs)

	// update params in genesis state
	defaultGs.Params.IbcPause = uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_ENABLED
	app.UIbcQuotaKeeper.InitGenesis(ctx, *defaultGs)

	eGs = app.UIbcQuotaKeeper.ExportGenesis(ctx)
	s.Require().Equal(eGs.Params.IbcPause, uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_ENABLED)
}
