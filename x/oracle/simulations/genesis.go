package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/umee-network/umee/v4/x/oracle/types"
)

// Simulation parameter constants
const (
	votePeriodKey               = "vote_period"
	voteThresholdKey            = "vote_threshold"
	rewardBandKey               = "reward_band"
	rewardDistributionWindowKey = "reward_distribution_window"
	slashFractionKey            = "slash_fraction"
	slashWindowKey              = "slash_window"
	minValidPerWindowKey        = "min_valid_per_window"
	historicStampPeriodKey      = "historic_stamp_period"
	medianStampPeriodKey        = "median_stamp_period"
	maximumPriceStampsKey       = "maximum_price_stamps"
	maximumMedianStampsKey      = "maximum_median_stamps"
)

// GenVotePeriod produces a randomized VotePeriod in the range of [5, 100]
func GenVotePeriod(r *rand.Rand) uint64 {
	return uint64(5 + r.Intn(100))
}

// GenVoteThreshold produces a randomized VoteThreshold in the range of [0.34, 0.67]
func GenVoteThreshold(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(34, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(33)), 2))
}

// GenRewardBand produces a randomized RewardBand in the range of [0.000, 0.100]
func GenRewardBand(r *rand.Rand) sdk.Dec {
	return sdk.ZeroDec().Add(sdk.NewDecWithPrec(int64(r.Intn(100)), 3))
}

// GenRewardDistributionWindow produces a randomized RewardDistributionWindow in the range of [100, 100000]
func GenRewardDistributionWindow(r *rand.Rand) uint64 {
	return uint64(100 + r.Intn(100000))
}

// GenSlashFraction produces a randomized SlashFraction in the range of [0.000, 0.100]
func GenSlashFraction(r *rand.Rand) sdk.Dec {
	return sdk.ZeroDec().Add(sdk.NewDecWithPrec(int64(r.Intn(100)), 3))
}

// GenSlashWindow produces a randomized SlashWindow in the range of [100, 100000]
func GenSlashWindow(r *rand.Rand) uint64 {
	return uint64(100 + r.Intn(100000))
}

// GenMinValidPerWindow produces a randomized MinValidPerWindow in the range of [0, 0.500]
func GenMinValidPerWindow(r *rand.Rand) sdk.Dec {
	return sdk.ZeroDec().Add(sdk.NewDecWithPrec(int64(r.Intn(500)), 3))
}

// GenHistoricStampPeriod produces a randomized HistoricStampPeriod in the range of [100, 1000]
func GenHistoricStampPeriod(r *rand.Rand) uint64 {
	return uint64(100 + r.Intn(1000))
}

// GenMedianStampPeriod produces a randomized MedianStampPeriod in the range of [100, 1000]
func GenMedianStampPeriod(r *rand.Rand) uint64 {
	return uint64(10001 + r.Intn(100000))
}

// GenMaximumPriceStamps produces a randomized MaximumPriceStamps in the range of [10, 100]
func GenMaximumPriceStamps(r *rand.Rand) uint64 {
	return uint64(11 + r.Intn(100))
}

// GenMaximumMedianStamps produces a randomized MaximumMedianStamps in the range of [10, 100]
func GenMaximumMedianStamps(r *rand.Rand) uint64 {
	return uint64(11 + r.Intn(100))
}

// RandomizedGenState generates a random GenesisState for oracle
func RandomizedGenState(simState *module.SimulationState) {
	var votePeriod uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, votePeriodKey, &votePeriod, simState.Rand,
		func(r *rand.Rand) { votePeriod = GenVotePeriod(r) },
	)

	var voteThreshold sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, voteThresholdKey, &voteThreshold, simState.Rand,
		func(r *rand.Rand) { voteThreshold = GenVoteThreshold(r) },
	)

	var rewardBand sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, rewardBandKey, &rewardBand, simState.Rand,
		func(r *rand.Rand) { rewardBand = GenRewardBand(r) },
	)

	var rewardDistributionWindow uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, rewardDistributionWindowKey, &rewardDistributionWindow, simState.Rand,
		func(r *rand.Rand) { rewardDistributionWindow = GenRewardDistributionWindow(r) },
	)

	var slashFraction sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, slashFractionKey, &slashFraction, simState.Rand,
		func(r *rand.Rand) { slashFraction = GenSlashFraction(r) },
	)

	var slashWindow uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, slashWindowKey, &slashWindow, simState.Rand,
		func(r *rand.Rand) { slashWindow = GenSlashWindow(r) },
	)

	var minValidPerWindow sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, minValidPerWindowKey, &minValidPerWindow, simState.Rand,
		func(r *rand.Rand) { minValidPerWindow = GenMinValidPerWindow(r) },
	)

	var historicStampPeriod uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, historicStampPeriodKey, &historicStampPeriod, simState.Rand,
		func(r *rand.Rand) { historicStampPeriod = GenHistoricStampPeriod(r) },
	)

	var medianStampPeriod uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, medianStampPeriodKey, &medianStampPeriod, simState.Rand,
		func(r *rand.Rand) { medianStampPeriod = GenMedianStampPeriod(r) },
	)

	var maximumPriceStamps uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, maximumPriceStampsKey, &maximumPriceStamps, simState.Rand,
		func(r *rand.Rand) { maximumPriceStamps = GenMaximumPriceStamps(r) },
	)

	var maximumMedianStamps uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, maximumMedianStampsKey, &maximumMedianStamps, simState.Rand,
		func(r *rand.Rand) { maximumMedianStamps = GenMaximumMedianStamps(r) },
	)

	oracleGenesis := types.DefaultGenesisState()
	oracleGenesis.Params = types.Params{
		VotePeriod:               votePeriod,
		VoteThreshold:            voteThreshold,
		RewardBand:               rewardBand,
		RewardDistributionWindow: rewardDistributionWindow,
		AcceptList: types.DenomList{
			{SymbolDenom: types.UmeeSymbol, BaseDenom: types.UmeeDenom},
		},
		SlashFraction:       slashFraction,
		SlashWindow:         slashWindow,
		MinValidPerWindow:   minValidPerWindow,
		HistoricStampPeriod: historicStampPeriod,
		MedianStampPeriod:   medianStampPeriod,
		MaximumPriceStamps:  historicStampPeriod,
		MaximumMedianStamps: historicStampPeriod,
	}

	bz, err := json.MarshalIndent(&oracleGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated oracle parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(oracleGenesis)
}
