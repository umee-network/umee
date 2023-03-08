package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/oracle/types"
)

// VotePeriod returns the number of blocks during which voting takes place.
func (k Keeper) VotePeriod(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyVotePeriod, &res)
	return
}

// VoteThreshold returns the minimum portion of combined validator power of votes
// that must be received for a ballot to pass.
func (k Keeper) VoteThreshold(ctx sdk.Context) (res sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyVoteThreshold, &res)
	return
}

// SetVoteThreshold sets min combined validator power voting on a denom to accept
// it as valid.
// TODO: this is used in tests, we should refactor the way how this is handled.
func (k Keeper) SetVoteThreshold(ctx sdk.Context, threshold sdk.Dec) error {
	if err := types.ValidateVoteThreshold(threshold); err != nil {
		return err
	}
	k.paramSpace.Set(ctx, types.KeyVoteThreshold, &threshold)
	return nil
}

// RewardBand returns the ratio of allowable exchange rate error that a validator
// can be rewarded.
func (k Keeper) RewardBand(ctx sdk.Context) (res sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyRewardBand, &res)
	return
}

// RewardDistributionWindow returns the number of vote periods during which
// seigniorage reward comes in and then is distributed.
func (k Keeper) RewardDistributionWindow(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyRewardDistributionWindow, &res)
	return
}

// AcceptList returns the denom list that can be activated
func (k Keeper) AcceptList(ctx sdk.Context) (res types.DenomList) {
	k.paramSpace.Get(ctx, types.KeyAcceptList, &res)
	return
}

// SetAcceptList updates the accepted list of assets supported by the x/oracle
// module.
func (k Keeper) SetAcceptList(ctx sdk.Context, acceptList types.DenomList) {
	k.paramSpace.Set(ctx, types.KeyAcceptList, acceptList)
}

// SlashFraction returns oracle voting penalty rate
func (k Keeper) SlashFraction(ctx sdk.Context) (res sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeySlashFraction, &res)
	return
}

// SlashWindow returns # of vote period for oracle slashing
func (k Keeper) SlashWindow(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeySlashWindow, &res)
	return
}

// MinValidPerWindow returns oracle slashing threshold
func (k Keeper) MinValidPerWindow(ctx sdk.Context) (res sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyMinValidPerWindow, &res)
	return
}

// HistoricStampPeriod returns the amount of blocks the oracle module waits
// before recording a new historic price.
func (k Keeper) HistoricStampPeriod(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyHistoricStampPeriod, &res)
	return
}

// SetHistoricStampPeriod updates the amount of blocks the oracle module waits
// before recording a new historic price.
func (k Keeper) SetHistoricStampPeriod(ctx sdk.Context, historicPriceStampPeriod uint64) {
	k.paramSpace.Set(ctx, types.KeyHistoricStampPeriod, historicPriceStampPeriod)
}

// MedianStampPeriod returns the amount blocks the oracle module waits between
// calculating a new median and standard deviation of that median.
func (k Keeper) MedianStampPeriod(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyMedianStampPeriod, &res)
	return
}

// SetMedianStampPeriod updates the amount blocks the oracle module waits between
// calculating a new median and standard deviation of that median.
func (k Keeper) SetMedianStampPeriod(ctx sdk.Context, medianStampPeriod uint64) {
	k.paramSpace.Set(ctx, types.KeyMedianStampPeriod, medianStampPeriod)
}

// MaximumMedianStamps returns the maximum amount of historic prices the oracle
// module will hold.
func (k Keeper) MaximumPriceStamps(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyMaximumPriceStamps, &res)
	return
}

// SetMaximumPriceStamps updates the the maximum amount of historic prices the
// oracle module will hold.
func (k Keeper) SetMaximumPriceStamps(ctx sdk.Context, maximumPriceStamps uint64) {
	k.paramSpace.Set(ctx, types.KeyMaximumPriceStamps, maximumPriceStamps)
}

// MaximumMedianStamps returns the maximum amount of medians the oracle module will
// hold.
func (k Keeper) MaximumMedianStamps(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyMaximumMedianStamps, &res)
	return
}

// SetMaximumMedianStamps updates the the maximum amount of medians the oracle module will
// hold.
func (k Keeper) SetMaximumMedianStamps(ctx sdk.Context, maximumMedianStamps uint64) {
	k.paramSpace.Set(ctx, types.KeyMaximumMedianStamps, maximumMedianStamps)
}

// GetParams returns the total set of oracle parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of oracle parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
