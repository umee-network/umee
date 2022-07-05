package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewMsgSupply(supplier sdk.AccAddress, amount sdk.Coin) *MsgSupply {
	return &MsgSupply{
		Supplier: supplier.String(),
		Amount:   amount,
	}
}

func (msg MsgSupply) Route() string { return ModuleName }
func (msg MsgSupply) Type() string  { return EventTypeLoanAsset }

func (msg *MsgSupply) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetSupplier())
	if err != nil {
		return err
	}

	if asset := msg.GetAmount(); !asset.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, asset.String())
	}

	return nil
}

func (msg *MsgSupply) GetSigners() []sdk.AccAddress {
	supplier, _ := sdk.AccAddressFromBech32(msg.GetSupplier())
	return []sdk.AccAddress{supplier}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgSupply) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgWithdrawAsset(supplier sdk.AccAddress, amount sdk.Coin) *MsgWithdrawAsset {
	return &MsgWithdrawAsset{
		Supplier: supplier.String(),
		Amount:   amount,
	}
}

func (msg MsgWithdrawAsset) Route() string { return ModuleName }
func (msg MsgWithdrawAsset) Type() string  { return EventTypeWithdrawLoanedAsset }

func (msg *MsgWithdrawAsset) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetSupplier())
	if err != nil {
		return err
	}

	if asset := msg.GetAmount(); !asset.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, asset.String())
	}

	return nil
}

func (msg *MsgWithdrawAsset) GetSigners() []sdk.AccAddress {
	supplier, _ := sdk.AccAddressFromBech32(msg.GetSupplier())
	return []sdk.AccAddress{supplier}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgWithdrawAsset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgAddCollateral(borrower sdk.AccAddress, coin sdk.Coin) *MsgAddCollateral {
	return &MsgAddCollateral{
		Borrower: borrower.String(),
		Coin:     coin,
	}
}

func (msg MsgAddCollateral) Route() string { return ModuleName }
func (msg MsgAddCollateral) Type() string  { return EventTypeAddCollateral }

func (msg *MsgAddCollateral) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetBorrower())
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgAddCollateral) GetSigners() []sdk.AccAddress {
	borrower, _ := sdk.AccAddressFromBech32(msg.GetBorrower())
	return []sdk.AccAddress{borrower}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgAddCollateral) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgRemoveCollateral(borrower sdk.AccAddress, coin sdk.Coin) *MsgRemoveCollateral {
	return &MsgRemoveCollateral{
		Borrower: borrower.String(),
		Coin:     coin,
	}
}

func (msg MsgRemoveCollateral) Route() string { return ModuleName }
func (msg MsgRemoveCollateral) Type() string  { return EventTypeRemoveCollateral }

func (msg *MsgRemoveCollateral) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetBorrower())
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgRemoveCollateral) GetSigners() []sdk.AccAddress {
	borrower, _ := sdk.AccAddressFromBech32(msg.GetBorrower())
	return []sdk.AccAddress{borrower}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgRemoveCollateral) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgBorrowAsset(borrower sdk.AccAddress, amount sdk.Coin) *MsgBorrowAsset {
	return &MsgBorrowAsset{
		Borrower: borrower.String(),
		Amount:   amount,
	}
}

func (msg MsgBorrowAsset) Route() string { return ModuleName }
func (msg MsgBorrowAsset) Type() string  { return EventTypeBorrowAsset }

func (msg *MsgBorrowAsset) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetBorrower())
	if err != nil {
		return err
	}

	if asset := msg.GetAmount(); !asset.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, asset.String())
	}

	return nil
}

func (msg *MsgBorrowAsset) GetSigners() []sdk.AccAddress {
	borrower, _ := sdk.AccAddressFromBech32(msg.GetBorrower())
	return []sdk.AccAddress{borrower}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgBorrowAsset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgRepayAsset(borrower sdk.AccAddress, amount sdk.Coin) *MsgRepayAsset {
	return &MsgRepayAsset{
		Borrower: borrower.String(),
		Amount:   amount,
	}
}

func (msg MsgRepayAsset) Route() string { return ModuleName }
func (msg MsgRepayAsset) Type() string  { return EventTypeRepayBorrowedAsset }

func (msg *MsgRepayAsset) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetBorrower())
	if err != nil {
		return err
	}

	if asset := msg.GetAmount(); !asset.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, asset.String())
	}

	return nil
}

func (msg *MsgRepayAsset) GetSigners() []sdk.AccAddress {
	borrower, _ := sdk.AccAddressFromBech32(msg.GetBorrower())
	return []sdk.AccAddress{borrower}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgRepayAsset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgLiquidate(liquidator, borrower sdk.AccAddress, repayment, reward sdk.Coin) *MsgLiquidate {
	return &MsgLiquidate{
		Liquidator: liquidator.String(),
		Borrower:   borrower.String(),
		Repayment:  repayment,
		Reward:     reward,
	}
}

func (msg MsgLiquidate) Route() string { return ModuleName }
func (msg MsgLiquidate) Type() string  { return EventTypeLiquidate }

func (msg *MsgLiquidate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetLiquidator())
	if err != nil {
		return err
	}
	_, err = sdk.AccAddressFromBech32(msg.GetBorrower())
	if err != nil {
		return err
	}

	if asset := msg.GetRepayment(); !asset.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, asset.String())
	}

	if asset := msg.GetReward(); !asset.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, asset.String())
	}

	return nil
}

func (msg *MsgLiquidate) GetSigners() []sdk.AccAddress {
	liquidator, _ := sdk.AccAddressFromBech32(msg.GetLiquidator())
	return []sdk.AccAddress{liquidator}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgLiquidate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}
