package keeper

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/x/oracle/types"
)

// Keeper of the oracle store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	paramSpace paramstypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	distrKeeper   types.DistributionKeeper
	StakingKeeper types.StakingKeeper

	distrName string
}

// NewKeeper constructs a new keeper for oracle
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	paramspace paramstypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	distrKeeper types.DistributionKeeper,
	stakingKeeper types.StakingKeeper,
	distrName string,
) Keeper {

	// ensure oracle module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramspace.HasKeyTable() {
		paramspace = paramspace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramspace,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		distrKeeper:   distrKeeper,
		StakingKeeper: stakingKeeper,
		distrName:     distrName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetExchangeRate gets the consensus exchange rate of an asset
// based on USD
func (k Keeper) GetExchangeRate(ctx sdk.Context, denom string) (sdk.Dec, error) {
	if denom == types.USDDenom {
		return sdk.OneDec(), nil
	}

	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetExchangeRateKey(denom))
	if b == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrUnknownDenom, denom)
	}

	decProto := sdk.DecProto{}
	k.cdc.MustUnmarshal(b, &decProto)

	return decProto.Dec, nil
}

// GetExchangeRateBase gets the consensus exchange rate of an asset
// in the base denom (e.g. ATOM -> uatom)
func (k Keeper) GetExchangeRateBase(ctx sdk.Context, denom string) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetExchangeRateKey(denom))
	if b == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrUnknownDenom, denom)
	}

	decProto := sdk.DecProto{}
	k.cdc.MustUnmarshal(b, &decProto)

	params := k.GetParams(ctx)

	for _, acceptedDenom := range params.Whitelist {
		if denom == acceptedDenom.SymbolDenom {
			exponent := sdk.NewDecFromBigInt(big.NewInt(int64(acceptedDenom.Exponent)))
			return decProto.Dec.Quo(exponent), nil
		}
	}

	return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrUnknownDenom, denom)
}

// SetExchangeRate sets the consensus exchange rate of USD denominated in the
// denom asset to the store.
func (k Keeper) SetExchangeRate(ctx sdk.Context, denom string, exchangeRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: exchangeRate})
	store.Set(types.GetExchangeRateKey(denom), bz)
}

// SetExchangeRateWithEvent sets an consensus
// exchange rate to the store with ABCI event
func (k Keeper) SetExchangeRateWithEvent(ctx sdk.Context, denom string, exchangeRate sdk.Dec) {
	k.SetExchangeRate(ctx, denom, exchangeRate)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeExchangeRateUpdate,
			sdk.NewAttribute(types.EventAttrKeyDenom, denom),
			sdk.NewAttribute(types.EventAttrKeyExchangeRate, exchangeRate.String()),
		),
	)
}

// DeleteExchangeRate deletes the consensus exchange rate of USD denominated in
// the denom asset from the store.
func (k Keeper) DeleteExchangeRate(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetExchangeRateKey(denom))
}

// IterateExchangeRates iterates over USD rates in the store.
func (k Keeper) IterateExchangeRates(ctx sdk.Context, handler func(string, sdk.Dec) bool) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.ExchangeRateKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		denom := string(iter.Key()[len(types.ExchangeRateKey):])
		dp := sdk.DecProto{}

		k.cdc.MustUnmarshal(iter.Value(), &dp)
		if handler(denom, dp.Dec) {
			break
		}
	}
}

// GetFeederDelegation gets the account address to which the validator operator
// delegated oracle vote rights.
func (k Keeper) GetFeederDelegation(ctx sdk.Context, operator sdk.ValAddress) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetFeederDelegationKey(operator))
	if bz == nil {
		// by default the right is delegated to the validator itself
		return sdk.AccAddress(operator)
	}

	return sdk.AccAddress(bz)
}

// SetFeederDelegation sets the account address to which the validator operator
// delegated oracle vote rights.
func (k Keeper) SetFeederDelegation(ctx sdk.Context, operator sdk.ValAddress, delegatedFeeder sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetFeederDelegationKey(operator), delegatedFeeder.Bytes())
}

type IterateFeederDelegationHandler = func(delegator sdk.ValAddress, delegate sdk.AccAddress) (stop bool)

// IterateFeederDelegations iterates over the feed delegates and performs a
// callback function.
func (k Keeper) IterateFeederDelegations(ctx sdk.Context, handler IterateFeederDelegationHandler) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.FeederDelegationKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		delegator := sdk.ValAddress(iter.Key()[2:])
		delegate := sdk.AccAddress(iter.Value())

		if handler(delegator, delegate) {
			break
		}
	}
}

// GetMissCounter retrieves the # of vote periods missed in this oracle slash
// window.
func (k Keeper) GetMissCounter(ctx sdk.Context, operator sdk.ValAddress) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetMissCounterKey(operator))
	if bz == nil {
		// by default the counter is zero
		return 0
	}

	var missCounter gogotypes.UInt64Value
	k.cdc.MustUnmarshal(bz, &missCounter)

	return missCounter.Value
}

// SetMissCounter updates the # of vote periods missed in this oracle slash
// window.
func (k Keeper) SetMissCounter(ctx sdk.Context, operator sdk.ValAddress, missCounter uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: missCounter})
	store.Set(types.GetMissCounterKey(operator), bz)
}

// DeleteMissCounter removes miss counter for the validator.
func (k Keeper) DeleteMissCounter(ctx sdk.Context, operator sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetMissCounterKey(operator))
}

// IterateMissCounters iterates over the miss counters and performs a callback
// function.
func (k Keeper) IterateMissCounters(ctx sdk.Context, handler func(sdk.ValAddress, uint64) bool) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.MissCounterKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		operator := sdk.ValAddress(iter.Key()[2:])

		var missCounter gogotypes.UInt64Value
		k.cdc.MustUnmarshal(iter.Value(), &missCounter)

		if handler(operator, missCounter.Value) {
			break
		}
	}
}

// GetAggregateExchangeRatePrevote retrieves an oracle prevote from the store.
func (k Keeper) GetAggregateExchangeRatePrevote(
	ctx sdk.Context,
	voter sdk.ValAddress,
) (types.AggregateExchangeRatePrevote, error) {

	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetAggregateExchangeRatePrevoteKey(voter))
	if bz == nil {
		err := sdkerrors.Wrap(types.ErrNoAggregatePrevote, voter.String())
		if err != nil {
			return types.AggregateExchangeRatePrevote{}, err
		}
	}

	var aggregatePrevote types.AggregateExchangeRatePrevote
	k.cdc.MustUnmarshal(bz, &aggregatePrevote)

	return aggregatePrevote, nil
}

// SetAggregateExchangeRatePrevote set an oracle aggregate prevote to the store.
func (k Keeper) SetAggregateExchangeRatePrevote(
	ctx sdk.Context,
	voter sdk.ValAddress,
	prevote types.AggregateExchangeRatePrevote,
) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&prevote)
	store.Set(types.GetAggregateExchangeRatePrevoteKey(voter), bz)
}

// DeleteAggregateExchangeRatePrevote deletes an oracle prevote from the store.
func (k Keeper) DeleteAggregateExchangeRatePrevote(ctx sdk.Context, voter sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAggregateExchangeRatePrevoteKey(voter))
}

// IterateAggregateExchangeRatePrevotes iterates rate over prevotes in the store
func (k Keeper) IterateAggregateExchangeRatePrevotes(
	ctx sdk.Context,
	handler func(sdk.ValAddress, types.AggregateExchangeRatePrevote) bool,
) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.AggregateExchangeRatePrevoteKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		voterAddr := sdk.ValAddress(iter.Key()[2:])

		var aggregatePrevote types.AggregateExchangeRatePrevote
		k.cdc.MustUnmarshal(iter.Value(), &aggregatePrevote)

		if handler(voterAddr, aggregatePrevote) {
			break
		}
	}
}

// GetAggregateExchangeRateVote retrieves an oracle prevote from the store.
func (k Keeper) GetAggregateExchangeRateVote(
	ctx sdk.Context,
	voter sdk.ValAddress,
) (types.AggregateExchangeRateVote, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetAggregateExchangeRateVoteKey(voter))
	if bz == nil {
		err := sdkerrors.Wrap(types.ErrNoAggregateVote, voter.String())
		if err != nil {
			return types.AggregateExchangeRateVote{}, err
		}
	}

	var aggregateVote types.AggregateExchangeRateVote
	k.cdc.MustUnmarshal(bz, &aggregateVote)

	return aggregateVote, nil
}

// SetAggregateExchangeRateVote adds an oracle aggregate prevote to the store.
func (k Keeper) SetAggregateExchangeRateVote(
	ctx sdk.Context,
	voter sdk.ValAddress,
	vote types.AggregateExchangeRateVote,
) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&vote)
	store.Set(types.GetAggregateExchangeRateVoteKey(voter), bz)
}

// DeleteAggregateExchangeRateVote deletes an oracle prevote from the store.
func (k Keeper) DeleteAggregateExchangeRateVote(ctx sdk.Context, voter sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAggregateExchangeRateVoteKey(voter))
}

type IterateExchangeRateVote = func(
	voterAddr sdk.ValAddress,
	aggregateVote types.AggregateExchangeRateVote,
) (stop bool)

// IterateAggregateExchangeRateVotes iterates rate over prevotes in the store.
func (k Keeper) IterateAggregateExchangeRateVotes(
	ctx sdk.Context,
	handler IterateExchangeRateVote,
) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.AggregateExchangeRateVoteKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		voterAddr := sdk.ValAddress(iter.Key()[2:])

		var aggregateVote types.AggregateExchangeRateVote
		k.cdc.MustUnmarshal(iter.Value(), &aggregateVote)

		if handler(voterAddr, aggregateVote) {
			break
		}
	}
}

// ValidateFeeder returns the given feeder is allowed to feed the message or not.
func (k Keeper) ValidateFeeder(ctx sdk.Context, feederAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if !feederAddr.Equals(valAddr) {
		delegate := k.GetFeederDelegation(ctx, valAddr)
		if !delegate.Equals(feederAddr) {
			return sdkerrors.Wrap(types.ErrNoVotingPermission, feederAddr.String())
		}
	}

	// check that the given validator exists
	if val := k.StakingKeeper.Validator(ctx, valAddr); val == nil || !val.IsBonded() {
		return sdkerrors.Wrapf(stakingtypes.ErrNoValidatorFound, "validator %s is not active set", valAddr.String())
	}

	return nil
}
