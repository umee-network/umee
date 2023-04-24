package keeper

import (
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	ltypes "github.com/umee-network/umee/v4/x/leverage/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

var ten = sdk.MustNewDecFromStr("10")

// GetAllOutflows returns sum of outflows of all tokens in USD value.
func (k Keeper) GetAllOutflows(ctx sdk.Context) (sdk.DecCoins, error) {
	var outflows sdk.DecCoins
	// creating PrefixStore upfront will remove the prefix from the key when running the iterator.
	store := k.PrefixStore(&ctx, uibc.KeyPrefixDenomOutflows)
	iter := sdk.KVStorePrefixIterator(store, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		o := sdk.DecCoin{Denom: string(iter.Key())}
		if err := o.Amount.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}
		outflows = append(outflows, o)
	}

	return outflows, nil
}

// GetTokenOutflows returns sum of denom outflows in USD value in the DecCoin structure.
func (k Keeper) GetTokenOutflows(ctx sdk.Context, denom string) (sdk.DecCoin, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.KeyTotalOutflows(denom))
	if bz == nil {
		return coin.ZeroDec(denom), nil
	}

	var d sdk.Dec
	err := d.Unmarshal(bz)

	return sdk.NewDecCoinFromDec(denom, d), err
}

// SetTokenOutflows saves provided updated IBC outflows as a pair: USD value, denom name in the
// DecCoin structure.
func (k Keeper) SetTokenOutflows(ctx sdk.Context, outflows sdk.DecCoins) {
	for _, q := range outflows {
		k.SetTokenOutflow(ctx, q)
	}
}

// SetTokenOutflow save the outflows of denom into store.
func (k Keeper) SetTokenOutflow(ctx sdk.Context, outflow sdk.DecCoin) {
	store := ctx.KVStore(k.storeKey)
	key := uibc.KeyTotalOutflows(outflow.Denom)
	bz, err := outflow.Amount.Marshal()
	if err != nil {
		panic(fmt.Sprint("can't marshal outflow: ", outflow))
	}
	store.Set(key, bz)
}

// GetTotalOutflow returns the total outflow of ibc-transfer amount.
func (k Keeper) GetTotalOutflow(ctx sdk.Context) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.KeyPrefixTotalOutflows)
	return sdk.MustNewDecFromStr(string(bz))
}

// SetTotalOutflowSum save the total outflow of ibc-transfer amount.
func (k Keeper) SetTotalOutflowSum(ctx sdk.Context, amount sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	store.Set(uibc.KeyPrefixTotalOutflows, []byte(amount.String()))
}

// SetExpire save the quota expire time of ibc denom into.
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
	now := time.Time{}
	if err := now.UnmarshalBinary(bz); err != nil {
		return nil, err
	}
	return &now, nil
}

// ResetAllQuotas will zero the ibc-transfer quotas
func (k Keeper) ResetAllQuotas(ctx sdk.Context) error {
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
	store := k.PrefixStore(&ctx, uibc.KeyPrefixDenomOutflows)
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
	exchangePrice, err := k.getExchangePrice(ctx, denom, newOutflow)
	if err != nil {
		if ltypes.ErrNotRegisteredToken.Is(err) {
			return nil
		} else if err != nil {
			return err
		}
	}

	o, err := k.GetTokenOutflows(ctx, denom)
	if err != nil {
		return err
	}
	o.Amount = o.Amount.Add(exchangePrice)
	if !params.TokenQuota.IsZero() && o.Amount.GT(params.TokenQuota) {
		return uibc.ErrQuotaExceeded
	}

	totalOutflowSum := k.GetTotalOutflow(ctx).Add(exchangePrice)
	if !params.TotalQuota.IsZero() && totalOutflowSum.GT(params.TotalQuota) {
		return uibc.ErrQuotaExceeded
	}

	k.SetTokenOutflow(ctx, o)
	k.SetTotalOutflowSum(ctx, totalOutflowSum)
	return nil
}

func (k Keeper) getExchangePrice(ctx sdk.Context, denom string, amount sdkmath.Int) (sdk.Dec, error) {
	transferCoin := sdk.NewCoin(denom, amount)
	var (
		err          error
		exchangeRate sdk.Dec
	)

	// convert to base asset if it is `uToken`
	if ltypes.HasUTokenPrefix(denom) {
		transferCoin, err = k.leverageKeeper.ExchangeUToken(ctx, transferCoin)
		if err != nil {
			return sdk.Dec{}, err
		}
	}

	ts, err := k.leverageKeeper.GetTokenSettings(ctx, transferCoin.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}

	// get the exchange price (eg: UMEE) in USD from oracle using SYMBOL Denom eg: `UMEE` (uumee)
	exchangeRate, err = k.oracle.Price(ctx, strings.ToUpper(ts.SymbolDenom))
	if err != nil {
		return sdk.Dec{}, err
	}
	// calculate total exchange rate
	powerReduction := ten.Power(uint64(ts.Exponent))
	return sdk.NewDecFromInt(transferCoin.Amount).Quo(powerReduction).Mul(exchangeRate), nil
}

// UndoUpdateQuota subtracts `amount` from quota metric of the ibc denom.
func (k Keeper) UndoUpdateQuota(ctx sdk.Context, denom string, amount sdkmath.Int) error {
	o, err := k.GetTokenOutflows(ctx, denom)
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
	o.Amount = o.Amount.Sub(exchangePrice)
	if o.Amount.IsNegative() {
		return nil
	}

	k.SetTokenOutflow(ctx, o)

	totalOutflowSum := k.GetTotalOutflow(ctx)
	k.SetTotalOutflowSum(ctx, totalOutflowSum.Sub(exchangePrice))
	return nil
}
