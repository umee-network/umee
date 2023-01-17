package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v5/modules/apps/29-fee/types"
	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"

	"github.com/umee-network/umee/v4/x/uibc"
)

type Keeper struct {
	storeKey       storetypes.StoreKey
	cdc            codec.BinaryCodec
	leverageKeeper uibc.LeverageKeeper
	ics4Wrapper    porttypes.ICS4Wrapper
}

func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, ics4Wrapper types.ICS4Wrapper, leverageKeeper uibc.LeverageKeeper,
) Keeper {

	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		ics4Wrapper:    ics4Wrapper,
		leverageKeeper: leverageKeeper,
	}
}

// UpdateQuotaParams update the ibc-transfer quota params for ibc denoms
func (k Keeper) UpdateQuotaParams(ctx sdk.Context, totalQuota, quotaPerDenom sdk.Dec, quotaDuration time.Duration,
) error {
	params := k.GetParams(ctx)

	params.TokenQuota = totalQuota
	params.QuotaDuration = quotaDuration
	params.TokenQuota = quotaPerDenom

	return k.SetParams(ctx, params)
}

// SetIBCPause update the ibc pause status in module params.
func (k Keeper) SetIBCPause(ctx sdk.Context, ibcStatus uibc.IBCTransferStatus) error {
	params := k.GetParams(ctx)
	if params.IbcPause == ibcStatus {
		return uibc.ErrIBCPauseStatus.Wrapf("ibc-transfer status already have same status %s", ibcStatus.String())
	}

	params.IbcPause = ibcStatus

	return k.SetParams(ctx, params)
}
