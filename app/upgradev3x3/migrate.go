package upgradev3x3

import (
	"github.com/cosmos/gogoproto/proto"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
)

type migrator struct {
	gov govkeeper.Keeper
}

// Creates migration handler for gov leverage proposals to new the gov system
// and MsgGovUpdateRegistry type.
func Migrator(gk govkeeper.Keeper, _ cdctypes.InterfaceRegistry) module.MigrationHandler {
	// note: content was removed, use v3.3 for implementation details.
	m := migrator{gk}
	return m.migrate
}

func (m migrator) migrate(_ sdk.Context) error {
	panic("use v3.3 binary")
}
