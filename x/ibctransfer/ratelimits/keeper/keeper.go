package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v5/modules/apps/29-fee/types"
	"github.com/umee-network/umee/v3/x/ibctransfer"
)

type Keeper struct {
	storeKey       storetypes.StoreKey
	cdc            codec.BinaryCodec
	paramSpace     paramtypes.Subspace
	oracleKeeper   ibctransfer.OracleKeeper
	leverageKeeper ibctransfer.LeverageKeeper
	ics4Wrapper    ibctransfer.ICS4Wrapper

	authority string // the gov module account
}

func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace, ics4Wrapper types.ICS4Wrapper,
	oracleKeeper ibctransfer.OracleKeeper, leverageKeeper ibctransfer.LeverageKeeper, authority string,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(ibctransfer.ParamKeyTable())
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		paramSpace:     paramSpace,
		ics4Wrapper:    ics4Wrapper,
		oracleKeeper:   oracleKeeper,
		leverageKeeper: leverageKeeper,
		authority:      authority,
	}
}

// UpdateIBCTansferStatus update the ibc pause status in module params.
func (k Keeper) UpdateIBCTansferStatus(ctx sdk.Context, ibcStatus bool) error {
	var ibcPause bool
	k.paramSpace.Get(ctx, ibctransfer.KeyIBCPause, &ibcPause)

	if ibcPause == ibcStatus {
		return ibctransfer.ErrIBCPauseStatus
	}

	// update the ibc status
	k.paramSpace.Set(ctx, ibctransfer.KeyIBCPause, ibcStatus)

	return nil
}
