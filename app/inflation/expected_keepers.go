package inflation

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	ugovkeeper "github.com/umee-network/umee/v5/x/ugov/keeper"
)

type MintKeeper interface {
	StakingTokenSupply(ctx sdk.Context) math.Int
	SetParams(ctx sdk.Context, params minttypes.Params)
}

type UGovKeeper interface {
	ugovkeeper.IKeeper
}

type UGovBKeeperI interface {
	ugovkeeper.BKeeperI
}
