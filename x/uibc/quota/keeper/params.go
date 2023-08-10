package keeper

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/uibc"
)

// SetParams sets the x/uibc module's parameters.
func (k Keeper) SetParams(params uibc.Params) error {
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	k.store.Set(keyParams, bz)

	return nil
}

// GetParams gets the x/uibc module's parameters.
func (k Keeper) GetParams() (params uibc.Params) {
	bz := k.store.Get(keyParams)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// UpdateQuotaParams update the ibc-transfer quota params for ibc denoms
func (k Keeper) UpdateQuotaParams(totalQuota, quotaPerDenom sdk.Dec, quotaDuration time.Duration,
	byEmergencyGroup bool) error {

	pOld := k.GetParams()
	pNew := pOld
	pNew.TotalQuota = totalQuota
	pNew.QuotaDuration = quotaDuration
	pNew.TokenQuota = quotaPerDenom
	if byEmergencyGroup {
		if err := validateEmergencyQuotaParamsUpdate(pOld, pNew); err != nil {
			return err
		}
	}

	return k.SetParams(pNew)
}

func validateEmergencyQuotaParamsUpdate(pOld, pNew uibc.Params) error {
	var errs []error
	if pNew.TotalQuota.GT(pOld.TotalQuota) || pNew.TokenQuota.GT(pOld.TokenQuota) {
		errs = append(errs, errors.New("Emergency Group can't increase TotalQuota nor TokenQuota"))
	}
	if pNew.QuotaDuration != pOld.QuotaDuration {
		errs = append(errs, errors.New("Emergency Group can't change QuotaDuration"))
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

// SetIBCStatus update the ibc-transfer status in module params.
func (k Keeper) SetIBCStatus(ibcStatus uibc.IBCTransferStatus) error {
	params := k.GetParams()
	params.IbcStatus = ibcStatus

	return k.SetParams(params)
}
