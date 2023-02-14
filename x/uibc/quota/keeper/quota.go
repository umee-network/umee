package keeper

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/umee-network/umee/v4/util/coin"
	ltypes "github.com/umee-network/umee/v4/x/leverage/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

// GetAllQuotas returns quota of all tokens.
func (k Keeper) GetAllQuotas(ctx sdk.Context) (sdk.DecCoins, error) {
	var quotas sdk.DecCoins
	store := k.PrefixStore(&ctx, uibc.KeyPrefixDenomQuota)
	iter := sdk.KVStorePrefixIterator(store, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var quotaOfIBCDenom = sdk.DecCoin{Denom: string(iter.Key())}
		if err := quotaOfIBCDenom.Amount.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}

		quotas = append(quotas, quotaOfIBCDenom)
	}

	return quotas, nil
}

// GetQuota retunes the rate limits of ibc denom.
func (k Keeper) GetQuota(ctx sdk.Context, ibcDenom string) (sdk.DecCoin, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.KeyTotalOutflows(ibcDenom))
	if bz == nil {
		return coin.ZeroDec(ibcDenom), nil
	}

	var d sdk.Dec
	err := d.Unmarshal(bz)

	return sdk.NewDecCoinFromDec(ibcDenom, d), err
}

// SetDenomQuotas save the updated token quota.
func (k Keeper) SetDenomQuotas(ctx sdk.Context, quotas sdk.DecCoins) {
	for _, q := range quotas {
		k.SetDenomQuota(ctx, q)
	}
}

// SetDenomQuota save the quota of denom into store.
func (k Keeper) SetDenomQuota(ctx sdk.Context, quota sdk.DecCoin) {
	store := ctx.KVStore(k.storeKey)
	key := uibc.KeyTotalOutflows(quota.Denom)
	bz, err := quota.Amount.Marshal()
	if err != nil {
		panic(fmt.Sprint("can't marshal quota: ", quota))
	}
	store.Set(key, bz)
}

// GetTotalOutflowSum returns the total outflow of ibc-transfer amount.
// TODO: remove to GetTotalOutflow
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

// ResetQuota will zero the ibc-transfer quotas
func (k Keeper) ResetQuota(ctx sdk.Context) error {
	qd := k.GetParams(ctx).QuotaDuration
	newExpires := ctx.BlockTime().Add(qd)
	if err := k.SetExpire(ctx, newExpires); err != nil {
		return err
	}
	zero := sdk.NewDec(0)
	zeroBz, err := zero.Marshal()
	if err != nil {
		return err
	}
	k.SetTotalOutflowSum(ctx, zero)
	store := k.PrefixStore(&ctx, uibc.KeyPrefixDenomQuota)
	iter := sdk.KVStorePrefixIterator(store, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		ibcDenom := iter.Key()
		store.Set(ibcDenom, zeroBz)
	}
	return nil
}

// CheckAndUpdateQuota checks if adding a newOutflow is doesn't exceed the max quota and
// updates the current quota metrics.
func (k Keeper) CheckAndUpdateQuota(ctx sdk.Context, denom string, newOutflow sdkmath.Int) error {
	params := k.GetParams(ctx)

	quota, err := k.GetQuota(ctx, denom)
	if err != nil {
		return err
	}

	exchangePrice, err := k.getExchangePrice(ctx, denom, newOutflow)
	if err != nil {
		// Note: skip the ibc-transfer quota checking if `denom` is not support by leverage
		// TODO: write test case for this
		if ltypes.ErrNotRegisteredToken.Is(err) {
			return nil
		} else if err != nil {
			return err
		}
	}

	// checking ibc-transfer token quota
	quota.Amount = quota.Amount.Add(exchangePrice)
	if quota.Amount.GT(params.TokenQuota) {
		return uibc.ErrQuotaExceeded
	}

	// checking total outflow quota
	totalOutflowSum := k.GetTotalOutflowSum(ctx).Add(exchangePrice)
	if totalOutflowSum.GT(params.TotalQuota) {
		return uibc.ErrQuotaExceeded
	}

	// update the per token outflow sum
	k.SetDenomQuota(ctx, quota)
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
	} else {
		if _, err := k.leverageKeeper.GetTokenSettings(ctx, denom); err != nil {
			return sdk.Dec{}, err
		}
	}

	// get the exchange price (eg: UMEE) in USD from oracle using base denom eg: `uumee`
	return k.oracleKeeper.HistoricAvgPrice(ctx, transferCoin.Denom)
}

// UndoUpdateQuota subtracts `amount` from quota metric of the ibc denom.
func (k Keeper) UndoUpdateQuota(ctx sdk.Context, denom string, amount sdkmath.Int) error {
	quota, err := k.GetQuota(ctx, denom)
	if err != nil {
		return err
	}

	// check the token is register or not
	exchangePrice, err := k.getExchangePrice(ctx, denom, amount)
	if err != nil {
		// Note: skip the ibc-transfer quota checking if `denom` is not support by leverage
		if ltypes.ErrNotRegisteredToken.Is(err) {
			return nil
		} else if err != nil {
			return err
		}
	}

	// We ignore the update if the result is negative (due to quota reset on epoch)
	quota.Amount = quota.Amount.Sub(exchangePrice)
	if quota.Amount.IsNegative() {
		return nil
	}

	k.SetDenomQuota(ctx, quota)

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
