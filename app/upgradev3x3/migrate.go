package upgradev3x3

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/gogo/protobuf/proto"

	"github.com/umee-network/umee/v6/x/leverage/types"
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

var done bool

func (m migrator) migrate(ctx sdk.Context) error {
	logger := ctx.Logger()
	if done {
		logger.Error("Migration already done")
		return nil
	}
	done = true

	proposals := m.gov.GetProposals(ctx)
	alreadyAdded := map[string]bool{
		"uume": true,
		"ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9": true, // "atom"
	}
	for _, p := range proposals {
		if len(p.Messages) != 1 {
			continue
		}

		cached := p.Messages[0].GetCachedValue()
		msg, ok := cached.(*govv1.MsgExecLegacyContent)
		if !ok {
			logger.Info("Ignoring, non legacy proposal", "msg_type", proto.MessageName(cached.(proto.Message)))
			continue
		}
		cached = msg.Content.GetCachedValue()
		lp, ok := cached.(*UpdateRegistryProposal)
		if !ok {
			logger.Info("Ignoring, not UpdateRegistryProposal",
				"msg_type", proto.MessageName(cached.(proto.Message)))
			continue
		}

		added := []types.Token{}
		updated := []types.Token{}
		for _, t := range lp.Registry {
			if alreadyAdded[t.BaseDenom] {
				updated = append(updated, t)
			} else {
				added = append(added, t)
				alreadyAdded[t.BaseDenom] = true
			}
		}
		newMsg := types.MsgGovUpdateRegistry{
			Authority:    msg.Authority,
			Title:        lp.Title,
			Description:  lp.Description,
			AddTokens:    added,
			UpdateTokens: updated,
		}
		var err error
		p.Messages[0], err = cdctypes.NewAnyWithValue(&newMsg)
		if err != nil {
			logger.Error("Can't pack ANY", err)
		}
		logger.Info("\n\nMIGRATING proposal:\n" + newMsg.String())
		m.gov.SetProposal(ctx, *p)
	}
	return nil
}
