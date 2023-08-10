package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/v6/x/metoken"
)

// Builder constructs Keeper by preparing all related dependencies (notably the store).
type Builder struct {
	cdc            codec.Codec
	storeKey       storetypes.StoreKey
	bankKeeper     metoken.BankKeeper
	leverageKeeper metoken.LeverageKeeper
	oracleKeeper   metoken.OracleKeeper
}

// NewKeeperBuilder returns Builder object.
func NewKeeperBuilder(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bankKeeper metoken.BankKeeper,
	leverageKeeper metoken.LeverageKeeper,
	oracleKeeper metoken.OracleKeeper,
) Builder {
	return Builder{
		cdc:            cdc,
		storeKey:       storeKey,
		bankKeeper:     bankKeeper,
		leverageKeeper: leverageKeeper,
		oracleKeeper:   oracleKeeper,
	}
}

type Keeper struct {
	cdc            codec.Codec
	store          sdk.KVStore
	bankKeeper     metoken.BankKeeper
	leverageKeeper metoken.LeverageKeeper
	oracleKeeper   metoken.OracleKeeper

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
		ctx:            ctx,
	}
}

// Logger returns module Logger
func (k Keeper) Logger() log.Logger {
	return k.ctx.Logger().With("module", "x/"+metoken.ModuleName)
}
