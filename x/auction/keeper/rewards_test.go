package keeper

import (
	"testing"
	"time"

	// sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/x/auction"
)

func TestFinalizeAuction(t *testing.T) {
	require := require.New(t)
	k := initKeeperWithGen(t)
	params := k.GetRewardsParams()

	// check first auction
	err := k.FinalizeRewardsAuction()
	require.NoError(err)
	a, id := k.getRewardsAuction(0)
	require.EqualValues(1, id)
	require.Equal(
		auction.Rewards{
			EndsAt: k.ctx.BlockHeader().Time.Add(time.Duration(params.BidDuration) * time.Second)},
		*a)
}
