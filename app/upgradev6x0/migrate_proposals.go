package upgradev6x0

import (
	"fmt"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/v6/x/leverage/types"
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

	logger.Info("Proposals to migrate: ", "len", len(proposals))

	for _, p := range proposals {
		if len(p.Messages) != 1 {
			logger.Error("Ignoring, too many messages", "msgs", p.Messages)
			continue
		}

		cached := p.Messages[0].GetCachedValue()
		if oldUpdateRegistry, ok := cached.(*MsgGovUpdateRegistry); ok {
			logger.Info(">>>>>>> start migrating", p.Id)
			migrateMsgGovUpdateRegistry(ctx, p, oldUpdateRegistry, gk, logger)
		} else {
			logger.Error("Ignoring, not MsgGovUpdateRegistry",
				"msg_type", proto.MessageName(cached.(proto.Message)))
		}
	}
	return nil
}

func migrateMsgGovUpdateRegistry(
	ctx sdk.Context, p *govv1.Proposal, old *MsgGovUpdateRegistry, gk govkeeper.Keeper, logger log.Logger,
) {
	newMsg := types.MsgGovUpdateRegistry{
		Authority:    old.Authority,
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
	if p.Metadata != "" {
		p.Metadata = fmt.Sprintf("{\"title\":%q,\"summary\":%q}", old.Title, old.Description)
	}
	logger.Info(">>>>>>>>>>>> MIGRATING proposal:\n" + p.String())
	gk.SetProposal(ctx, *p)
}
