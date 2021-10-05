package leverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/umee-network/umee/x/leverage/keeper"
	"github.com/umee-network/umee/x/leverage/types"
)

func NewUpdateRegistryProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.UpdateRegistryProposal:
			return handleUpdateRegistryProposalHandler(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized proposal content type: %T", c)
		}
	}
}

func handleUpdateRegistryProposalHandler(ctx sdk.Context, k keeper.Keeper, p *types.UpdateRegistryProposal) error {
	for _, token := range p.Registry {
		k.SetRegisteredToken(ctx, token)
	}

	return nil
}
