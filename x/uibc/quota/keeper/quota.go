package keeper

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	"github.com/umee-network/umee/v4/x/uibc"
)

// GetQuotaOfIBCDenoms returns quota of all registered ibc denoms.
func (k Keeper) GetQuotaOfIBCDenoms(ctx sdk.Context) ([]uibc.Quota, error) {
	var quotaOfIBCDenoms []uibc.Quota

	prefix := uibc.KeyPrefixForIBCDenom
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

// GetQuotaOfIBCDenom retunes the rate limits of ibc denom.
func (k Keeper) GetQuotaOfIBCDenom(ctx sdk.Context, ibcDenom string) (*uibc.Quota, error) {
	rate, err := k.getQuotaOfIBCDenom(ctx, ibcDenom)
	if err != nil {
		return nil, err
	}

	return rate, nil
}

// getQuotaOfIBCDenom retunes the quota of ibc denom.
func (k Keeper) getQuotaOfIBCDenom(ctx sdk.Context, ibcDenom string) (*uibc.Quota, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.CreateKeyForQuotaOfIBCDenom(ibcDenom))
	if bz == nil {
		return nil, uibc.ErrNoQuotaForIBCDenom
	}

	var quotaOfIBCDenom uibc.Quota
	if err := k.cdc.Unmarshal(bz, &quotaOfIBCDenom); err != nil {
		return nil, err
	}

	return &quotaOfIBCDenom, nil
}

// SetQuotaOfIBCDenoms save the rate limits of ibc denoms into store.
func (k Keeper) SetQuotaOfIBCDenoms(ctx sdk.Context, quotaOfIBCDenoms []uibc.Quota) error {
	for _, quotaOfIBCDenom := range quotaOfIBCDenoms {
		if err := k.SetQuotaOfIBCDenom(ctx, quotaOfIBCDenom); err != nil {
			return err
		}
	}

	return nil
}

// SetQuotaOfIBCDenom save the quota of ibc denom into store.
func (k Keeper) SetQuotaOfIBCDenom(ctx sdk.Context, quotaOfIBCDenom uibc.Quota) error {
	store := ctx.KVStore(k.storeKey)
	key := uibc.CreateKeyForQuotaOfIBCDenom(quotaOfIBCDenom.IbcDenom)

	bz, err := k.cdc.Marshal(&quotaOfIBCDenom)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

// GetTotalOutflowSum returns the total outflow of ibc-transfer amount.
func (k Keeper) GetTotalOutflowSum(ctx sdk.Context) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.KeyTotalOutflowSum)
	return sdk.MustNewDecFromStr(string(bz))
}

// SetTotalOutflowSum save the total outflow of ibc-transfer amount.
func (k Keeper) SetTotalOutflowSum(ctx sdk.Context, amount sdk.Dec) error {
	store := ctx.KVStore(k.storeKey)

	store.Set(uibc.KeyTotalOutflowSum, []byte(amount.String()))

	return nil
}

// SetExpire save the quota expires time of ibc denom into store.
func (k Keeper) SetExpire(ctx sdk.Context, expires time.Time) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := expires.MarshalBinary()
	if err != nil {
		return err
	}

	store.Set(uibc.QuotaExpiresKey, bz)

	return nil
}

// GetExpire returns ibc-transfer quota expires time.
func (k Keeper) GetExpire(ctx sdk.Context) (*time.Time, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.QuotaExpiresKey)
	if bz == nil {
		return nil, nil
	}
	quotaExpires := ctx.BlockTime()
	if err := quotaExpires.UnmarshalBinary(bz); err != nil {
		return &quotaExpires, err
	}
	return &quotaExpires, nil
}

// CheckAndUpdateQuota checks the quota of ibc-transfer of denom
func (k Keeper) CheckAndUpdateQuota(ctx sdk.Context, denom string, sendAmount string) error {
	params := k.GetParams(ctx)

	quotaOfIBCDenom, err := k.getQuotaOfIBCDenom(ctx, denom)
	if err != nil {
		return err
	}

	if quotaOfIBCDenom == nil {
		return nil
	}

	amount, err := strconv.ParseInt(sendAmount, 10, 64)
	if err != nil {
		return err
	}

	// get the registered token settings from leverage
	// denom => ibc/XXXX
	tokenSettings, err := k.leverageKeeper.GetTokenSettings(ctx, denom)
	if err != nil {
		return nil
	}

	// get the exchange rate of denom (ibc/XXXX)in USD from base denom
	exchangeRate, err := k.oracleKeeper.GetExchangeRateBase(ctx, denom)
	if err != nil {
		return err
	}

	sendingAmount := sdk.NewDec(amount).Quo(sdk.NewDec(10).Power(uint64(tokenSettings.Exponent)))
	amountInUSD := exchangeRate.Mul(sendingAmount)

	// checking token quota
	quotaOfIBCDenom.OutflowSum = quotaOfIBCDenom.OutflowSum.Add(amountInUSD)
	if quotaOfIBCDenom.OutflowSum.GT(params.TokenQuota) {
		return uibc.ErrQuotaExceeded
	}

	// checking total outflow quota
	totalOutflowSum := k.GetTotalOutflowSum(ctx).Add(amountInUSD)
	if totalOutflowSum.GT(params.TotalQuota) {
		return uibc.ErrQuotaExceeded
	}

	// update the per token outflow sum
	if err := k.SetQuotaOfIBCDenom(ctx, *quotaOfIBCDenom); err != nil {
		return err
	}

	// updating the total outflow sum
	return k.SetTotalOutflowSum(ctx, totalOutflowSum)
}

// UndoUpdateQuota undo the quota of ibc denom
func (k Keeper) UndoUpdateQuota(ctx sdk.Context, denom, sendAmount string) error {
	quotaOfIBCDenom, err := k.getQuotaOfIBCDenom(ctx, denom)
	if err != nil {
		return err
	}

	if quotaOfIBCDenom == nil {
		return nil
	}

	// get the registered token settings from leverage
	tokenSettings, err := k.leverageKeeper.GetTokenSettings(ctx, denom)
	if err != nil {
		return nil
	}

	amount, err := strconv.ParseInt(sendAmount, 10, 64)
	if err != nil {
		return err
	}

	// get the exchange rate of denom (ibc/XXXX)in USD from base denom
	exchangeRate, err := k.oracleKeeper.GetExchangeRateBase(ctx, denom)
	if err != nil {
		return err
	}

	sendingAmount := sdk.NewDec(amount).Quo(sdk.NewDec(10).Power(uint64(tokenSettings.Exponent)))
	amountInUSD := exchangeRate.Mul(sendingAmount)
	// reset the outflow limit of per token
	quotaOfIBCDenom.OutflowSum = quotaOfIBCDenom.OutflowSum.Sub(amountInUSD)
	if err := k.SetQuotaOfIBCDenom(ctx, *quotaOfIBCDenom); err != nil {
		return err
	}

	// reset the total outflow sum
	totalOutflowSum := k.GetTotalOutflowSum(ctx)
	return k.SetTotalOutflowSum(ctx, totalOutflowSum.Sub(amountInUSD))
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
