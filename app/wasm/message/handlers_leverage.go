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

// HandleWithdraw handles the Withdraw value of an address.
func (m UmeeMsg) HandleWithdraw(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.Withdraw(sdk.WrapSDKContext(ctx), m.Withdraw)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Withdraw", err)}
	}

	return nil
}

// HandleCollateralize handles the enable selected uTokens as collateral.
func (m UmeeMsg) HandleCollateralize(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.Collateralize(sdk.WrapSDKContext(ctx), m.Collateralize)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Collateralize", err)}
	}

	return nil
}

// HandleDecollateralize handles the disable amount of an selected uTokens
// as collateral.
func (m UmeeMsg) HandleDecollateralize(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.Decollateralize(sdk.WrapSDKContext(ctx), m.Decollateralize)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Decollateralize", err)}
	}

	return nil
}

// HandleBorrow handles the borrowing coins from the capital facility.
func (m UmeeMsg) HandleBorrow(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.Borrow(sdk.WrapSDKContext(ctx), m.Borrow)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Borrow", err)}
	}

	return nil
}

// HandleRepay handles repaying borrowed coins to the capital facility.
func (m UmeeMsg) HandleRepay(
	ctx sdk.Context,
	s lvtypes.MsgServer,
) error {
	_, err := s.Repay(sdk.WrapSDKContext(ctx), m.Repay)
	if err != nil {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned msg Repay", err)}
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
