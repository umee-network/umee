package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

func (k Keeper) validateLendAsset(ctx sdk.Context, loan sdk.Coin) error {
	if !loan.IsValid() {
		return types.ErrInvalidAsset.Wrap(loan.String())
	}
	token, err := k.GetTokenSettings(ctx, loan.Denom)
	if err != nil {
		return err
	}
	return token.AssertLendEnabled()
}

func (k Keeper) validateBorrowAsset(ctx sdk.Context, borrow sdk.Coin) error {
	if !borrow.IsValid() {
		return types.ErrInvalidAsset.Wrap(borrow.String())
	}

	token, err := k.GetTokenSettings(ctx, borrow.Denom)
	if err != nil {
		return err
	}
	return token.AssertBorrowEnabled()
}
