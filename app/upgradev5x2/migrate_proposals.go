package upgradev5x2

import (
	"fmt"
	"strings"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

var done bool

func Migrate(ctx sdk.Context, gk govkeeper.Keeper) error {
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
		} else {
			logger.Debug("Ignoring, not MsgGovUpdateRegistry",
				"msg_type", proto.MessageName(cached.(proto.Message)))
		}
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
