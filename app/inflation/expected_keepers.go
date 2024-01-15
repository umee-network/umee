package inflation

import (
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
)

type MintKeeper interface {
	mintkeeper.Keeper
}
