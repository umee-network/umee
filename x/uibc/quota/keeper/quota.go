package keeper

import (
	"encoding/json"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	ltypes "github.com/umee-network/umee/v4/x/leverage/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

// GetQuotaOfIBCDenoms returns quota of all tokens.
func (k Keeper) GetQuotaOfIBCDenoms(ctx sdk.Context) ([]uibc.Quota, error) {
	var quotaOfIBCDenoms []uibc.Quota

	prefix := uibc.KeyPrefixDenomQuota
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var quotaOfIBCDenom uibc.Quota
		if err := quotaOfIBCDenom.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}

		quotaOfIBCDenoms = append(quotaOfIBCDenoms, quotaOfIBCDenom)
	}

	return quotaOfIBCDenoms, nil
}

// GetQuotaByDenom retunes the rate limits of ibc denom.
func (k Keeper) GetQuotaByDenom(ctx sdk.Context, ibcDenom string) (*uibc.Quota, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.KeyTotalOutflows(ibcDenom))
	if bz == nil {
		return nil, uibc.ErrNoQuotaForIBCDenom
	}

	var quotaOfIBCDenom uibc.Quota
	if err := k.cdc.Unmarshal(bz, &quotaOfIBCDenom); err != nil {
		return nil, err
	}

	return &quotaOfIBCDenom, nil
}

// SetDenomQuotas save the updated token quota.
func (k Keeper) SetDenomQuotas(ctx sdk.Context, quotas []uibc.Quota) {
	for _, quotaOfIBCDenom := range quotas {
		k.SetDenomQuota(ctx, quotaOfIBCDenom)
	}
}

// SetDenomQuota save the quota of denom into store.
func (k Keeper) SetDenomQuota(ctx sdk.Context, quotaOfIBCDenom uibc.Quota) {
	store := ctx.KVStore(k.storeKey)
	key := uibc.KeyTotalOutflows(quotaOfIBCDenom.IbcDenom)
	store.Set(key, k.cdc.MustMarshal(&quotaOfIBCDenom))
}

// GetTotalOutflowSum returns the total outflow of ibc-transfer amount.
func (k Keeper) GetTotalOutflowSum(ctx sdk.Context) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.KeyPrefixTotalOutflows)
	return sdk.MustNewDecFromStr(string(bz))
}

// SetTotalOutflowSum save the total outflow of ibc-transfer amount.
func (k Keeper) SetTotalOutflowSum(ctx sdk.Context, amount sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	store.Set(uibc.KeyPrefixTotalOutflows, []byte(amount.String()))
}

// SetExpire save the quota expires time of ibc denom into store.
func (k Keeper) SetExpire(ctx sdk.Context, expires time.Time) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := expires.MarshalBinary()
	if err != nil {
		return err
	}
	store.Set(uibc.KeyPrefixQuotaExpires, bz)

	return nil
}

// GetExpire returns ibc-transfer quota expires time.
func (k Keeper) GetExpire(ctx sdk.Context) (*time.Time, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.KeyPrefixQuotaExpires)
	if bz == nil {
		return nil, nil
	}
	quotaExpires := ctx.BlockTime()
	if err := quotaExpires.UnmarshalBinary(bz); err != nil {
		return &quotaExpires, err
	}
	return &quotaExpires, nil
}

// ResetQuota will reset the ibc-transfer quotas
func (k Keeper) ResetQuota(ctx sdk.Context) error {
	qd := k.GetParams(ctx).QuotaDuration
	zero := sdk.NewDec(0)
	newExpires := ctx.BlockTime().Add(qd)
	if err := k.SetExpire(ctx, newExpires); err != nil {
		return err
	}
	k.SetTotalOutflowSum(ctx, zero)

	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, uibc.KeyPrefixDenomQuota)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		ibcDenom := uibc.DenomFromKeyTotalOutflows(iter.Key())
		q := uibc.Quota{IbcDenom: ibcDenom, OutflowSum: zero}
		store.Set(uibc.KeyTotalOutflows(ibcDenom), k.cdc.MustMarshal(&q))
	}
	return nil
}

// CheckAndUpdateQuota checks the quota of ibc-transfer of denom
func (k Keeper) CheckAndUpdateQuota(ctx sdk.Context, denom string, amount sdkmath.Int) error {
	params := k.GetParams(ctx)

	quotaOfIBCDenom, err := k.GetQuotaByDenom(ctx, denom)
	if err != nil {
		return err
	}

	if quotaOfIBCDenom == nil {
		return nil
	}

	exchangePrice, err := k.getExchangePrice(ctx, denom, amount)
	if err != nil {
		// Note: skip the ibc-transfer quota checking if `denom` is not support by leverage
		if err.Error() == sdkerrors.Wrap(ltypes.ErrNotRegisteredToken, denom).Error() {
			return nil
		} else if err != nil {
			return err
		}
	}

	// checking ibc-transfer token quota
	quotaOfIBCDenom.OutflowSum = quotaOfIBCDenom.OutflowSum.Add(exchangePrice)
	if quotaOfIBCDenom.OutflowSum.GT(params.TokenQuota) {
		return uibc.ErrQuotaExceeded
	}

	// checking total outflow quota
	totalOutflowSum := k.GetTotalOutflowSum(ctx).Add(exchangePrice)
	if totalOutflowSum.GT(params.TotalQuota) {
		return uibc.ErrQuotaExceeded
	}

	// update the per token outflow sum
	k.SetDenomQuota(ctx, *quotaOfIBCDenom)
	// updating the total outflow sum
	k.SetTotalOutflowSum(ctx, totalOutflowSum)
	return nil
}

func (k Keeper) getExchangePrice(ctx sdk.Context, denom string, amount sdkmath.Int) (sdk.Dec, error) {
	transferCoin := sdk.NewCoin(denom, amount)
	var err error

	// convert to base asset if it is `uToken`
	if ltypes.HasUTokenPrefix(denom) {
		transferCoin, err = k.leverageKeeper.ExchangeUToken(ctx, transferCoin)
		if err != nil {
			return sdk.Dec{}, err
		}
	}

	// get the exchange price (eg: UMEE) in USD from oracle using base denom eg: `uumee`
	return k.leverageKeeper.TokenValue(ctx, transferCoin, ltypes.PriceModeSpot)
}

// UndoUpdateQuota undo the quota of ibc denom
func (k Keeper) UndoUpdateQuota(ctx sdk.Context, denom string, amount sdkmath.Int) error {
	quotaOfIBCDenom, err := k.GetQuotaByDenom(ctx, denom)
	if err != nil {
		return err
	}

	if quotaOfIBCDenom == nil {
		return nil
	}

	exchangePrice, err := k.getExchangePrice(ctx, denom, amount)
	if err != nil {
		// Note: skip the ibc-transfer quota checking if `denom` is not support by leverage
		if err.Error() == sdkerrors.Wrap(ltypes.ErrNotRegisteredToken, denom).Error() {
			return nil
		} else if err != nil {
			return err
		}
	}

	// reset the outflow limit of token
	quotaOfIBCDenom.OutflowSum = quotaOfIBCDenom.OutflowSum.Sub(exchangePrice)
	k.SetDenomQuota(ctx, *quotaOfIBCDenom)

	// reset the total outflow sum
	totalOutflowSum := k.GetTotalOutflowSum(ctx)
	k.SetTotalOutflowSum(ctx, totalOutflowSum.Sub(exchangePrice))
	return nil
}

// GetFundsFromPacket returns transfer amount and denom
func (k Keeper) GetFundsFromPacket(packet exported.PacketI) (sdkmath.Int, string, error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return sdkmath.Int{}, "", err
	}

	amount, ok := sdkmath.NewIntFromString(packetData.Amount)
	if !ok {
		return sdkmath.Int{}, "", sdkerrors.ErrInvalidRequest.Wrapf("invalid transfer amount %s", packetData.Amount)
	}

	return amount, k.GetLocalDenom(packetData.Denom), nil
}

// GetLocalDenom retruns ibc denom
// Expected denoms in the following cases:
//
// send non-native: transfer/channel-0/denom -> ibc/xxx
// send native: denom -> denom
// recv (B)non-native: denom
// recv (B)native: transfer/channel-0/denom
func (k Keeper) GetLocalDenom(denom string) string {
	if strings.HasPrefix(denom, "transfer/") {
		denomTrace := transfertypes.ParseDenomTrace(denom)
		return denomTrace.IBCDenom()
	}

	return denom
}
