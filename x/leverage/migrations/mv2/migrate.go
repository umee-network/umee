package mv2

import (
	// "github.com/cosmos/cosmos-sdk/codec"
	// cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"fmt"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/umee-network/umee/v3/x/leverage/keeper"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

// Implements Proposal Interface
var _ gov.Content = &UpdateRegistryProposal{}

func registerInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*gov.Content)(nil),
		&UpdateRegistryProposal{},
	)
}

// Migrates leverage module from v1 to v3
func MigrateToV2(ctx sdk.Context, k keeper.Keeper, gk govkeeper.Keeper) {
	// gk.RemoveFromActiveProposalQueue(ctx types.Context, proposalID uint64, endTime time.Time)
	proposals := gk.GetProposals(ctx)
	for _, p := range proposals {
		for _, m := range p.Messages {
			if m, ok := m.GetCachedValue().(*UpdateRegistryProposal); ok {
				fmt.Println("\nn================", m)
			}
		}
	}
}
