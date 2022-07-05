package message

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	lvtypes "github.com/umee-network/umee/v2/x/leverage/types"
)

// HandleLendAsset handles the LendAsset value of an address.
func (umeeMsg UmeeMsg) HandleLendAsset(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.LendAsset(sdk.WrapSDKContext(ctx), umeeMsg.LendAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Lend Asset", err)}
	}

	return nil
}

// HandleWithdrawAsset handles the WithdrawAsset value of an address.
func (umeeMsg UmeeMsg) HandleWithdrawAsset(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.WithdrawAsset(sdk.WrapSDKContext(ctx), umeeMsg.WithdrawAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Withdraw Asset", err)}
	}

	return nil
}

// HandleAddCollateral handles the enable selected uTokens as collateral.
func (umeeMsg UmeeMsg) HandleAddCollateral(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.AddCollateral(sdk.WrapSDKContext(ctx), umeeMsg.AddCollateral)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Add Collateral", err)}
	}

	return nil
}

// HandleRemoveCollateral handles the disable amount of an selected uTokens
// as collateral.
func (umeeMsg UmeeMsg) HandleRemoveCollateral(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.RemoveCollateral(sdk.WrapSDKContext(ctx), umeeMsg.RemoveCollateral)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Remove Collateral", err)}
	}

	return nil
}

// HandleBorrowAsset handles the borrowing coins from the capital facility.
func (umeeMsg UmeeMsg) HandleBorrowAsset(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.BorrowAsset(sdk.WrapSDKContext(ctx), umeeMsg.BorrowAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Borrow Asset", err)}
	}

	return nil
}

// HandleRepayAsset handles repaying borrowed coins to the capital facility.
func (umeeMsg UmeeMsg) HandleRepayAsset(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.RepayAsset(sdk.WrapSDKContext(ctx), umeeMsg.RepayAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Repay Asset", err)}
	}

	return nil
}

// HandleLiquidate handles the repaying a different user's borrowed coins
// to the capital facility in exchange for some of their collateral.
func (umeeMsg UmeeMsg) HandleLiquidate(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.Liquidate(sdk.WrapSDKContext(ctx), umeeMsg.Liquidate)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Liquidate", err)}
	}

	return nil
}
