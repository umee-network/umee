package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

// SetParams sets the x/uibc module's parameters.
func (k Keeper) SetParams(params uibc.Params) error {
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	k.store.Set(uibc.KeyPrefixParams, bz)

	return nil
}

// GetParams gets the x/uibc module's parameters.
func (k Keeper) GetParams() (params uibc.Params) {
	bz := k.store.Get(uibc.KeyPrefixParams)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// UpdateQuotaParams update the ibc-transfer quota params for ibc denoms
func (k Keeper) UpdateQuotaParams(totalQuota, quotaPerDenom sdk.Dec, quotaDuration time.Duration,
) error {
	params := k.GetParams()
	params.TotalQuota = totalQuota
	params.QuotaDuration = quotaDuration
	params.TokenQuota = quotaPerDenom

	return k.SetParams(params)
}

// SetIBCStatus update the ibc-transfer status in module params.
func (k Keeper) SetIBCStatus(ibcStatus uibc.IBCTransferStatus) error {
	params := k.GetParams()
	params.IbcStatus = ibcStatus

	return k.SetParams(params)
}
