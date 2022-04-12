package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v2/x/oracle/types"
)

func (s *IntegrationTestSuite) TestVoteThreshold() {
	app, ctx := s.app, s.ctx

	voteDec := app.OracleKeeper.VoteThreshold(ctx)
	s.Require().Equal(sdk.MustNewDecFromStr("0.5"), voteDec)

	newVoteTreshold := sdk.MustNewDecFromStr("0.6")
	defaultParams := types.DefaultParams()
	defaultParams.VoteThreshold = newVoteTreshold
	app.OracleKeeper.SetParams(ctx, defaultParams)

	voteThresholdDec := app.OracleKeeper.VoteThreshold(ctx)
	s.Require().Equal(newVoteTreshold, voteThresholdDec)
}
