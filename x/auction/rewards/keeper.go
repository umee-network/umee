package rewards

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperBuilder struct{}

func (kb KeeperBuilder) Keeper(ctx *sdk.Context) Keeper {
	return Keeper{
		ctx: ctx,
	}
}

type Keeper struct {
	ctx *sdk.Context
}
