package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v2/util/checkers"
)

func NewMsgSupply(supplier sdk.AccAddress, coin sdk.Coin) *MsgSupply {
	return &MsgSupply{
		Supplier: supplier.String(),
		Coin:     coin,
	}
}

func (msg MsgSupply) Route() string { return ModuleName }
func (msg MsgSupply) Type() string  { return EventTypeSupply }

func (msg *MsgSupply) ValidateBasic() error {
	return validateSenderAndCoin(msg.Supplier, &msg.Coin)
}

func (msg *MsgSupply) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgSupply) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgWithdraw(supplier sdk.AccAddress, coin sdk.Coin) *MsgWithdraw {
	return &MsgWithdraw{
		Supplier: supplier.String(),
		Coin:     coin,
	}
}

func (msg MsgWithdraw) Route() string { return ModuleName }
func (msg MsgWithdraw) Type() string  { return EventTypeWithdraw }

func (msg *MsgWithdraw) ValidateBasic() error {
	return validateSenderAndCoin(msg.Supplier, &msg.Coin)
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
	return validateSenderAndCoin(msg.Borrower, nil)
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
	return validateSenderAndCoin(msg.Borrower, nil)
}

func (msg *MsgDecollateralize) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgDecollateralize) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgBorrow(borrower sdk.AccAddress, coin sdk.Coin) *MsgBorrow {
	return &MsgBorrow{
		Borrower: borrower.String(),
		Coin:     coin,
	}
}

func (msg MsgBorrow) Route() string { return ModuleName }
func (msg MsgBorrow) Type() string  { return EventTypeBorrow }

func (msg *MsgBorrow) ValidateBasic() error {
	return validateSenderAndCoin(msg.Borrower, &msg.Coin)
}

func (msg *MsgBorrow) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgBorrow) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgRepay(borrower sdk.AccAddress, coin sdk.Coin) *MsgRepay {
	return &MsgRepay{
		Borrower: borrower.String(),
		Coin:     coin,
	}
}

func (msg MsgRepay) Route() string { return ModuleName }
func (msg MsgRepay) Type() string  { return EventTypeRepay }

func (msg *MsgRepay) ValidateBasic() error {
	return validateSenderAndCoin(msg.Borrower, &msg.Coin)
}

func (msg *MsgRepay) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgRepay) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgLiquidate(liquidator, borrower sdk.AccAddress, repayment sdk.Coin, rewardDenom string) *MsgLiquidate {
	return &MsgLiquidate{
		Liquidator:  liquidator.String(),
		Borrower:    borrower.String(),
		Repayment:   repayment,
		RewardDenom: rewardDenom,
	}
}

func (msg MsgLiquidate) Route() string { return ModuleName }
func (msg MsgLiquidate) Type() string  { return EventTypeLiquidate }

func (msg *MsgLiquidate) ValidateBasic() error {
	if err := validateSenderAndCoin(msg.Borrower, &msg.Repayment); err != nil {
		return err
	}
	if err := sdk.ValidateDenom(msg.RewardDenom); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.Liquidator)
	return err
}

func (msg *MsgLiquidate) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Liquidator)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgLiquidate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func validateSenderAndCoin(sender string, coin *sdk.Coin) error {
	_, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}
	if coin != nil && !coin.IsValid() {
		return sdkerrors.Wrap(ErrInvalidAsset, coin.String())
	}
	return nil
}
