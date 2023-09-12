package msg

import (
	"context"

	"github.com/cosmos/gogoproto/proto"

	lvtypes "github.com/umee-network/umee/v6/x/leverage/types"
)

// HandleSupply handles the Supply value of an address.
func (m UmeeMsg) HandleSupply(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgSupply{Supplier: sender, Asset: m.Supply.Asset}
	return s.Supply(ctx, req)
}

// HandleWithdraw handles the Withdraw value of an address.
func (m UmeeMsg) HandleWithdraw(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgWithdraw{Supplier: sender, Asset: m.Withdraw.Asset}
	return s.Withdraw(ctx, req)
}

// HandleMaxWithdraw handles the maximum withdraw value of an address.
func (m UmeeMsg) HandleMaxWithdraw(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgMaxWithdraw{Supplier: sender, Denom: m.MaxWithdraw.Denom}
	return s.MaxWithdraw(ctx, req)
}

// HandleCollateralize handles the enable selected uTokens as collateral.
func (m UmeeMsg) HandleCollateralize(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgCollateralize{Borrower: sender, Asset: m.Collateralize.Asset}
	return s.Collateralize(ctx, req)
}

// HandleDecollateralize handles the disable amount of an selected uTokens
// as collateral.
func (m UmeeMsg) HandleDecollateralize(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgDecollateralize{Borrower: sender, Asset: m.Decollateralize.Asset}
	return s.Decollateralize(ctx, req)
}

// HandleBorrow handles the borrowing coins from the capital facility.
func (m UmeeMsg) HandleBorrow(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgBorrow{Borrower: sender, Asset: m.Borrow.Asset}
	return s.Borrow(ctx, req)
}

// HandleMaxBorrow handles the borrowing maximum coins from the capital facility.
func (m UmeeMsg) HandleMaxBorrow(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgMaxBorrow{Borrower: sender, Denom: m.MaxBorrow.Denom}
	return s.MaxBorrow(ctx, req)
}

// HandleRepay handles repaying borrowed coins to the capital facility.
func (m UmeeMsg) HandleRepay(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgRepay{Borrower: sender, Asset: m.Repay.Asset}
	return s.Repay(ctx, req)
}

// HandleLiquidate handles the repaying a different user's borrowed coins
// to the capital facility in exchange for some of their collateral.
func (m UmeeMsg) HandleLiquidate(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgLiquidate{
		Liquidator:  sender,
		Borrower:    m.Liquidate.Borrower,
		Repayment:   m.Liquidate.Repayment,
		RewardDenom: m.Liquidate.RewardDenom,
	}
	return s.Liquidate(ctx, req)
}

// HandleSupplyCollateral handles the supply the assets and collateral their assets.
func (m UmeeMsg) HandleSupplyCollateral(
	ctx context.Context, sender string,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgSupplyCollateral{Supplier: sender, Asset: m.SupplyCollateral.Asset}
	return s.SupplyCollateral(ctx, req)
}
