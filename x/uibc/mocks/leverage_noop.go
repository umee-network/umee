package mocks

import (
	"context"

	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
)

type lvgNoop struct{}

func NewLvgNoopMsgSrv() ltypes.MsgServer {
	return lvgNoop{}
}

func (l lvgNoop) Supply(context.Context, *ltypes.MsgSupply) (*ltypes.MsgSupplyResponse, error) {
	return nil, nil
}

func (l lvgNoop) Withdraw(context.Context, *ltypes.MsgWithdraw) (*ltypes.MsgWithdrawResponse, error) {
	return nil, nil
}

func (l lvgNoop) MaxWithdraw(context.Context, *ltypes.MsgMaxWithdraw) (*ltypes.MsgMaxWithdrawResponse, error) {
	return nil, nil
}

func (l lvgNoop) Collateralize(context.Context, *ltypes.MsgCollateralize,
) (*ltypes.MsgCollateralizeResponse, error) {
	return nil, nil
}

func (l lvgNoop) Decollateralize(context.Context, *ltypes.MsgDecollateralize,
) (*ltypes.MsgDecollateralizeResponse, error) {
	return nil, nil
}

func (l lvgNoop) Borrow(context.Context, *ltypes.MsgBorrow) (*ltypes.MsgBorrowResponse, error) {
	return nil, nil
}

func (l lvgNoop) MaxBorrow(context.Context, *ltypes.MsgMaxBorrow) (*ltypes.MsgMaxBorrowResponse, error) {
	return nil, nil
}

func (l lvgNoop) Repay(context.Context, *ltypes.MsgRepay) (*ltypes.MsgRepayResponse, error) {
	return nil, nil
}

func (l lvgNoop) Liquidate(context.Context, *ltypes.MsgLiquidate) (*ltypes.MsgLiquidateResponse, error) {
	return nil, nil
}

func (l lvgNoop) LeveragedLiquidate(context.Context, *ltypes.MsgLeveragedLiquidate,
) (*ltypes.MsgLeveragedLiquidateResponse, error) {
	return nil, nil
}

func (l lvgNoop) SupplyCollateral(context.Context, *ltypes.MsgSupplyCollateral,
) (*ltypes.MsgSupplyCollateralResponse, error) {
	return nil, nil
}

func (l lvgNoop) GovUpdateRegistry(context.Context, *ltypes.MsgGovUpdateRegistry,
) (*ltypes.MsgGovUpdateRegistryResponse, error) {
	return nil, nil
}

func (l lvgNoop) GovUpdateSpecialAssets(context.Context, *ltypes.MsgGovUpdateSpecialAssets,
) (*ltypes.MsgGovUpdateSpecialAssetsResponse, error) {
	return nil, nil
}

func (l lvgNoop) GovSetParams(context.Context, *ltypes.MsgGovSetParams,
) (*ltypes.MsgGovSetParamsResponse, error) {
	return nil, nil
}
