package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/util/sdkutil"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/leverage
// module.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

func (s msgServer) Supply(
	goCtx context.Context,
	msg *types.MsgSupply,
) (*types.MsgSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}
	received, err := s.keeper.Supply(ctx, supplierAddr, msg.Asset)
	if err != nil {
		return nil, err
	}

	// Fail here if MaxSupply is exceeded
	if err = s.keeper.checkMaxSupply(ctx, msg.Asset.Denom); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets supplied",
		"supplier", msg.Supplier,
		"supplied", msg.Asset.String(),
		"received", received.String(),
	)
	sdkutil.Emit(&ctx, &types.EventSupply{
		Supplier: msg.Supplier,
		Asset:    msg.Asset,
		Utoken:   received,
	})
	return &types.MsgSupplyResponse{
		Received: received,
	}, nil
}

func (s msgServer) Withdraw(
	goCtx context.Context,
	msg *types.MsgWithdraw,
) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}
	received, isFromCollateral, err := s.keeper.Withdraw(ctx, supplierAddr, msg.Asset)
	if err != nil {
		return nil, err
	}

	// Fail here if supplier ends up over their borrow limit under current or historic prices
	// Tolerates missing collateral prices if the rest of the borrower's collateral can cover all borrows
	if isFromCollateral {
		err = s.keeper.assertBorrowerHealth(ctx, supplierAddr)
		if err != nil {
			return nil, err
		}
	}

	// Ensure MinCollateralLiquidity is still satisfied after the transaction
	if err = s.keeper.checkCollateralLiquidity(ctx, received.Denom); err != nil {
		return nil, err
	}

	s.logWithdrawal(ctx, msg.Supplier, msg.Asset, received, "supplied assets withdrawn")
	return &types.MsgWithdrawResponse{
		Received: received,
	}, nil
}

func (s msgServer) MaxWithdraw(
	goCtx context.Context,
	msg *types.MsgMaxWithdraw,
) (*types.MsgMaxWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}
	if types.HasUTokenPrefix(msg.Denom) {
		return nil, types.ErrUToken
	}
	if _, err = s.keeper.GetTokenSettings(ctx, msg.Denom); err != nil {
		return nil, err
	}

	// If a price is missing for the borrower's collateral,
	// but not this uToken or any of their borrows, error
	// will be nil and the resulting value will be what
	// can safely be withdrawn even with missing prices.
	uToken, err := s.keeper.maxWithdraw(ctx, supplierAddr, msg.Denom)
	if err != nil {
		return nil, err
	}

	if uToken.IsZero() {
		zeroCoin := coin.Zero(msg.Denom)
		return &types.MsgMaxWithdrawResponse{Withdrawn: uToken, Received: zeroCoin}, nil
	}

	// Get the total available for uToken to prevent withdraws above this limit.
	uTokenTotalAvailable, err := s.keeper.AvailableCollateralLiquidity(ctx, msg.Denom)
	//todo: analyze exit path
	if err != nil {
		return nil, err
	}
	uToken.Amount = sdk.MinInt(uToken.Amount, uTokenTotalAvailable)

	received, isFromCollateral, err := s.keeper.Withdraw(ctx, supplierAddr, uToken)
	if err != nil {
		return nil, err
	}

	// Fail here if supplier ends up over their borrow limit under current or historic prices
	// Tolerates missing collateral prices if the rest of the borrower's collateral can cover all borrows
	if isFromCollateral {
		err = s.keeper.assertBorrowerHealth(ctx, supplierAddr)
		if err != nil {
			return nil, err
		}
	}

	// Ensure MinCollateralLiquidity is still satisfied after the transaction
	if err = s.keeper.checkCollateralLiquidity(ctx, received.Denom); err != nil {
		return nil, err
	}

	s.logWithdrawal(ctx, msg.Supplier, uToken, received, "maximum supplied assets withdrawn")
	return &types.MsgMaxWithdrawResponse{
		Withdrawn: uToken,
		Received:  received,
	}, nil
}

func (s msgServer) logWithdrawal(ctx sdk.Context, supplier string, redeemed, received sdk.Coin, desc string) {
	s.keeper.Logger(ctx).Debug(
		desc,
		"supplier", supplier,
		"redeemed", redeemed.String(),
		"received", received.String(),
	)
	sdkutil.Emit(&ctx, &types.EventWithdraw{
		Supplier: supplier,
		Utoken:   redeemed,
		Asset:    received,
	})
}

func (s msgServer) Collateralize(
	goCtx context.Context,
	msg *types.MsgCollateralize,
) (*types.MsgCollateralizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	if err := s.keeper.Collateralize(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	if err := s.keeper.checkCollateralLiquidity(ctx, types.ToTokenDenom(msg.Asset.Denom)); err != nil {
		return nil, err
	}

	// Fail here if collateral share restrictions are violated,
	// based on only collateral with known oracle prices
	if err := s.keeper.checkCollateralShare(ctx, msg.Asset.Denom); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral added",
		"borrower", msg.Borrower,
		"amount", msg.Asset.String(),
	)
	sdkutil.Emit(&ctx, &types.EventCollaterize{
		Borrower: msg.Borrower,
		Utoken:   msg.Asset,
	})
	return &types.MsgCollateralizeResponse{}, nil
}

func (s msgServer) SupplyCollateral(
	goCtx context.Context,
	msg *types.MsgSupplyCollateral,
) (*types.MsgSupplyCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}
	uToken, err := s.keeper.Supply(ctx, supplierAddr, msg.Asset)
	if err != nil {
		return nil, err
	}
	if err = s.keeper.Collateralize(ctx, supplierAddr, uToken); err != nil {
		return nil, err
	}

	// Fail here if MaxSupply is exceeded
	if err = s.keeper.checkMaxSupply(ctx, msg.Asset.Denom); err != nil {
		return nil, err
	}

	// Fail here if collateral liquidity restrictions are violated
	if err := s.keeper.checkCollateralLiquidity(ctx, msg.Asset.Denom); err != nil {
		return nil, err
	}

	// Fail here if collateral share restrictions are violated,
	// based on only collateral with known oracle prices
	if err := s.keeper.checkCollateralShare(ctx, uToken.Denom); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets supplied",
		"supplier", msg.Supplier,
		"supplied", msg.Asset.String(),
		"received", uToken.String(),
	)
	s.keeper.Logger(ctx).Debug(
		"collateral added",
		"borrower", msg.Supplier,
		"amount", uToken.String(),
	)
	sdkutil.Emit(&ctx, &types.EventSupply{
		Supplier: msg.Supplier,
		Asset:    msg.Asset,
		Utoken:   uToken,
	})
	sdkutil.Emit(&ctx, &types.EventCollaterize{
		Borrower: msg.Supplier,
		Utoken:   uToken,
	})
	return &types.MsgSupplyCollateralResponse{
		Collateralized: uToken,
	}, nil
}

func (s msgServer) Decollateralize(
	goCtx context.Context,
	msg *types.MsgDecollateralize,
) (*types.MsgDecollateralizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	if err := s.keeper.Decollateralize(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	// Fail here if borrower ends up over their borrow limit under current or historic prices
	// Tolerates missing collateral prices if the rest of the borrower's collateral can cover all borrows
	err = s.keeper.assertBorrowerHealth(ctx, borrowerAddr)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral removed",
		"borrower", msg.Borrower,
		"amount", msg.Asset.String(),
	)
	sdkutil.Emit(&ctx, &types.EventDecollaterize{
		Borrower: msg.Borrower,
		Utoken:   msg.Asset,
	})
	return &types.MsgDecollateralizeResponse{}, nil
}

func (s msgServer) Borrow(
	goCtx context.Context,
	msg *types.MsgBorrow,
) (*types.MsgBorrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	if err := s.keeper.Borrow(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	// Fail here if borrower ends up over their borrow limit under current or historic prices
	// Tolerates missing collateral prices if the rest of the borrower's collateral can cover all borrows
	err = s.keeper.assertBorrowerHealth(ctx, borrowerAddr)
	if err != nil {
		return nil, err
	}

	// Check MaxSupplyUtilization after transaction
	if err = s.keeper.checkSupplyUtilization(ctx, msg.Asset.Denom); err != nil {
		return nil, err
	}

	// Check MinCollateralLiquidity is still satisfied after the transaction
	if err = s.keeper.checkCollateralLiquidity(ctx, msg.Asset.Denom); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets borrowed",
		"borrower", msg.Borrower,
		"amount", msg.Asset.String(),
	)
	sdkutil.Emit(&ctx, &types.EventBorrow{
		Borrower: msg.Borrower,
		Asset:    msg.Asset,
	})
	return &types.MsgBorrowResponse{}, nil
}

func (s msgServer) MaxBorrow(
	goCtx context.Context,
	msg *types.MsgMaxBorrow,
) (*types.MsgMaxBorrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	// If a price is missing for the borrower's collateral,
	// but not this token or any of their borrows, error
	// will be nil and the resulting value will be what
	// can safely be borrowed even with missing prices.
	maxBorrow, err := s.keeper.maxBorrow(ctx, borrowerAddr, msg.Denom)
	if err != nil {
		return nil, err
	}
	if maxBorrow.IsZero() {
		return &types.MsgMaxBorrowResponse{Borrowed: coin.Zero(msg.Denom)}, nil
	}

	if err := s.keeper.Borrow(ctx, borrowerAddr, maxBorrow); err != nil {
		return nil, err
	}

	// Fail here if borrower ends up over their borrow limit under current or historic prices
	// Tolerates missing collateral prices if the rest of the borrower's collateral can cover all borrows
	err = s.keeper.assertBorrowerHealth(ctx, borrowerAddr)
	if err != nil {
		return nil, err
	}

	// Check MaxSupplyUtilization after transaction
	if err = s.keeper.checkSupplyUtilization(ctx, maxBorrow.Denom); err != nil {
		return nil, err
	}

	// Check MinCollateralLiquidity is still satisfied after the transaction
	if err = s.keeper.checkCollateralLiquidity(ctx, maxBorrow.Denom); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets borrowed",
		"borrower", msg.Borrower,
		"amount", maxBorrow.String(),
	)
	sdkutil.Emit(&ctx, &types.EventBorrow{
		Borrower: msg.Borrower,
		Asset:    maxBorrow,
	})
	return &types.MsgMaxBorrowResponse{
		Borrowed: maxBorrow,
	}, nil
}

func (s msgServer) Repay(
	goCtx context.Context,
	msg *types.MsgRepay,
) (*types.MsgRepayResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	repaid, err := s.keeper.Repay(ctx, borrowerAddr, msg.Asset)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"borrowed assets repaid",
		"borrower", msg.Borrower,
		"attempted", msg.Asset.String(),
		"repaid", repaid.String(),
	)
	sdkutil.Emit(&ctx, &types.EventRepay{
		Borrower: msg.Borrower,
		Repaid:   repaid,
	})
	return &types.MsgRepayResponse{
		Repaid: repaid,
	}, nil
}

func (s msgServer) Liquidate(
	goCtx context.Context,
	msg *types.MsgLiquidate,
) (*types.MsgLiquidateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidator, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		return nil, err
	}
	borrower, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	repaid, liquidated, reward, err := s.keeper.Liquidate(ctx, liquidator, borrower, msg.Repayment, msg.RewardDenom)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"unhealthy borrower liquidated",
		"liquidator", msg.Liquidator,
		"borrower", msg.Borrower,
		"attempted", msg.Repayment.String(),
		"repaid", repaid.String(),
		"liquidated", liquidated.String(),
		"reward", reward.String(),
	)
	sdkutil.Emit(&ctx, &types.EventLiquidate{
		Liquidator: msg.Liquidator,
		Borrower:   msg.Borrower,
		Liquidated: liquidated,
	})
	return &types.MsgLiquidateResponse{
		Repaid:     repaid,
		Collateral: liquidated,
		Reward:     reward,
	}, nil
}

// GovUpdateRegistry updates existing tokens with new settings
// or adds the new tokens to registry.
func (s msgServer) GovUpdateRegistry(
	goCtx context.Context,
	msg *types.MsgGovUpdateRegistry,
) (*types.MsgGovUpdateRegistryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	registeredTokens := s.keeper.GetAllRegisteredTokens(ctx)
	registeredTokenDenoms := make(map[string]bool)

	for _, token := range registeredTokens {
		registeredTokenDenoms[token.BaseDenom] = true
	}

	// update the token settings
	err := s.keeper.SaveOrUpdateTokenSettingsToRegistry(ctx, msg.UpdateTokens, registeredTokenDenoms, true)
	if err != nil {
		return &types.MsgGovUpdateRegistryResponse{}, err
	}

	// adds  the new token settings
	err = s.keeper.SaveOrUpdateTokenSettingsToRegistry(ctx, msg.AddTokens, registeredTokenDenoms, false)
	if err != nil {
		return &types.MsgGovUpdateRegistryResponse{}, err
	}

	// cleans blacklisted tokens from the registry if they have not been supplied
	err = s.keeper.CleanTokenRegistry(ctx)
	if err != nil {
		return &types.MsgGovUpdateRegistryResponse{}, err
	}

	return &types.MsgGovUpdateRegistryResponse{}, nil
}
