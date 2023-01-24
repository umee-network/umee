//go:build experimental
// +build experimental

package keeper_test

import (
	"github.com/umee-network/umee/v4/x/uibc"
)

func (s *IntegrationTestSuite) TestParams() {
	app, ctx := s.app, s.ctx

	params := app.UIbcQuotaKeeper.GetParams(ctx)
	defaultParams := uibc.DefaultParams()
	s.Require().Equal(params, defaultParams)

	// update params
	params.IbcPause = uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_ENABLED
	err := app.UIbcQuotaKeeper.SetParams(ctx, params)
	s.Require().NoError(err)
	// check the update param
	params = app.UIbcQuotaKeeper.GetParams(ctx)
	s.Require().Equal(params.IbcPause, uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_ENABLED)
}
