package message

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	lvtypes "github.com/umee-network/umee/v2/x/leverage/types"
)

// HandleSupply handles the Supply value of an address.
func (m UmeeMsg) HandleSupply(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.Supply(sdk.WrapSDKContext(ctx), m.Supply)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Supply", err)}
	}

	return nil
}

// HandleWithdrawAsset handles the WithdrawAsset value of an address.
func (m UmeeMsg) HandleWithdrawAsset(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.WithdrawAsset(sdk.WrapSDKContext(ctx), m.WithdrawAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Withdraw Asset", err)}
	}

	return nil
}

// HandleAddCollateral handles the enable selected uTokens as collateral.
func (m UmeeMsg) HandleAddCollateral(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.AddCollateral(sdk.WrapSDKContext(ctx), m.AddCollateral)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Add Collateral", err)}
	}

	return nil
}

// HandleRemoveCollateral handles the disable amount of an selected uTokens
// as collateral.
func (m UmeeMsg) HandleRemoveCollateral(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.RemoveCollateral(sdk.WrapSDKContext(ctx), m.RemoveCollateral)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Remove Collateral", err)}
	}

	return nil
}

// HandleBorrowAsset handles the borrowing coins from the capital facility.
func (m UmeeMsg) HandleBorrowAsset(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.BorrowAsset(sdk.WrapSDKContext(ctx), m.BorrowAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Borrow Asset", err)}
	}

	return nil
}

// HandleRepayAsset handles repaying borrowed coins to the capital facility.
func (m UmeeMsg) HandleRepayAsset(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.RepayAsset(sdk.WrapSDKContext(ctx), m.RepayAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Repay Asset", err)}
	}

	return nil
}

// HandleLiquidate handles the repaying a different user's borrowed coins
// to the capital facility in exchange for some of their collateral.
func (m UmeeMsg) HandleLiquidate(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.Liquidate(sdk.WrapSDKContext(ctx), m.Liquidate)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Liquidate", err)}
	}

	return nil
}
