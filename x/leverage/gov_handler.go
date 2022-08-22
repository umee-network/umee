package leverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/umee-network/umee/v2/x/leverage/keeper"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

func NewUpdateRegistryProposalHandler(k keeper.Keeper) gov1b1.Handler {
	return func(ctx sdk.Context, content gov1b1.Content) error {
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
		if err := k.SetTokenSettings(ctx, token); err != nil {
			return err
		}
	}

	return nil
}
