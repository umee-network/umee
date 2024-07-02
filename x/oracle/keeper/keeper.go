package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/oracle/types"
)

var ten = sdkmath.LegacyMustNewDecFromStr("10")

// Keeper of the oracle store
type Keeper struct {
	cdc        codec.Codec
	storeKey   storetypes.StoreKey
	paramSpace paramstypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	distrKeeper   types.DistributionKeeper
	StakingKeeper types.StakingKeeper

	distrName string
}

// NewKeeper constructs a new keeper for oracle
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetExchangeRate gets the consensus exchange rate of USD denominated in the
// denom asset from the store.
func (k Keeper) GetExchangeRate(ctx sdk.Context, symbol string) (types.ExchangeRate, error) {
	v := store.GetValue[*types.ExchangeRate](ctx.KVStore(k.storeKey), types.KeyExchangeRate(symbol),
		"exchange_rate")
	if v == nil {
		return types.ExchangeRate{}, types.ErrUnknownDenom.Wrap(symbol)
	}
	return *v, nil
}

// GetExchangeRateBase gets the consensus exchange rate of an asset
// in the base denom (e.g. ATOM -> uatom)
// TODO: needs to return timestamp as well
func (k Keeper) GetExchangeRateBase(ctx sdk.Context, denom string) (sdkmath.LegacyDec, error) {
	var symbol string
	var exponent uint64
	// Translate the base denom -> symbol
	params := k.GetParams(ctx)
	for _, listDenom := range params.AcceptList {
		if listDenom.BaseDenom == denom {
			symbol = listDenom.SymbolDenom
			exponent = uint64(listDenom.Exponent)
			break
		}
	}
	if len(symbol) == 0 {
		return sdkmath.LegacyZeroDec(), types.ErrUnknownDenom.Wrap(denom)
	}

	exchangeRate, err := k.GetExchangeRate(ctx, symbol)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	powerReduction := ten.Power(exponent)
	return exchangeRate.Rate.Quo(powerReduction), nil
}

// SetExchangeRateWithTimestamp sets the consensus exchange rate of USD denominated in the
// denom asset to the store with a timestamp specified instead of using ctx.
// NOTE: must not be used outside of genesis import.
func (k Keeper) SetExchangeRateWithTimestamp(ctx sdk.Context, denom string, rate sdkmath.LegacyDec, t time.Time) {
	key := types.KeyExchangeRate(denom)
	val := types.ExchangeRate{Rate: rate, Timestamp: t}
	err := store.SetValue[*types.ExchangeRate](ctx.KVStore(k.storeKey), key, &val, "exchange_rate")
	util.Panic(err)
}

// SetExchangeRate sets an consensus
// exchange rate to the store with ABCI event
func (k Keeper) SetExchangeRate(ctx sdk.Context, denom string, rate sdkmath.LegacyDec) {
	k.SetExchangeRateWithTimestamp(ctx, denom, rate, ctx.BlockTime())
	sdkutil.Emit(&ctx, &types.EventSetFxRate{
		Denom: denom, Rate: rate,
	})
}

// IterateExchangeRates iterates over all USD rates in the store.
func (k Keeper) IterateExchangeRates(ctx sdk.Context, handler func(string, sdkmath.LegacyDec, time.Time) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.KeyPrefixExchangeRate)
	defer iter.Close()
	prefixLen := len(types.KeyPrefixExchangeRate)
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		denom := string(key[prefixLen : len(key)-1]) // -1 to remove the null suffix
		var exgRate types.ExchangeRate
		err := exgRate.Unmarshal(iter.Value())
		util.Panic(err)

		if handler(denom, exgRate.Rate, exgRate.Timestamp) {
			break
		}
	}
}

func (k Keeper) ClearExchangeRates(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.KeyPrefixExchangeRate)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// GetFeederDelegation gets the account address to which the validator operator
// delegated oracle vote rights.
func (k Keeper) GetFeederDelegation(ctx sdk.Context, vAddr sdk.ValAddress) (sdk.AccAddress, error) {
	// check that the given validator exists
	if val, err := k.StakingKeeper.Validator(ctx, vAddr); err == nil || !val.IsBonded() {
		return nil, stakingtypes.ErrNoValidatorFound.Wrapf("validator %s is not in active set", vAddr)
	}

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyFeederDelegation(vAddr))
	if bz == nil {
		// no delegation, so validator itself must provide price feed
		return sdk.AccAddress(vAddr), nil
	}
	return sdk.AccAddress(bz), nil
}

// SetFeederDelegation sets the account address to which the validator operator
// delegated oracle vote rights.
func (k Keeper) SetFeederDelegation(ctx sdk.Context, operator sdk.ValAddress, delegatedFeeder sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyFeederDelegation(operator), delegatedFeeder.Bytes())
}

type IterateFeederDelegationHandler func(delegator sdk.ValAddress, delegate sdk.AccAddress) (stop bool)

// IterateFeederDelegations iterates over the feed delegates and performs a
// callback function.
func (k Keeper) IterateFeederDelegations(ctx sdk.Context, handler IterateFeederDelegationHandler) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.KeyPrefixFeederDelegation)
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

	bz := store.Get(types.KeyMissCounter(operator))
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
	store.Set(types.KeyMissCounter(operator), bz)
}

// DeleteMissCounter removes miss counter for the validator.
func (k Keeper) DeleteMissCounter(ctx sdk.Context, operator sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyMissCounter(operator))
}

// IterateMissCounters iterates over the miss counters and performs a callback
// function.
func (k Keeper) IterateMissCounters(ctx sdk.Context, handler func(sdk.ValAddress, uint64) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.KeyPrefixMissCounter)
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

// GetAllMissCounters returns all bonded validators pf miss counter value
func (k Keeper) GetAllMissCounters(ctx sdk.Context) []types.PriceMissCounter {
	missCounters := make([]types.PriceMissCounter, 0)
	validators, _ := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
	for _, v := range validators {
		valAddress, _ := k.StakingKeeper.ValidatorAddressCodec().StringToBytes(v.GetOperator())
		missCounters = append(missCounters, types.PriceMissCounter{
			Validator:   v.OperatorAddress,
			MissCounter: k.GetMissCounter(ctx, valAddress),
		})
	}
	return missCounters
}

// GetAggregateExchangeRatePrevote retrieves an oracle prevote from the store.
func (k Keeper) GetAggregateExchangeRatePrevote(
	ctx sdk.Context,
	voter sdk.ValAddress,
) (types.AggregateExchangeRatePrevote, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyAggregateExchangeRatePrevote(voter))
	if bz == nil {
		return types.AggregateExchangeRatePrevote{}, types.ErrNoAggregatePrevote.Wrap(voter.String())
	}

	var aggregatePrevote types.AggregateExchangeRatePrevote
	k.cdc.MustUnmarshal(bz, &aggregatePrevote)

	return aggregatePrevote, nil
}

// HasAggregateExchangeRatePrevote checks if a validator has an existing prevote.
func (k Keeper) HasAggregateExchangeRatePrevote(
	ctx sdk.Context,
	voter sdk.ValAddress,
) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.KeyAggregateExchangeRatePrevote(voter))
}

// SetAggregateExchangeRatePrevote set an oracle aggregate prevote to the store.
func (k Keeper) SetAggregateExchangeRatePrevote(
	ctx sdk.Context,
	voter sdk.ValAddress,
	prevote types.AggregateExchangeRatePrevote,
) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&prevote)
	store.Set(types.KeyAggregateExchangeRatePrevote(voter), bz)
}

// DeleteAggregateExchangeRatePrevote deletes an oracle prevote from the store.
func (k Keeper) DeleteAggregateExchangeRatePrevote(ctx sdk.Context, voter sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyAggregateExchangeRatePrevote(voter))
}

// IterateAggregateExchangeRatePrevotes iterates rate over prevotes in the store
func (k Keeper) IterateAggregateExchangeRatePrevotes(
	ctx sdk.Context,
	handler func(sdk.ValAddress, types.AggregateExchangeRatePrevote) bool,
) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.KeyPrefixAggregateExchangeRatePrevote)
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

	bz := store.Get(types.KeyAggregateExchangeRateVote(voter))
	if bz == nil {
		return types.AggregateExchangeRateVote{}, types.ErrNoAggregateVote.Wrap(voter.String())
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
	store.Set(types.KeyAggregateExchangeRateVote(voter), bz)
}

// DeleteAggregateExchangeRateVote deletes an oracle prevote from the store.
func (k Keeper) DeleteAggregateExchangeRateVote(ctx sdk.Context, voter sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyAggregateExchangeRateVote(voter))
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
	iter := storetypes.KVStorePrefixIterator(store, types.KeyPrefixAggregateExchangeRateVote)
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

// ValidateFeeder returns error if the given feeder is not allowed to feed the message.
func (k Keeper) ValidateFeeder(ctx sdk.Context, feederAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	delegate, err := k.GetFeederDelegation(ctx, valAddr)
	if err != nil {
		return err
	}
	if !delegate.Equals(feederAddr) {
		return types.ErrNoVotingPermission.Wrap(feederAddr.String())
	}

	return nil
}

// IterateOldExchangeRates iterates over all old exchange rates from store and returns them.
// Note: this is only used for v6.1 Migrations
func (k Keeper) IterateOldExchangeRates(ctx sdk.Context, handler func(string, sdkmath.LegacyDec) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.KeyPrefixExchangeRate)
	defer iter.Close()
	prefixLen := len(types.KeyPrefixExchangeRate)

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		denom := string(key[prefixLen : len(key)-1]) // -1 to remove the null suffix
		dp := sdk.DecProto{}
		k.cdc.MustUnmarshal(iter.Value(), &dp)

		if handler(denom, dp.Dec) {
			break
		}
	}
}
