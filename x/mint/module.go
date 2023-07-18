package mint

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mk "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	abci "github.com/tendermint/tendermint/abci/types"

	ugovkeeper "github.com/umee-network/umee/v5/x/ugov/keeper"
)

var (
	_ module.AppModule = AppModule{}
)

type AppModule struct {
	mint.AppModule
	mintKeeper mk.Keeper
	ugovKB     ugovkeeper.Builder
}

func NewAppModule(cdc codec.Codec, keeper mk.Keeper, ak types.AccountKeeper, ugovkb ugovkeeper.Builder,
) AppModule {
	return AppModule{
		AppModule:  mint.NewAppModule(cdc, keeper, ak, nil),
		mintKeeper: keeper,
		ugovKB:     ugovkb,
	}
}

// Name implements module.AppModule.
func (AppModule) Name() string {
	return types.ModuleName
}

// BeginBlock executes all ABCI BeginBlock logic respective to the x/uibc module.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	BeginBlock(ctx, am.ugovKB.Keeper(&ctx), am.mintKeeper)
}
