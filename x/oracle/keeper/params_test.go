package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/umee-network/umee/v6/x/oracle/types"
)

func (s *IntegrationTestSuite) TestVoteThreshold() {
	app, ctx := s.app, s.ctx

	voteDec := app.OracleKeeper.VoteThreshold(ctx)
	s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("0.5"), voteDec)

	newVoteTreshold := sdkmath.LegacyMustNewDecFromStr("0.6")
	defaultParams := types.DefaultParams()
	defaultParams.VoteThreshold = newVoteTreshold
	app.OracleKeeper.SetParams(ctx, defaultParams)

	voteThresholdDec := app.OracleKeeper.VoteThreshold(ctx)
	s.Require().Equal(newVoteTreshold, voteThresholdDec)
}
