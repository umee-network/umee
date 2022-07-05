package message

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	lvtypes "github.com/umee-network/umee/v2/x/leverage/types"
)

// HandleLendAsset handles the LendAsset value of an address.
func (umeeQuery UmeeMsg) HandleLendAsset(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.LendAsset(sdk.WrapSDKContext(ctx), umeeQuery.LendAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Lend Asset", err)}
	}

	return nil
}

// HandleWithdrawAsset handles the WithdrawAsset value of an address.
func (umeeQuery UmeeMsg) HandleWithdrawAsset(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.WithdrawAsset(sdk.WrapSDKContext(ctx), umeeQuery.WithdrawAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Withdraw Asset", err)}
	}

	return nil
}

// HandleBorrowAsset handles the borrowing coins from the capital facility.
func (umeeQuery UmeeMsg) HandleBorrowAsset(
	ctx sdk.Context,
	msgServer lvtypes.MsgServer,
) error {
	_, err := msgServer.BorrowAsset(sdk.WrapSDKContext(ctx), umeeQuery.BorrowAsset)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrow Asset", err)}
	}

	return nil
}
