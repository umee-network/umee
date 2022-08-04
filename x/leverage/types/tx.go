package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v2/util/checkers"
)

func NewMsgSupply(supplier sdk.AccAddress, asset sdk.Coin) *MsgSupply {
	return &MsgSupply{
		Supplier: supplier.String(),
		Asset:    asset,
	}
}

func (msg MsgSupply) Route() string { return ModuleName }
func (msg MsgSupply) Type() string  { return EventTypeSupply }

func (msg *MsgSupply) ValidateBasic() error {
	return validateSenderAndAsset(msg.Supplier, &msg.Asset)
}

func (msg *MsgSupply) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgSupply) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgWithdraw(supplier sdk.AccAddress, asset sdk.Coin) *MsgWithdraw {
	return &MsgWithdraw{
		Supplier: supplier.String(),
		Asset:    asset,
	}
}

func (msg MsgWithdraw) Route() string { return ModuleName }
func (msg MsgWithdraw) Type() string  { return EventTypeWithdraw }

func (msg *MsgWithdraw) ValidateBasic() error {
	return validateSenderAndAsset(msg.Supplier, &msg.Asset)
}

func (msg *MsgWithdraw) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgCollateralize(borrower sdk.AccAddress, coin sdk.Coin) *MsgCollateralize {
	return &MsgCollateralize{
		Borrower: borrower.String(),
		Coin:     coin,
	}
}

func (msg MsgCollateralize) Route() string { return ModuleName }
func (msg MsgCollateralize) Type() string  { return EventTypeCollateralize }

func (msg *MsgCollateralize) ValidateBasic() error {
	return validateSenderAndAsset(msg.Borrower, nil)
}

func (msg *MsgCollateralize) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgCollateralize) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgDecollateralize(borrower sdk.AccAddress, coin sdk.Coin) *MsgDecollateralize {
	return &MsgDecollateralize{
		Borrower: borrower.String(),
		Coin:     coin,
	}
}

func (msg MsgDecollateralize) Route() string { return ModuleName }
func (msg MsgDecollateralize) Type() string  { return EventTypeDecollateralize }

func (msg *MsgDecollateralize) ValidateBasic() error {
	return validateSenderAndAsset(msg.Borrower, nil)
}

func (msg *MsgDecollateralize) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgDecollateralize) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgBorrow(borrower sdk.AccAddress, asset sdk.Coin) *MsgBorrow {
	return &MsgBorrow{
		Borrower: borrower.String(),
		Asset:    asset,
	}
}

func (msg MsgBorrow) Route() string { return ModuleName }
func (msg MsgBorrow) Type() string  { return EventTypeBorrow }

func (msg *MsgBorrow) ValidateBasic() error {
	return validateSenderAndAsset(msg.Borrower, &msg.Asset)
}

func (msg *MsgBorrow) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgBorrow) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgRepay(borrower sdk.AccAddress, asset sdk.Coin) *MsgRepay {
	return &MsgRepay{
		Borrower: borrower.String(),
		Asset:    asset,
	}
}

func (msg MsgRepay) Route() string { return ModuleName }
func (msg MsgRepay) Type() string  { return EventTypeRepayBorrowedAsset }

func (msg *MsgRepay) ValidateBasic() error {
	return validateSenderAndAsset(msg.Borrower, &msg.Asset)
}

func (msg *MsgRepay) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgRepay) GetSignBytes() []byte {
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
	if err := validateSenderAndAsset(msg.Borrower, &msg.Repayment); err != nil {
		return err
	}
	return validateSenderAndAsset(msg.Liquidator, &msg.Reward)
}

func (msg *MsgLiquidate) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Liquidator)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgLiquidate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func validateSenderAndAsset(sender string, asset *sdk.Coin) error {
	_, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}
	if asset != nil && !asset.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, asset.String())
	}
	return nil
}
