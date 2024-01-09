package keeper

import (
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/umee-network/umee/v6/x/ugov"

	"github.com/umee-network/umee/v6/x/metoken"
)

// Builder constructs Keeper by preparing all related dependencies (notably the store).
type Builder struct {
	cdc            codec.Codec
	storeKey       storetypes.StoreKey
	bankKeeper     metoken.BankKeeper
	leverageKeeper metoken.LeverageKeeper
	oracleKeeper   metoken.OracleKeeper
	ugov           ugov.EmergencyGroupBuilder
}

// NewKeeperBuilder returns Builder object.
func NewKeeperBuilder(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bankKeeper metoken.BankKeeper,
	leverageKeeper metoken.LeverageKeeper,
	oracleKeeper metoken.OracleKeeper,
	ugov ugov.EmergencyGroupBuilder,
) Builder {
	return Builder{
		cdc:            cdc,
		storeKey:       storeKey,
		bankKeeper:     bankKeeper,
		leverageKeeper: leverageKeeper,
		oracleKeeper:   oracleKeeper,
		ugov:           ugov,
	}
}

type Keeper struct {
	cdc            codec.Codec
	store          sdk.KVStore
	bankKeeper     metoken.BankKeeper
	leverageKeeper metoken.LeverageKeeper
	oracleKeeper   metoken.OracleKeeper
	ugov           ugov.EmergencyGroupBuilder
	meTokenAddr    sdk.AccAddress

	// TODO: ctx should be removed when we migrate leverageKeeper and oracleKeeper
	ctx *sdk.Context
}

// Keeper creates a new Keeper object
func (b Builder) Keeper(ctx *sdk.Context) Keeper {
	return Keeper{
		cdc:            b.cdc,
		store:          ctx.KVStore(b.storeKey),
		bankKeeper:     b.bankKeeper,
		leverageKeeper: b.leverageKeeper,
		oracleKeeper:   b.oracleKeeper,
		ugov:           b.ugov,
		meTokenAddr:    authtypes.NewModuleAddress(metoken.ModuleName),
		ctx:            ctx,
	}
}

// Logger returns module Logger
func (k Keeper) Logger() log.Logger {
	return k.ctx.Logger().With("module", "x/"+metoken.ModuleName)
}
