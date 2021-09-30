package leverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/umee-network/umee/x/leverage/keeper"
	"github.com/umee-network/umee/x/leverage/types"
)

func NewUpdateAssetsProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.UpdateAssetsProposal:
			return handleUpdateAssetsProposalHandler(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized proposal content type: %T", c)
		}
	}
}

func handleUpdateAssetsProposalHandler(ctx sdk.Context, k keeper.Keeper, p *types.UpdateAssetsProposal) error {
	for _, asset := range p.Assets {
		k.SetAsset(ctx, asset)
	}

	return nil
}
