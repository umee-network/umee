package ugov

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper interface {
	ParamsKeeper
	WithEmergencyGroup
	ExportGenesis() *GenesisState
	InitGenesis(gs *GenesisState) error
}

type ParamsKeeper interface {
	SetMinGasPrice(p sdk.DecCoin) error
	MinGasPrice() sdk.DecCoin
	SetEmergencyGroup(p sdk.AccAddress)
	SetInflationParams(lp InflationParams) error
	InflationParams() InflationParams
	SetInflationCycleEnd(startTime time.Time) error
	InflationCycleEnd() time.Time
}

type WithEmergencyGroup interface {
	EmergencyGroup() sdk.AccAddress
}

// Builder functions

type EmergencyGroupBuilder func(*sdk.Context) WithEmergencyGroup
type ParamsKeeperBuilder func(*sdk.Context) ParamsKeeper
