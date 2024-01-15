package keeper

import (
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	store          store.KVStore
	bankKeeper     metoken.BankKeeper
	leverageKeeper metoken.LeverageKeeper
	oracleKeeper   metoken.OracleKeeper
	ugov           ugov.EmergencyGroupBuilder

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
		ctx:            ctx,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(k.ctx)
	return sdkCtx.Logger().With("module", "x/"+metoken.ModuleName)
}
