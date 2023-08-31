package upgradev5x2

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"

	gravitytypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

var done bool

func Migrate(ctx sdk.Context, ir cdctypes.InterfaceRegistry, gk govkeeper.Keeper) error {
	logger := ctx.Logger()
	if done {
		logger.Error("Migration already done")
		return nil
	}
	done = true

	proposals := gk.GetProposals(ctx)
	for _, p := range proposals {
		if len(p.Messages) != 1 {
			logger.Debug("Ignoring, too many messages", "msgs", p.Messages)
			continue
		}

		cached := p.Messages[0].GetCachedValue()
		if oldUpdateRegistry, ok := cached.(*types.MsgGovUpdateRegistry); ok {
			migrateMsgGovUpdateRegistry(ctx, p, oldUpdateRegistry, gk, logger)
			continue
		}

		if oldLegacyContent, ok := cached.(*govv1.MsgExecLegacyContent); ok {
			if oldLegacyContent.Content.TypeUrl != "/gravity.v1.IBCMetadataProposal" {
				logger.Debug("Ignoring, not IBCMetadataProposal", "type_url", oldLegacyContent.Content.TypeUrl)
				continue
			}

			if err := migrateIBCMetadataProposal(ctx, ir, p, oldLegacyContent, gk, logger); err != nil {
				return err
			}
		}

		logger.Debug("Ignoring, not MsgGovUpdateRegistry nor MsgExecLegacyContent",
			"msg_type", proto.MessageName(cached.(proto.Message)))

	}
	return nil
}

func migrateMsgGovUpdateRegistry(
	ctx sdk.Context, p *govv1.Proposal, old *types.MsgGovUpdateRegistry, gk govkeeper.Keeper, logger log.Logger,
) {
	logger.Info(">>> MIGRATING proposal", "id", p.Id)
	if old.Title == "" && old.Description == "" {
		// nothing to set in metadata
		return
	}
	newMsg := types.MsgGovUpdateRegistry{
		Authority:    old.Authority,
		Title:        "",
		Description:  "",
		AddTokens:    old.AddTokens,
		UpdateTokens: old.UpdateTokens,
	}
	var err error
	p.Messages[0], err = cdctypes.NewAnyWithValue(&newMsg)
	if err != nil {
		logger.Error("Can't pack ANY", err)
		return
	}
	// overwrite only when metadata doesn't contain "title"
	if !strings.Contains(strings.ToLower(p.Metadata), "title") {
		p.Metadata = fmt.Sprintf("{\"title\":%q,\"summary\":%q}", old.Title, old.Description)
		return
	}

	gk.SetProposal(ctx, *p)
}

func migrateIBCMetadataProposal(
	ctx sdk.Context, ir cdctypes.InterfaceRegistry, p *govv1.Proposal, old *govv1.MsgExecLegacyContent, gk govkeeper.Keeper, logger log.Logger,
) error {
	logger.Info(">>> MIGRATING proposal", "id", p.Id)
	var gravityIBCProp *gravitytypes.IBCMetadataProposal
	err := ir.UnpackAny(old.Content, &gravityIBCProp)
	if err != nil {
		return err
	}

	content := govv1beta1.NewTextProposal(gravityIBCProp.Title, gravityIBCProp.String())
	msg, err := govv1.NewLegacyContent(content, "")
	if err != nil {
		return err
	}

	p.Messages[0], err = cdctypes.NewAnyWithValue(msg)
	if err != nil {
		return err
	}

	p.Metadata = fmt.Sprintf("{\"title\":%q,\"summary\":%q}", gravityIBCProp.Title, gravityIBCProp.String())
	gk.SetProposal(ctx, *p)
	return nil
}
