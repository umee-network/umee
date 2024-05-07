package keeper

import (
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/x/auction"
)

func TestGenesis(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	k := initKeeper(t)

	check := func(gs *auction.GenesisState) {
		err := k.InitGenesis(gs)
		require.NoError(err)

		gsOut, err := k.ExportGenesis()
		require.NoError(err)
		require.Equal(gs, gsOut)
	}

	gs := auction.DefaultGenesis()
	check(gs)

	newCoin := func(a int64) []sdk.Coin {
		return sdk.Coins{sdk.NewInt64Coin("atom", a)}
	}
	randTime := func() time.Time {
		return time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC).Add(time.Duration(rand.Int()) * time.Nanosecond)
	}

	gs.RewardAuctionId = 3
	gs.RewardsAuctions = []auction.RewardsKV{
		{1, auction.Rewards{newCoin(412313), randTime()}},
		{2, auction.Rewards{newCoin(412313), randTime()}},
		{4, auction.Rewards{newCoin(412313), randTime()}},
	}
	gs.RewardsBids = []auction.BidKV{
		{1, auction.Bid{accs.Alice, sdkmath.NewInt(153252)}},
		{2, auction.Bid{accs.Alice, sdkmath.NewInt(8521)}},
	}
	check(gs)
}
