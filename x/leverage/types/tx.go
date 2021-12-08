package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewMsgLendAsset(lender sdk.AccAddress, amount sdk.Coin) *MsgLendAsset {
	return &MsgLendAsset{
		Lender: lender.String(),
		Amount: amount,
	}
}

func (msg *MsgLendAsset) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		return err
	}

	if asset := msg.GetAmount(); !asset.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, asset.String())
	}

	return nil
}

func (msg *MsgLendAsset) GetSigners() []sdk.AccAddress {
	lender, _ := sdk.AccAddressFromBech32(msg.GetLender())
	return []sdk.AccAddress{lender}
}

func NewMsgWithdrawAsset(lender sdk.AccAddress, amount sdk.Coin) *MsgWithdrawAsset {
	return &MsgWithdrawAsset{
		Lender: lender.String(),
		Amount: amount,
	}
}

func (msg *MsgWithdrawAsset) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		return err
	}

	if asset := msg.GetAmount(); !asset.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, asset.String())
	}

	return nil
}

func (msg *MsgWithdrawAsset) GetSigners() []sdk.AccAddress {
	lender, _ := sdk.AccAddressFromBech32(msg.GetLender())
	return []sdk.AccAddress{lender}
}

func NewMsgSetCollateral(borrower sdk.AccAddress, denom string, enable bool) *MsgSetCollateral {
	return &MsgSetCollateral{
		Borrower: borrower.String(),
		Denom:    denom,
		Enable:   enable,
	}
}

func (msg *MsgSetCollateral) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetBorrower())
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgSetCollateral) GetSigners() []sdk.AccAddress {
	borrower, _ := sdk.AccAddressFromBech32(msg.GetBorrower())
	return []sdk.AccAddress{borrower}
}

func NewMsgBorrowAsset(borrower sdk.AccAddress, amount sdk.Coin) *MsgBorrowAsset {
	return &MsgBorrowAsset{
		Borrower: borrower.String(),
		Amount:   amount,
	}
}

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

func NewMsgRepayAsset(borrower sdk.AccAddress, amount sdk.Coin) *MsgRepayAsset {
	return &MsgRepayAsset{
		Borrower: borrower.String(),
		Amount:   amount,
	}
}

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

func NewMsgLiquidate(liquidator, borrower sdk.AccAddress, repayment sdk.Coin, rewardDenom string) *MsgLiquidate {
	return &MsgLiquidate{
		Liquidator:  borrower.String(),
		Borrower:    borrower.String(),
		Repayment:   repayment,
		RewardDenom: rewardDenom,
	}
}

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

	if msg.GetRewardDenom() == "" {
		return errors.New("empty reward denom")
	}

	return nil
}

func (msg *MsgLiquidate) GetSigners() []sdk.AccAddress {
	liquidator, _ := sdk.AccAddressFromBech32(msg.GetLiquidator())
	return []sdk.AccAddress{liquidator}
}
