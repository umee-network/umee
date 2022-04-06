package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/oracle/types"
)

// Test the reward giving mechanism
func (s *IntegrationTestSuite) TestRewardBallotWinners() {
	// Add claim pools
	claims := map[string]types.Claim{
		valAddr.String():  types.NewClaim(10, 10, 0, valAddr),
		valAddr2.String(): types.NewClaim(20, 20, 0, valAddr2),
	}

	// Prepare reward pool
	givingAmt := sdk.NewCoins(sdk.NewInt64Coin(types.UmeeDenom, 30000000))
	err := s.app.BankKeeper.MintCoins(s.ctx, "leverage", givingAmt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, "leverage", "oracle", givingAmt)
	s.Require().NoError(err)

	var voteTargets []string
	var voteTargetDenoms []string
	params := s.app.OracleKeeper.GetParams(s.ctx)
	for _, v := range params.AcceptList {
		voteTargets = append(voteTargets, v.SymbolDenom)
		voteTargetDenoms = append(voteTargetDenoms, v.BaseDenom)
	}

	votePeriodsPerWindow := sdk.NewDec((int64)(s.app.OracleKeeper.RewardDistributionWindow(s.ctx))).
		QuoInt64((int64)(s.app.OracleKeeper.VotePeriod(s.ctx))).
		TruncateInt64()
	s.app.OracleKeeper.RewardBallotWinners(s.ctx, (int64)(s.app.OracleKeeper.VotePeriod(s.ctx)), (int64)(s.app.OracleKeeper.RewardDistributionWindow(s.ctx)), voteTargets, claims)
	outstandingRewardsDec := s.app.DistrKeeper.GetValidatorOutstandingRewardsCoins(s.ctx, valAddr)
	outstandingRewards, _ := outstandingRewardsDec.TruncateDecimal()
	s.Require().Equal(sdk.NewDecFromInt(givingAmt.AmountOf(types.UmeeDenom)).QuoInt64(votePeriodsPerWindow).QuoInt64(3).TruncateInt(),
		outstandingRewards.AmountOf(types.UmeeDenom))
}
