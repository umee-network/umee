package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// This file adds the required GetSigners and ValidateBasic methods to the messages defined in tx.pb.go

func (msg *MsgLendAsset) ValidateBasic() error {
	lender, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		return err
	}
	if lender.Empty() {
		return errors.New("empty address")
	}
	asset := msg.GetAmount()
	if !asset.IsValid() {
		// Denom did not match ^[a-z][a-z0-9/]{2,63}$
		// or amount was negative
		return errors.New("invalid asset")
	}
	return nil
}

func (msg *MsgWithdrawAsset) ValidateBasic() error {
	lender, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		return err
	}
	if lender.Empty() {
		return errors.New("empty address")
	}
	asset := msg.GetAmount()
	if !asset.IsValid() {
		// Denom did not match ^[a-z][a-z0-9/]{2,63}$
		// or amount was negative
		return errors.New("invalid asset")
	}
	return nil
}

func (msg *MsgLendAsset) GetSigners() []sdk.AccAddress {
	lender, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		// Question: There is no error return, but is there a way to not use a panic here?
		panic(err)
	}
	return []sdk.AccAddress{lender}
}

func (msg *MsgWithdrawAsset) GetSigners() []sdk.AccAddress {
	lender, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		// Question: There is no error return, but is there a way to not use a panic here?
		panic(err)
	}
	return []sdk.AccAddress{lender}
}
