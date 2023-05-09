package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
	"github.com/umee-network/umee/v4/x/leverage/fixtures"
)

func TestAPYQuery(t *testing.T) {
	k := newTestKeeper(t)
	q := Querier{k.Keeper}
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
		coin.New(atom, 10_000000),
	)

	// init a supplier with bonded uTokens
	_ = k.newBondedAccount(
		coin.New("u/"+fixtures.UmeeDenom, 100_000000),
	)

	// create three incentive programs, each of which will run for half a year but which will
	// start at slightly different times so we can test each one's contribution to total APY
	k.addIncentiveProgram(u_umee, 100, 15778800, sdk.NewInt64Coin(umee, 10_000000), true)
	k.addIncentiveProgram(u_umee, 120, 15778800, sdk.NewInt64Coin(umee, 30_000000), true)
	k.addIncentiveProgram(u_umee, 140, 15778800, sdk.NewInt64Coin(atom, 10_000000), true)

	// Advance last rewards time to 100, thus starting the first program
	k.advanceTimeTo(100)

	req1 := incentive.QueryCurrentRates{UToken: u_atom}
	expect1 := &incentive.QueryCurrentRatesResponse{
		ReferenceBond: coin.New(u_atom, 1), // zero exponent because this asset has never been incentivized
		Rewards:       sdk.NewCoins(),
	}
	resp1, err := q.CurrentRates(k.ctx, &req1)
	require.NoError(t, err)
	require.Equal(t, expect1, resp1, "zero token rates for bonded atom")

	req2 := incentive.QueryActualRates{UToken: u_atom}
	expect2 := &incentive.QueryActualRatesResponse{
		APY: sdk.ZeroDec(),
	}
	resp2, err := q.ActualRates(k.ctx, &req2)
	require.NoError(t, err)
	require.Equal(t, expect2, resp2, "zero USD rates for bonded atom")

	req3 := incentive.QueryCurrentRates{UToken: u_umee}
	expect3 := &incentive.QueryCurrentRatesResponse{
		ReferenceBond: coin.New(u_umee, 1_000000), // exponent = 6 due to proper initialization
		Rewards: sdk.NewCoins(
			coin.New(umee, 200_000), // 10 UMEE per 100 u/UMEE bonded, per half year, is 20% per year
		),
	}
	resp3, err := q.CurrentRates(k.ctx, &req3)
	require.NoError(t, err)
	require.Equal(t, expect3, resp3, "nonzero token rates for bonded umee")

	req4 := incentive.QueryActualRates{UToken: u_umee}
	expect4 := &incentive.QueryActualRatesResponse{
		APY: sdk.MustNewDecFromStr("0.2"),
	}
	resp4, err := q.ActualRates(k.ctx, &req4)
	require.NoError(t, err)
	require.Equal(t, expect4, resp4, "nonzero USD rates for bonded umee")

	// Advance last rewards time to 120, thus starting the second program and quadrupling APY
	k.advanceTimeTo(120)

	req5 := incentive.QueryCurrentRates{UToken: u_umee}
	expect5 := &incentive.QueryCurrentRatesResponse{
		ReferenceBond: coin.New(u_umee, 1_000000),
		Rewards: sdk.NewCoins(
			coin.New(umee, 800_000), // 40 UMEE per 100 u/UMEE bonded, per half year, is 80% per year
		),
	}
	resp5, err := q.CurrentRates(k.ctx, &req5)
	require.NoError(t, err)
	require.Equal(t, expect5, resp5, "increased token rates for bonded umee")

	req6 := incentive.QueryActualRates{UToken: u_umee}
	expect6 := &incentive.QueryActualRatesResponse{
		APY: sdk.MustNewDecFromStr("0.8"),
	}
	resp6, err := q.ActualRates(k.ctx, &req6)
	require.NoError(t, err)
	require.Equal(t, expect6, resp6, "increased USD rates for bonded umee")

	// Advance last rewards time to 140, thus starting the thurd program, which will add an APY based on
	// the ratio of umee and atom prices (very high) to the existing APY
	k.advanceTimeTo(140)

	req7 := incentive.QueryCurrentRates{UToken: u_umee}
	expect7 := &incentive.QueryCurrentRatesResponse{
		ReferenceBond: coin.New(u_umee, 1_000000),
		Rewards: sdk.NewCoins(
			coin.New(umee, 800_000), // 40 UMEE per 100 u/UMEE bonded, per half year, is 80% per year
			coin.New(atom, 200_000), // 10 ATOM per 100 u/UMEE bonded, per half year
		),
	}
	resp7, err := q.CurrentRates(k.ctx, &req7)
	require.NoError(t, err)
	require.Equal(t, expect7, resp7, "multi-token rates for bonded umee")

	req8 := incentive.QueryActualRates{UToken: u_umee}
	expect8 := &incentive.QueryActualRatesResponse{
		APY: sdk.MustNewDecFromStr("2.670783847980997625"), // a large but complicated APY due to price ratio
	}
	resp8, err := q.ActualRates(k.ctx, &req8)
	require.NoError(t, err)
	require.Equal(t, expect8, resp8, "multi-token USD rates for bonded umee")
}
