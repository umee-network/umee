package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/incentive"
)

func TestQueries(t *testing.T) {
	t.Parallel()
	k := newTestKeeper(t)
	q := Querier{k.Keeper}

	alice := k.initScenario1()

	expect1 := &incentive.QueryAccountBondsResponse{
		Bonded: sdk.NewCoins(
			coin.New(uUmee, 90_000000),
			coin.New(uAtom, 45_000000),
		),
		Unbonding: sdk.NewCoins(
			coin.New(uUmee, 10_000000),
			coin.New(uAtom, 5_000000),
		),
		Unbondings: []incentive.Unbonding{
			{
				Start:  90,
				End:    86490,
				UToken: coin.New(uAtom, 5_000000),
			},
			{
				Start:  90,
				End:    86490,
				UToken: coin.New(uUmee, 5_000000),
			},
			{
				Start:  90,
				End:    86490,
				UToken: coin.New(uUmee, 5_000000),
			},
		},
	}
	resp1, err := q.AccountBonds(k.ctx, &incentive.QueryAccountBonds{
		Address: alice.String(),
	})
	require.NoError(t, err)
	require.Equal(t, expect1, resp1, "account bonds query")

	expect2 := &incentive.QueryPendingRewardsResponse{
		Rewards: sdk.NewCoins(
			coin.New(umee, 13_333332),
		),
	}
	resp2, err := q.PendingRewards(k.ctx, &incentive.QueryPendingRewards{
		Address: alice.String(),
	})
	require.NoError(t, err)
	require.Equal(t, expect2, resp2, "pending rewards query")

	expect3 := &incentive.QueryTotalBondedResponse{
		Bonded: sdk.NewCoins(
			coin.New(uUmee, 90_000000),
			coin.New(uAtom, 45_000000),
		),
	}
	resp3, err := q.TotalBonded(k.ctx, &incentive.QueryTotalBonded{})
	require.NoError(t, err)
	require.Equal(t, expect3, resp3, "total bonded query (all denoms)")

	expect4 := &incentive.QueryTotalBondedResponse{
		Bonded: sdk.NewCoins(
			coin.New(uUmee, 90_000000),
		),
	}
	resp4, err := q.TotalBonded(k.ctx, &incentive.QueryTotalBonded{
		Denom: uUmee,
	})
	require.NoError(t, err)
	require.Equal(t, expect4, resp4, "total bonded query (one denom)")

	expect5 := &incentive.QueryTotalUnbondingResponse{
		Unbonding: sdk.NewCoins(
			coin.New(uUmee, 10_000000),
			coin.New(uAtom, 5_000000),
		),
	}
	resp5, err := q.TotalUnbonding(k.ctx, &incentive.QueryTotalUnbonding{})
	require.NoError(t, err)
	require.Equal(t, expect5, resp5, "total unbonding query (all denoms)")

	expect6 := &incentive.QueryTotalUnbondingResponse{
		Unbonding: sdk.NewCoins(
			coin.New(uUmee, 10_000000),
		),
	}
	resp6, err := q.TotalUnbonding(k.ctx, &incentive.QueryTotalUnbonding{
		Denom: uUmee,
	})
	require.NoError(t, err)
	require.Equal(t, expect6, resp6, "total unbonding query (one denom)")

	expect7 := &incentive.QueryLastRewardTimeResponse{
		Time: 100,
	}
	resp7, err := q.LastRewardTime(k.ctx, &incentive.QueryLastRewardTime{})
	require.NoError(t, err)
	require.Equal(t, expect7, resp7, "last reward time query")

	expect8 := &incentive.QueryParamsResponse{
		Params: k.GetParams(k.ctx),
	}
	resp8, err := q.Params(k.ctx, &incentive.QueryParams{})
	require.NoError(t, err)
	require.Equal(t, expect8, resp8, "params query")

	programs, err := k.getAllIncentivePrograms(k.ctx, incentive.ProgramStatusUpcoming)
	require.NoError(t, err)
	expect9 := &incentive.QueryUpcomingIncentiveProgramsResponse{
		Programs: programs,
	}
	resp9, err := q.UpcomingIncentivePrograms(k.ctx, &incentive.QueryUpcomingIncentivePrograms{})
	require.NoError(t, err)
	require.Equal(t, expect9, resp9, "upcoming programs query")

	programs, err = k.getAllIncentivePrograms(k.ctx, incentive.ProgramStatusOngoing)
	require.NoError(t, err)
	expect10 := &incentive.QueryOngoingIncentiveProgramsResponse{
		Programs: programs,
	}
	resp10, err := q.OngoingIncentivePrograms(k.ctx, &incentive.QueryOngoingIncentivePrograms{})
	require.NoError(t, err)
	require.Equal(t, expect10, resp10, "ongoing programs query")

	programs, err = k.getAllIncentivePrograms(k.ctx, incentive.ProgramStatusCompleted)
	require.NoError(t, err)
	expect11 := &incentive.QueryCompletedIncentiveProgramsResponse{
		Programs: programs,
	}
	resp11, err := q.CompletedIncentivePrograms(k.ctx, &incentive.QueryCompletedIncentivePrograms{})
	require.NoError(t, err)
	require.Equal(t, expect11, resp11, "completed programs query")

	program, _, err := k.getIncentiveProgram(k.ctx, 1)
	require.NoError(t, err)
	expect12 := &incentive.QueryIncentiveProgramResponse{
		Program: program,
	}
	resp12, err := q.IncentiveProgram(k.ctx, &incentive.QueryIncentiveProgram{Id: 1})
	require.NoError(t, err)
	require.Equal(t, expect12, resp12, "incentive program query")
}

func TestAPYQuery(t *testing.T) {
	t.Parallel()
	k := newTestKeeper(t)
	q := Querier{k.Keeper}
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
		coin.New(atom, 10_000000),
	)

	// init a supplier with bonded uTokens
	_ = k.newBondedAccount(
		coin.New(uUmee, 100_000000),
	)

	// create three incentive programs, each of which will run for half a year but which will
	// start at slightly different times so we can test each one's contribution to total APY
	k.addIncentiveProgram(uUmee, 100, 15778800, sdk.NewInt64Coin(umee, 10_000000), true)
	k.addIncentiveProgram(uUmee, 120, 15778800, sdk.NewInt64Coin(umee, 30_000000), true)
	k.addIncentiveProgram(uUmee, 140, 15778800, sdk.NewInt64Coin(atom, 10_000000), true)

	// Advance last rewards time to 100, thus starting the first program
	k.advanceTimeTo(100)

	req1 := incentive.QueryCurrentRates{UToken: uAtom}
	expect1 := &incentive.QueryCurrentRatesResponse{
		ReferenceBond: coin.New(uAtom, 1), // zero exponent because this asset has never been incentivized
		Rewards:       sdk.NewCoins(),
	}
	resp1, err := q.CurrentRates(k.ctx, &req1)
	require.NoError(t, err)
	require.Equal(t, expect1, resp1, "zero token rates for bonded atom")

	req2 := incentive.QueryActualRates{UToken: uAtom}
	expect2 := &incentive.QueryActualRatesResponse{
		APY: sdk.ZeroDec(),
	}
	resp2, err := q.ActualRates(k.ctx, &req2)
	require.NoError(t, err)
	require.Equal(t, expect2, resp2, "zero USD rates for bonded atom")

	req3 := incentive.QueryCurrentRates{UToken: uUmee}
	expect3 := &incentive.QueryCurrentRatesResponse{
		ReferenceBond: coin.New(uUmee, 1_000000), // exponent = 6 due to proper initialization
		Rewards: sdk.NewCoins(
			coin.New(umee, 200_000), // 10 UMEE per 100 u/UMEE bonded, per half year, is 20% per year
		),
	}
	resp3, err := q.CurrentRates(k.ctx, &req3)
	require.NoError(t, err)
	require.Equal(t, expect3, resp3, "nonzero token rates for bonded umee")

	req4 := incentive.QueryActualRates{UToken: uUmee}
	expect4 := &incentive.QueryActualRatesResponse{
		APY: sdk.MustNewDecFromStr("0.2"),
	}
	resp4, err := q.ActualRates(k.ctx, &req4)
	require.NoError(t, err)
	require.Equal(t, expect4, resp4, "nonzero USD rates for bonded umee")

	// Advance last rewards time to 120, thus starting the second program and quadrupling APY
	k.advanceTimeTo(120)

	req5 := incentive.QueryCurrentRates{UToken: uUmee}
	expect5 := &incentive.QueryCurrentRatesResponse{
		ReferenceBond: coin.New(uUmee, 1_000000),
		Rewards: sdk.NewCoins(
			coin.New(umee, 800_000), // 40 UMEE per 100 u/UMEE bonded, per half year, is 80% per year
		),
	}
	resp5, err := q.CurrentRates(k.ctx, &req5)
	require.NoError(t, err)
	require.Equal(t, expect5, resp5, "increased token rates for bonded umee")

	req6 := incentive.QueryActualRates{UToken: uUmee}
	expect6 := &incentive.QueryActualRatesResponse{
		APY: sdk.MustNewDecFromStr("0.8"),
	}
	resp6, err := q.ActualRates(k.ctx, &req6)
	require.NoError(t, err)
	require.Equal(t, expect6, resp6, "increased USD rates for bonded umee")

	// Advance last rewards time to 140, thus starting the thurd program, which will add an APY based on
	// the ratio of umee and atom prices (very high) to the existing APY
	k.advanceTimeTo(140)

	req7 := incentive.QueryCurrentRates{UToken: uUmee}
	expect7 := &incentive.QueryCurrentRatesResponse{
		ReferenceBond: coin.New(uUmee, 1_000000),
		Rewards: sdk.NewCoins(
			coin.New(umee, 800_000), // 40 UMEE per 100 u/UMEE bonded, per half year, is 80% per year
			coin.New(atom, 200_000), // 10 ATOM per 100 u/UMEE bonded, per half year
		),
	}
	resp7, err := q.CurrentRates(k.ctx, &req7)
	require.NoError(t, err)
	require.Equal(t, expect7, resp7, "multi-token rates for bonded umee")

	req8 := incentive.QueryActualRates{UToken: uUmee}
	expect8 := &incentive.QueryActualRatesResponse{
		APY: sdk.MustNewDecFromStr("2.670783847980997625"), // a large but complicated APY due to price ratio
	}
	resp8, err := q.ActualRates(k.ctx, &req8)
	require.NoError(t, err)
	require.Equal(t, expect8, resp8, "multi-token USD rates for bonded umee")
}
