package leverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/x/leverage/keeper"
	"github.com/umee-network/umee/x/leverage/types"
)

func handleMsgLendAsset(ctx sdk.Context, k keeper.Keeper, msg *types.MsgLendAsset) (*sdk.Result, error) {
	k.LendAsset(ctx, *msg)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgWithdrawAsset(ctx sdk.Context, k keeper.Keeper, msg *types.MsgWithdrawAsset) (*sdk.Result, error) {
	k.WithdrawAsset(ctx, *msg)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
