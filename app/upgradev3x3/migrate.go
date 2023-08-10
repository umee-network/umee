package upgradev3x3

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Implements Proposal Interface
var _ gov.Content = &UpdateRegistryProposal{}

func registerInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*gov.Content)(nil),
		&UpdateRegistryProposal{},
	)
}

type migrator struct {
	gov govkeeper.Keeper
}

// Creates migration handler for gov leverage proposals to new the gov system
// and MsgGovUpdateRegistry type.
func Migrator(gk govkeeper.Keeper, registry cdctypes.InterfaceRegistry) module.MigrationHandler {
	registerInterfaces(registry)
	m := migrator{gk}
	return m.migrate
}

func (m migrator) migrate(ctx sdk.Context) error {
	panic("use v3.3 binary")
}
