package inflation

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type MintKeeper interface {
	StakingTokenSupply(ctx sdk.Context) math.Int
	SetParams(ctx sdk.Context, params minttypes.Params) error
}
