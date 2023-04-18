package msg

import (
	"context"

	"github.com/gogo/protobuf/proto"

	lvtypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// HandleSupply handles the Supply value of an address.
func (m UmeeMsg) HandleSupply(
	ctx context.Context,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgSupply{Supplier: m.Supply.Supplier, Asset: m.Supply.Asset}
	return s.Supply(ctx, req)
}

// HandleWithdraw handles the Withdraw value of an address.
func (m UmeeMsg) HandleWithdraw(
	ctx context.Context,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgWithdraw{Supplier: m.Withdraw.Supplier, Asset: m.Withdraw.Asset}
	return s.Withdraw(ctx, req)
}

// HandleCollateralize handles the enable selected uTokens as collateral.
func (m UmeeMsg) HandleCollateralize(
	ctx context.Context,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgCollateralize{Borrower: m.Collateralize.Borrower, Asset: m.Collateralize.Asset}
	return s.Collateralize(ctx, req)
}

// HandleDecollateralize handles the disable amount of an selected uTokens
// as collateral.
func (m UmeeMsg) HandleDecollateralize(
	ctx context.Context,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgDecollateralize{Borrower: m.Decollateralize.Borrower, Asset: m.Decollateralize.Asset}
	return s.Decollateralize(ctx, req)
}

// HandleBorrow handles the borrowing coins from the capital facility.
func (m UmeeMsg) HandleBorrow(
	ctx context.Context,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgBorrow{Borrower: m.Borrow.Borrower, Asset: m.Borrow.Asset}
	return s.Borrow(ctx, req)
}

// HandleRepay handles repaying borrowed coins to the capital facility.
func (m UmeeMsg) HandleRepay(
	ctx context.Context,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgRepay{Borrower: m.Repay.Borrower, Asset: m.Repay.Asset}
	return s.Repay(ctx, req)
}

// HandleLiquidate handles the repaying a different user's borrowed coins
// to the capital facility in exchange for some of their collateral.
func (m UmeeMsg) HandleLiquidate(
	ctx context.Context,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgLiquidate{
		Liquidator:  m.Liquidate.Liquidator,
		Borrower:    m.Liquidate.Borrower,
		Repayment:   m.Liquidate.Repayment,
		RewardDenom: m.Liquidate.RewardDenom,
	}
	return s.Liquidate(ctx, req)
}

// HandleSupplyCollateral handles the supply the assets and collateral their assets.
func (m UmeeMsg) HandleSupplyCollateral(
	ctx context.Context,
	s lvtypes.MsgServer,
) (proto.Message, error) {
	req := &lvtypes.MsgSupplyCollateral{Supplier: m.SupplyCollateral.Supplier, Asset: m.Supply.Asset}
	return s.SupplyCollateral(ctx, req)
}
