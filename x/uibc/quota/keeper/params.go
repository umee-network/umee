package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

// SetParams sets the x/uibc module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params uibc.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(uibc.KeyPrefixParams, bz)

	return nil
}

// GetParams gets the x/uibc module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params uibc.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.KeyPrefixParams)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// UpdateQuotaParams update the ibc-transfer quota params for ibc denoms
func (k Keeper) UpdateQuotaParams(ctx sdk.Context, totalQuota, quotaPerDenom sdk.Dec, quotaDuration time.Duration,
) error {
	params := k.GetParams(ctx)
	params.TotalQuota = totalQuota
	params.QuotaDuration = quotaDuration
	params.TokenQuota = quotaPerDenom

	return k.SetParams(ctx, params)
}

// SetIBCStatus update the ibc pause status in module params.
func (k Keeper) SetIBCStatus(ctx sdk.Context, ibcStatus uibc.IBCTransferStatus) error {
	params := k.GetParams(ctx)
	params.IbcStatus = ibcStatus

	return k.SetParams(ctx, params)
}
