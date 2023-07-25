package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/x/ugov"
)

type BKeeperI interface {
	Keeper(ctx *sdk.Context) IKeeper
}

type IKeeper interface {
	IParams
	ExportGenesis() *ugov.GenesisState
	InitGenesis(gs *ugov.GenesisState) error
}

type IParams interface {
	SetMinGasPrice(p sdk.DecCoin) error
	MinGasPrice() sdk.DecCoin
	SetEmergencyGroup(p sdk.AccAddress)
	EmergencyGroup() sdk.AccAddress
	SetInflationParams(lp ugov.InflationParams) error
	InflationParams() ugov.InflationParams
	SetInflationCycleEnd(startTime time.Time) error
	GetInflationCycleEnd() (time.Time, error)
}

var _ IKeeper = Keeper{}
