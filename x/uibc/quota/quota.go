package quota

import (
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/util/store"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc"
)

var ten = sdk.MustNewDecFromStr("10")

// GetAllOutflows returns sum of outflows of all tokens in USD value.
func (k Keeper) GetAllOutflows() (sdk.DecCoins, error) {
	iter := sdk.KVStorePrefixIterator(k.store, keyPrefixDenomOutflows)
	return store.LoadAllDecCoins(iter, len(keyPrefixDenomOutflows))
}

// GetTokenOutflows returns sum of denom outflows in USD value in the DecCoin structure.
func (k Keeper) GetTokenOutflows(denom string) sdk.DecCoin {
	// When token outflow is not stored in store it will return 0
	amount, _ := store.GetDec(k.store, keyTokenOutflow(denom), "total_outflow")
	return sdk.NewDecCoinFromDec(denom, amount)
}

// SetTokenOutflows saves provided updated IBC outflows as a pair: USD value, denom name in the
// DecCoin structure.
func (k Keeper) SetTokenOutflows(outflows sdk.DecCoins) {
	for _, q := range outflows {
		k.SetTokenOutflow(q)
	}
}

// SetOutflowSum save the total outflow of ibc-transfer amount.
func (k Keeper) SetOutflowSum(amount sdk.Dec) {
	err := store.SetDec(k.store, keyOutflowSum, amount, "total_outflow_sum")
	util.Panic(err)
}

// GetOutflowSum returns the total outflow of ibc-transfer amount.
func (k Keeper) GetOutflowSum() sdk.Dec {
	// When total outflow is not stored in store it will return 0
	amount, _ := store.GetDec(k.store, keyOutflowSum, "total_outflow")
	return amount
}

// SetTokenOutflow save the outflows of denom into store.
func (k Keeper) SetTokenOutflow(outflow sdk.DecCoin) {
	key := keyTokenOutflow(outflow.Denom)
	err := store.SetDec(k.store, key, outflow.Amount, "total_outflow")
	util.Panic(err)
}

// GetAllInflows returns inflows of all registered tokens in USD value.
func (k Keeper) GetAllInflows() (sdk.DecCoins, error) {
	iter := sdk.KVStorePrefixIterator(k.store, keyPrefixDenomInflows)
	return store.LoadAllDecCoins(iter, len(keyPrefixDenomInflows))
}

// SetTokenInflows saves provided updated IBC inflows as a pair: USD value, denom name in the
// DecCoin structure.
func (k Keeper) SetTokenInflows(inflows sdk.DecCoins) {
	for _, q := range inflows {
		k.SetTokenInflow(q)
	}
}

// SetTokenInflow save the inflow of denom into store.
func (k Keeper) SetTokenInflow(inflow sdk.DecCoin) {
	key := keyTokenInflow(inflow.Denom)
	err := store.SetDec(k.store, key, inflow.Amount, "token_inflow")
	util.Panic(err)
}

// GetTokenInflow returns the inflow of denom from store.
func (k Keeper) GetTokenInflow(denom string) sdk.DecCoin {
	key := keyTokenInflow(denom)
	// When token inflow is not stored in store it will return 0
	amount, _ := store.GetDec(k.store, key, "token_inflow")
	return sdk.NewDecCoinFromDec(denom, amount)
}

// GetInflowSum returns the total inflow of ibc-transfer amount.
func (k Keeper) GetInflowSum() sdk.Dec {
	// When total inflow is not stored in store it will return 0
	amount, _ := store.GetDec(k.store, keyInflowSum, "total_inflows")
	return amount
}

// SetInflowSum save the total inflow of ibc-transfer amount.
func (k Keeper) SetInflowSum(amount sdk.Dec) {
	err := store.SetDec(k.store, keyInflowSum, amount, "total_inflows")
	util.Panic(err)
}

// SetExpire save the quota expire time of ibc denom into.
func (k Keeper) SetExpire(expires time.Time) error {
	return store.SetBinValue(k.store, keyQuotaExpires, &expires, "expire")
}

// GetExpire returns ibc-transfer quota expires time.
func (k Keeper) GetExpire() (*time.Time, error) {
	return store.GetBinValue[*time.Time](k.store, keyQuotaExpires, "expire")
}

// ResetAllQuotas will zero the ibc-transfer quotas
func (k Keeper) ResetAllQuotas() error {
	qd := k.GetParams().QuotaDuration
	newExpires := k.blockTime.Add(qd)
	if err := k.SetExpire(newExpires); err != nil {
		return err
	}
	zero := sdk.NewDec(0)
	// outflows
	k.SetOutflowSum(zero)
	ps := k.PrefixStore(keyPrefixDenomOutflows)
	store.DeleteByPrefixStore(ps)

	// inflows
	k.SetInflowSum(zero)
	ps = k.PrefixStore(keyPrefixDenomInflows)
	store.DeleteByPrefixStore(ps)
	return nil
}

// CheckAndUpdateQuota checks if adding a newOutflow doesn't exceed the max quota and
// updates the current quota metrics.
func (k Keeper) CheckAndUpdateQuota(denom string, newOutflow sdkmath.Int) error {
	params := k.GetParams()
	exchangePrice, err := k.getExchangePrice(denom, newOutflow)
	if err != nil {
		if ltypes.ErrNotRegisteredToken.Is(err) {
			return nil
		} else if err != nil {
			return err
		}
	}

	o := k.GetTokenOutflows(denom)
	o.Amount = o.Amount.Add(exchangePrice)
	inToken := k.GetTokenInflow(denom)
	if !params.TokenQuota.IsZero() {
		if o.Amount.GT(params.TokenQuota) ||
			o.Amount.GT(params.InflowOutflowTokenQuotaBase.Add((params.InflowOutflowQuotaRate.Mul(inToken.Amount)))) {
			return uibc.ErrQuotaExceeded
		}
	}

	// Allow outflow either of two conditions
	// 1. Outflow Sum <= Total Outflow Quota
	// 2. OR Outflow Sum <= params.InflowOutflowQuotaBase + (params.InflowOutflowQuotaRate * sum of all inflows)
	outflowSum := k.GetOutflowSum().Add(exchangePrice)
	inflowSum := k.GetInflowSum()
	if !params.TotalQuota.IsZero() {
		if outflowSum.GT(params.TotalQuota) ||
			outflowSum.GT(params.InflowOutflowQuotaBase.Add(inflowSum.Mul(params.InflowOutflowQuotaRate))) {
			return uibc.ErrQuotaExceeded
		}
	}
	k.SetTokenOutflow(o)
	k.SetOutflowSum(outflowSum)

	return nil
}

func (k Keeper) getExchangePrice(denom string, amount sdkmath.Int) (sdk.Dec, error) {
	transferCoin := sdk.NewCoin(denom, amount)
	var (
		err          error
		exchangeRate sdk.Dec
	)

	// convert to base asset if it is `uToken`
	if coin.HasUTokenPrefix(denom) {
		// NOTE: to avoid ctx, we can use similar approach: create a leverage keeper builder
		transferCoin, err = k.leverage.ToToken(*k.ctx, transferCoin)
		if err != nil {
			return sdk.Dec{}, err
		}
	}

	ts, err := k.leverage.GetTokenSettings(*k.ctx, transferCoin.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}

	// get the exchange price (eg: UMEE) in USD from oracle using SYMBOL Denom eg: `UMEE` (uumee)
	exchangeRate, err = k.oracle.Price(*k.ctx, strings.ToUpper(ts.SymbolDenom))
	if err != nil {
		return sdk.Dec{}, err
	}
	// calculate total exchange rate
	powerReduction := ten.Power(uint64(ts.Exponent))
	return sdk.NewDecFromInt(transferCoin.Amount).Quo(powerReduction).Mul(exchangeRate), nil
}

// UndoUpdateQuota subtracts `amount` from quota metric of the ibc denom.
func (k Keeper) UndoUpdateQuota(denom string, amount sdkmath.Int) error {
	o := k.GetTokenOutflows(denom)
	exchangePrice, err := k.getExchangePrice(denom, amount)
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
	k.SetTokenOutflow(o)

	outflowSum := k.GetOutflowSum()
	k.SetOutflowSum(outflowSum.Sub(exchangePrice))
	return nil
}

// RecordIBCInflow will save the inflow amount if token is registered otherwise it will skip
func (k Keeper) RecordIBCInflow(packet channeltypes.Packet, denom, amount string,
) exported.Acknowledgement {
	// if chain is recevier and sender chain is source then we need create ibc_denom (ibc/hash(channel,denom)) to
	// check ibc_denom is exists in leverage token registry
	if !ics20types.SenderChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), denom) {
		// since SendPacket did not prefix the denomination, we must prefix denomination here
		sourcePrefix := ics20types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
		// NOTE: sourcePrefix contains the trailing "/"
		prefixedDenom := sourcePrefix + denom
		// construct the denomination trace from the full raw denomination and get the ibc_denom
		ibcDenom := ics20types.ParseDenomTrace(prefixedDenom).IBCDenom()
		ts, err := k.leverage.GetTokenSettings(*k.ctx, ibcDenom)
		if err != nil {
			if ltypes.ErrNotRegisteredToken.Is(err) {
				return nil // skip recording inflow if the token is not registered
			}
			k.ctx.Logger().Error("can't get x/leverage token settings", "error", err)
			return channeltypes.NewErrorAcknowledgement(err)
		}

		// get the exchange price (eg: UMEE) in USD from oracle using SYMBOL Denom eg: `UMEE`
		exchangeRate, err := k.oracle.Price(*k.ctx, strings.ToUpper(ts.SymbolDenom))
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		// calculate total exchange rate
		powerReduction := ten.Power(uint64(ts.Exponent))
		inflowInUSD := sdk.MustNewDecFromStr(amount).Quo(powerReduction).Mul(exchangeRate)

		tokenInflow := k.GetTokenInflow(ibcDenom)
		tokenInflow.Amount = tokenInflow.Amount.Add(inflowInUSD)
		k.SetTokenInflow(tokenInflow)
		totalInflowSum := k.GetInflowSum()
		k.SetInflowSum(totalInflowSum.Add(inflowInUSD))
	}

	return nil
}
