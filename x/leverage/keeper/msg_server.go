package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/leverage/types"
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
		err = s.keeper.assertBorrowerHealth(ctx, supplierAddr, sdk.OneDec())
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
	if coin.HasUTokenPrefix(msg.Denom) {
		return nil, types.ErrUToken
	}
	if _, err = s.keeper.GetTokenSettings(ctx, msg.Denom); err != nil {
		return nil, err
	}

	// If a price is missing for the borrower's collateral,
	// but not this uToken or any of their borrows, error
	// will be nil and the resulting value will be what
	// can safely be withdrawn even with missing prices.
	uToken, userSpendableUtokens, err := s.keeper.userMaxWithdraw(ctx, supplierAddr, msg.Denom)
	if err != nil {
		return nil, err
	}

	if uToken.IsZero() {
		zeroCoin := coin.Zero(msg.Denom)
		return &types.MsgMaxWithdrawResponse{Withdrawn: uToken, Received: zeroCoin}, nil
	}

	// Get the total available for uToken to prevent withdraws above this limit.
	uTokenTotalAvailable, err := s.keeper.ModuleMaxWithdraw(ctx, userSpendableUtokens)
	if err != nil {
		return nil, err
	}

	// If zero uTokens are available from liquidity and collateral, nothing can be withdrawn.
	if uTokenTotalAvailable.IsZero() {
		zeroCoin := coin.Zero(msg.Denom)
		zeroUcoin := coin.Zero(uToken.Denom)
		return &types.MsgMaxWithdrawResponse{Withdrawn: zeroUcoin, Received: zeroCoin}, nil
	}

	// Use the minimum of the user's max withdraw based on borrows and the module's max withdraw based on liquidity
	uToken.Amount = sdk.MinInt(uToken.Amount, uTokenTotalAvailable)

	// Proceed to withdraw.
	received, isFromCollateral, err := s.keeper.Withdraw(ctx, supplierAddr, uToken)
	if err != nil {
		return nil, err
	}

	// Fail here if supplier ends up over their borrow limit under current or historic prices
	// Tolerates missing collateral prices if the rest of the borrower's collateral can cover all borrows
	if isFromCollateral {
		err = s.keeper.assertBorrowerHealth(ctx, supplierAddr, sdk.OneDec())
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

	if err := s.keeper.checkCollateralLiquidity(ctx, coin.StripUTokenDenom(msg.Asset.Denom)); err != nil {
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
	err = s.keeper.assertBorrowerHealth(ctx, borrowerAddr, sdk.OneDec())
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
	err = s.keeper.assertBorrowerHealth(ctx, borrowerAddr, sdk.OneDec())
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
	userMaxBorrow, err := s.keeper.userMaxBorrow(ctx, borrowerAddr, msg.Denom)
	if err != nil {
		return nil, err
	}
	if userMaxBorrow.IsZero() {
		return &types.MsgMaxBorrowResponse{Borrowed: coin.Zero(msg.Denom)}, nil
	}

	// Get the max available to borrow from the module
	moduleMaxBorrow, err := s.keeper.moduleMaxBorrow(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}
	if moduleMaxBorrow.IsZero() {
		return &types.MsgMaxBorrowResponse{Borrowed: coin.Zero(msg.Denom)}, nil
	}

	// Select the minimum between user_max_borrow and module_max_borrow
	userMaxBorrow.Amount = sdk.MinInt(userMaxBorrow.Amount, moduleMaxBorrow)

	// Proceed to borrow
	if err := s.keeper.Borrow(ctx, borrowerAddr, userMaxBorrow); err != nil {
		return nil, err
	}

	// Fail here if borrower ends up over their borrow limit under current or historic prices
	// Tolerates missing collateral prices if the rest of the borrower's collateral can cover all borrows
	err = s.keeper.assertBorrowerHealth(ctx, borrowerAddr, sdk.OneDec())
	if err != nil {
		return nil, err
	}

	// Check MaxSupplyUtilization after transaction
	if err = s.keeper.checkSupplyUtilization(ctx, userMaxBorrow.Denom); err != nil {
		return nil, err
	}

	// Check MinCollateralLiquidity is still satisfied after the transaction
	if err = s.keeper.checkCollateralLiquidity(ctx, userMaxBorrow.Denom); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets borrowed",
		"borrower", msg.Borrower,
		"amount", moduleMaxBorrow.String(),
	)
	sdkutil.Emit(&ctx, &types.EventBorrow{
		Borrower: msg.Borrower,
		Asset:    userMaxBorrow,
	})
	return &types.MsgMaxBorrowResponse{
		Borrowed: userMaxBorrow,
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

func (s msgServer) LeveragedLiquidate(
	goCtx context.Context,
	msg *types.MsgLeveragedLiquidate,
) (*types.MsgLeveragedLiquidateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidator, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		return nil, err
	}
	borrower, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	repaid, reward, err := s.keeper.LeveragedLiquidate(ctx, liquidator, borrower, msg.RepayDenom, msg.RewardDenom)
	if err != nil {
		return nil, err
	}

	// Fail here if liquidator ends up over 80% their borrow limit under current or historic prices
	// Tolerates missing collateral prices if the rest of the liquidator's collateral can cover all borrows
	err = s.keeper.assertBorrowerHealth(ctx, liquidator, sdk.MustNewDecFromStr("0.8"))
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"unhealthy borrower leverage-liquidated",
		"liquidator", msg.Liquidator,
		"borrower", msg.Borrower,
		"repaid", repaid.String(),
		"reward", reward.String(),
	)
	sdkutil.Emit(&ctx, &types.EventLiquidate{
		Liquidator: msg.Liquidator,
		Borrower:   msg.Borrower,
		Liquidated: reward,
	})
	return &types.MsgLeveragedLiquidateResponse{
		Repaid: repaid,
		Reward: reward,
	}, nil
}

// GovUpdateRegistry updates existing tokens with new settings
// or adds the new tokens to registry.
func (s msgServer) GovUpdateRegistry(
	goCtx context.Context,
	msg *types.MsgGovUpdateRegistry,
) (*types.MsgGovUpdateRegistryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	regDenoms := make(map[string]types.Token)
	registeredTokens := s.keeper.GetAllRegisteredTokens(ctx)
	for _, token := range registeredTokens {
		regDenoms[token.BaseDenom] = token
	}

	byEmergencyGroup, err := checkers.EmergencyGroupAuthority(msg.Authority, s.keeper.ugov(&ctx))
	if err != nil {
		return nil, err
	}

	err = s.keeper.UpdateTokenRegistry(ctx, msg.UpdateTokens, msg.AddTokens, regDenoms, byEmergencyGroup)
	if err != nil {
		return nil, err
	}

	// cleans blacklisted tokens from the registry if they have not been supplied
	if err := s.keeper.CleanTokenRegistry(ctx); err != nil {
		return nil, err
	}

	return &types.MsgGovUpdateRegistryResponse{}, nil
}

// GovUpdateSpecialAssets adds, updates, or deletes special asset pairs.
func (s msgServer) GovUpdateSpecialAssets(
	goCtx context.Context,
	msg *types.MsgGovUpdateSpecialAssets,
) (*types.MsgGovUpdateSpecialAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, set := range msg.Sets {
		// each special asset set is decomposed into its component pairs and stored in state.
		// overrides existing pairs between assets in the new set.
		for _, a := range set.Assets {
			for _, b := range set.Assets {
				if a != b {
					pair := types.SpecialAssetPair{
						Collateral:       a,
						Borrow:           b,
						CollateralWeight: set.CollateralWeight,
					}
					// sets or overrides (or deletes on zero collateral weight) each pair
					if err := s.keeper.SetSpecialAssetPair(ctx, pair); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	// individual pairs are applied after sets, so they can override specific relationships
	// between assets.
	for _, pair := range msg.Pairs {
		// sets or overrides (or deletes on zero collateral weight) each pair
		if err := s.keeper.SetSpecialAssetPair(ctx, pair); err != nil {
			return nil, err
		}
	}

	return &types.MsgGovUpdateSpecialAssetsResponse{}, nil
}
