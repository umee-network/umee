package keeper

import (
	"github.com/umee-network/umee/v3/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VotePeriod returns the number of blocks during which voting takes place.
func (k Keeper) VotePeriod(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyVotePeriod, &res)
	return
}

// VoteThreshold returns the minimum percentage of votes that must be received
// for a ballot to pass.
func (k Keeper) VoteThreshold(ctx sdk.Context) (res sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyVoteThreshold, &res)
	return
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

// StampPeriod returns the amount of blocks the oracle module waits
// between recording a set of prices.
func (k Keeper) StampPeriod(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyStampPeriod, &res)
	return
}

// SetStampPeriod updates the amount of blocks the oracle module waits
// between recording a set of prices.
func (k Keeper) SetStampPeriod(ctx sdk.Context, stampPeriod uint64) {
	k.paramSpace.Set(ctx, types.KeyStampPeriod, stampPeriod)
}

// PrunePeriod returns the max amount of blocks that a record of the set
// of exchanges is kept.
func (k Keeper) PrunePeriod(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyPrunePeriod, &res)
	return
}

// SetPrunePeriod updates the max amount of blocks that a record of the set
// of exchanges is kept.
func (k Keeper) SetPrunePeriod(ctx sdk.Context, prunePeriod uint64) {
	k.paramSpace.Set(ctx, types.KeyPrunePeriod, prunePeriod)
}

// MedianPeriod returns the amount blocks we will wait between calculating the
// median and standard deviation of the median of historic prices in the
// last Prune Period.
func (k Keeper) MedianPeriod(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyMedianPeriod, &res)
	return
}

// MedianPeriod updates the amount blocks we will wait between calculating the
// median and standard deviation of the median of historic prices in the
// last Prune Period.
func (k Keeper) SetMedianPeriod(ctx sdk.Context, medianPeriod uint64) {
	k.paramSpace.Set(ctx, types.KeyMedianPeriod, medianPeriod)
}

// GetParams returns the total set of oracle parameters.
// TODO: remove this and remove the param object from the parmspace
func (k Keeper) GetLegacyParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	k.paramSpace.Set(ctx sdk.Context, key []byte, value interface{})
	return params
}

