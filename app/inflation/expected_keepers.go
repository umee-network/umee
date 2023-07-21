package inflation

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	// ugov "github.com/umee-network/umee/v5/x/ugov"
)

type MintKeeper interface {
	SetParams(ctx sdk.Context, params minttypes.Params)
	GetParams(ctx sdk.Context) (params minttypes.Params)
	StakingTokenSupply(ctx sdk.Context) math.Int
	SetMinter(ctx sdk.Context, minter minttypes.Minter)
	GetMinter(ctx sdk.Context) (minter minttypes.Minter)
	BondedRatio(ctx sdk.Context) sdk.Dec
	MintCoins(ctx sdk.Context, newCoins sdk.Coins) error
	AddCollectedFees(ctx sdk.Context, fees sdk.Coins) error
}
