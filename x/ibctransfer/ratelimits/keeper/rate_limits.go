package keeper

import (
	"encoding/json"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/umee-network/umee/v3/x/ibctransfer"
)

// GetRateLimitsOfIBCDenoms returns rate limits of all registered ibc denoms.
func (k Keeper) GetRateLimitsOfIBCDenoms(ctx sdk.Context) ([]ibctransfer.RateLimit, error) {
	var rateLimitsOfIBCDenoms []ibctransfer.RateLimit

	prefix := ibctransfer.KeyPrefixForIBCDenom
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var rateLimitsOfIBCDenom ibctransfer.RateLimit
		if err := rateLimitsOfIBCDenom.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}

		rateLimitsOfIBCDenoms = append(rateLimitsOfIBCDenoms, rateLimitsOfIBCDenom)
	}

	return rateLimitsOfIBCDenoms, nil
}

// GetRateLimitsOfIBCDenom retunes the rate limits of ibc denom.
func (k Keeper) GetRateLimitsOfIBCDenom(ctx sdk.Context, ibcDenom string) (*ibctransfer.RateLimit, error) {
	rate, err := k.getRateLimitsOfIBCDenom(ctx, ibcDenom)
	if err != nil {
		return nil, err
	}
	if rate == nil {
		return nil, ibctransfer.ErrNoRateLimitsForIBCDenom
	}

	return rate, nil
}

// getRateLimitsOfIBCDenom retunes the rate limits of ibc denom.
func (k Keeper) getRateLimitsOfIBCDenom(ctx sdk.Context, ibcDenom string) (*ibctransfer.RateLimit, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ibctransfer.CreateKeyForRateLimitOfIBCDenom(ibcDenom))
	var rateLimitsOfIBCDenom ibctransfer.RateLimit
	k.cdc.Unmarshal(bz, &rateLimitsOfIBCDenom)

	return &rateLimitsOfIBCDenom, nil
}

// SetRateLimitsOfIBCDenoms save the rate limits of ibc denoms into store.
func (k Keeper) SetRateLimitsOfIBCDenoms(ctx sdk.Context, rateLimits []ibctransfer.RateLimit) error {
	for _, rateLimitOfIBCDenom := range rateLimits {
		if err := k.SetRateLimitsOfIBCDenom(ctx, &rateLimitOfIBCDenom); err != nil {
			return err
		}
	}

	return nil
}

// SetRateLimitsOfIBCDenom save the rate limits of ibc denom into store.
func (k Keeper) SetRateLimitsOfIBCDenom(ctx sdk.Context, rateLimitOfIBCDenom *ibctransfer.RateLimit) error {
	store := ctx.KVStore(k.storeKey)
	key := ibctransfer.CreateKeyForRateLimitOfIBCDenom(rateLimitOfIBCDenom.IbcDenom)

	bz, err := k.cdc.Marshal(rateLimitOfIBCDenom)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

// CheckAndUpdateRateLimits
func (k Keeper) CheckAndUpdateRateLimits(ctx sdk.Context, denom string, sendAmount string) error {
	rateLimitOfIBCDenom, err := k.GetRateLimitsOfIBCDenom(ctx, denom)
	if err != nil {
		return err
	}

	if rateLimitOfIBCDenom == nil {
		return nil
	}

	amount, err := strconv.ParseUint(sendAmount, 10, 64)
	if err != nil {
		return err
	}

	if rateLimitOfIBCDenom.ExpiredTime.Before(ctx.BlockTime()) {
		rateLimitOfIBCDenom, err = k.ResetRateLimitsSum(ctx, rateLimitOfIBCDenom)
		if err != nil {
			return err
		}
	}

	outflowSum := sdk.NewIntFromUint64(rateLimitOfIBCDenom.OutflowSum)
	sendingAmount := sdk.NewIntFromUint64(amount)
	outflowLimit := sdk.NewIntFromUint64(rateLimitOfIBCDenom.OutflowLimit)
	if outflowSum.Add(sendingAmount).GT(outflowLimit) {
		return ibctransfer.ErrRateLimitExceeded
	}

	rateLimitOfIBCDenom.OutflowSum = outflowSum.Add(sendingAmount).Uint64()
	k.SetRateLimitsOfIBCDenom(ctx, rateLimitOfIBCDenom)

	return nil
}

// UndoSendRateLimit
func (k Keeper) UndoSendRateLimit(ctx sdk.Context, denom, sendAmount string) error {
	rateLimitOfIBCDenom, err := k.GetRateLimitsOfIBCDenom(ctx, denom)
	if err != nil {
		return err
	}

	if rateLimitOfIBCDenom == nil {
		return nil
	}

	amount, err := strconv.ParseUint(sendAmount, 10, 64)
	if err != nil {
		return err
	}

	outflowSum := sdk.NewIntFromUint64(rateLimitOfIBCDenom.OutflowSum)
	sendingAmount := sdk.NewIntFromUint64(amount)
	rateLimitOfIBCDenom.OutflowSum = outflowSum.Sub(sendingAmount).Uint64()

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

// ResetRateLimitsSum reset the expire time and outflow sum of rate limit.
func (k Keeper) ResetRateLimitsSum(ctx sdk.Context, rateLimit *ibctransfer.RateLimit) (*ibctransfer.RateLimit, error) {
	expiredTime := rateLimit.ExpiredTime.Add(rateLimit.TimeWindow)

	rateLimit.ExpiredTime = &expiredTime
	rateLimit.OutflowSum = 0

	if err := k.SetRateLimitsOfIBCDenom(ctx, rateLimit); err != nil {
		return nil, err
	}

	return rateLimit, nil
}
