package keeper

import (
	"encoding/json"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
)

// GetRateLimitsOfIBCDenoms returns rate limits of all registered ibc denoms.
func (k Keeper) GetRateLimitsOfIBCDenoms(ctx sdk.Context) ([]types.RateLimit, error) {
	var rateLimitsOfIBCDenoms []types.RateLimit

	prefix := types.KeyPrefixForIBCDenom
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var rateLimitsOfIBCDenom types.RateLimit
		if err := rateLimitsOfIBCDenom.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}

		rateLimitsOfIBCDenoms = append(rateLimitsOfIBCDenoms, rateLimitsOfIBCDenom)
	}

	return rateLimitsOfIBCDenoms, nil
}

// GetRateLimitsOfIBCDenom retunes the rate limits of ibc denom.
func (k Keeper) GetRateLimitsOfIBCDenom(ctx sdk.Context, ibcDenom string) (types.RateLimit, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CreateKeyForRateLimitOfIBCDenom(ibcDenom))
	var rateLimitsOfIBCDenom types.RateLimit
	k.cdc.Unmarshal(bz, &rateLimitsOfIBCDenom)

	return rateLimitsOfIBCDenom, nil
}

// SetRateLimitsOfIBCDenoms save the rate limits of ibc denoms into store.
func (k Keeper) SetRateLimitsOfIBCDenoms(ctx sdk.Context, rateLimits []types.RateLimit) error {
	for _, rateLimitOfIBCDenom := range rateLimits {
		if err := k.SetRateLimitsOfIBCDenom(ctx, rateLimitOfIBCDenom); err != nil {
			return err
		}
	}

	return nil
}

// SetRateLimitsOfIBCDenom save the rate limits of ibc denom into store.
func (k Keeper) SetRateLimitsOfIBCDenom(ctx sdk.Context, rateLimitOfIBCDenom types.RateLimit) error {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateKeyForRateLimitOfIBCDenom(rateLimitOfIBCDenom.IbcDenom)

	bz, err := k.cdc.Marshal(&rateLimitOfIBCDenom)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

// CheckAndUpdateRateLimits
func (k Keeper) CheckAndUpdateRateLimits(ctx sdk.Context, denom, amount string) error {
	rateLimitOfIBCDenom, err := k.GetRateLimitsOfIBCDenom(ctx, denom)
	if err != nil {
		return err
	}

	sentAmount := sdk.MustNewDecFromStr(amount)

	if rateLimitOfIBCDenom.ExpiredTime.Before(ctx.BlockTime()) {
		rateLimitOfIBCDenom, err = k.ResetRateLimitsSum(ctx, rateLimitOfIBCDenom)
		if err != nil {
			return err
		}
	}

	if rateLimitOfIBCDenom.InflowSum+sentAmount.BigInt().Uint64() > rateLimitOfIBCDenom.InflowLimit {
		return types.ErrRateLimitExceeded
	}

	rateLimitOfIBCDenom.InflowSum = rateLimitOfIBCDenom.InflowSum + sentAmount.BigInt().Uint64()
	k.SetRateLimitsOfIBCDenom(ctx, rateLimitOfIBCDenom)

	return nil
}

// UndoSendRateLimit
func (k Keeper) UndoSendRateLimit(ctx sdk.Context, denom, amount string) error {
	rateLimitOfIBCDenom, err := k.GetRateLimitsOfIBCDenom(ctx, denom)
	if err != nil {
		return err
	}

	sentAmount := sdk.MustNewDecFromStr(amount)
	rateLimitOfIBCDenom.InflowSum = rateLimitOfIBCDenom.InflowSum - sentAmount.BigInt().Uint64()
	k.SetRateLimitsOfIBCDenom(ctx, rateLimitOfIBCDenom)

	return nil
}

// GetFundsFromPacket
func (k Keeper) GetFundsFromPacket(packet exported.PacketI) (string, string, error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return "", "", err
	}
	return packetData.Amount, k.GetLocalDenom(packetData.Denom), nil
}

// GetLocalDenom
func (k Keeper) GetLocalDenom(denom string) string {
	// Expected denoms in the following cases:
	//
	// send non-native: transfer/channel-0/denom -> ibc/xxx
	// send native: denom -> denom
	// recv (B)non-native: denom
	// recv (B)native: transfer/channel-0/denom
	//
	if strings.HasPrefix(denom, "transfer/") {
		denomTrace := transfertypes.ParseDenomTrace(denom)
		return denomTrace.IBCDenom()
	} else {
		return denom
	}
}

// ResetRateLimitsSum reset the expire time and inflow and outflow sum of rate limit.
func (k Keeper) ResetRateLimitsSum(ctx sdk.Context, rateLimit types.RateLimit) (types.RateLimit, error) {
	expiredTime := rateLimit.ExpiredTime.Add(rateLimit.TimeWindow)

	rateLimit.ExpiredTime = &expiredTime
	rateLimit.InflowSum = 0
	rateLimit.OutflowSum = 0

	if err := k.SetRateLimitsOfIBCDenom(ctx, rateLimit); err != nil {
		return types.RateLimit{}, err
	}

	return rateLimit, nil
}
