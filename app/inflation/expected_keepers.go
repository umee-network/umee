package inflation

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MintKeeper interface {
	StakingTokenSupply(ctx sdk.Context) math.Int
}
