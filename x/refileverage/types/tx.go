package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/checkers"
)

func NewMsgMaxWithdraw(supplier sdk.AccAddress, denom string) *MsgMaxWithdraw {
	return &MsgMaxWithdraw{
		Supplier: supplier.String(),
		Denom:    denom,
	}
}

func (msg MsgMaxWithdraw) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgMaxWithdraw) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg *MsgMaxWithdraw) ValidateBasic() error {
	return validateSenderAndDenom(msg.Supplier, msg.Denom)
}

func (msg *MsgMaxWithdraw) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgMaxWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgSupplyCollateral(supplier sdk.AccAddress, asset sdk.Coin) *MsgSupplyCollateral {
	return &MsgSupplyCollateral{
		Supplier: supplier.String(),
		Asset:    asset,
	}
}

func (msg MsgSupplyCollateral) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgSupplyCollateral) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg *MsgSupplyCollateral) ValidateBasic() error {
	return validateSenderAndAsset(msg.Supplier, &msg.Asset)
}

func (msg *MsgSupplyCollateral) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgSupplyCollateral) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgDecollateralize(borrower sdk.AccAddress, asset sdk.Coin) *MsgDecollateralize {
	return &MsgDecollateralize{
		Borrower: borrower.String(),
		Asset:    asset,
	}
}

func (msg MsgDecollateralize) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgDecollateralize) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg *MsgDecollateralize) ValidateBasic() error {
	return validateSenderAndAsset(msg.Borrower, &msg.Asset)
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

func (msg MsgBorrow) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgBorrow) Type() string  { return sdk.MsgTypeURL(&msg) }

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

func NewMsgMaxBorrow(borrower sdk.AccAddress, denom string) *MsgMaxBorrow {
	return &MsgMaxBorrow{
		Borrower: borrower.String(),
		Denom:    denom,
	}
}

func (msg MsgMaxBorrow) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgMaxBorrow) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg *MsgMaxBorrow) ValidateBasic() error {
	return validateSenderAndDenom(msg.Borrower, msg.Denom)
}

func (msg *MsgMaxBorrow) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgMaxBorrow) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgRepay(borrower sdk.AccAddress, asset sdk.Coin) *MsgRepay {
	return &MsgRepay{
		Borrower: borrower.String(),
		Asset:    asset,
	}
}

func (msg MsgRepay) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgRepay) Type() string  { return sdk.MsgTypeURL(&msg) }

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

func NewMsgLiquidate(liquidator, borrower sdk.AccAddress, repayment sdk.Coin, rewardDenom string) *MsgLiquidate {
	return &MsgLiquidate{
		Liquidator:  liquidator.String(),
		Borrower:    borrower.String(),
		Repayment:   repayment,
		RewardDenom: rewardDenom,
	}
}

func (msg MsgLiquidate) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgLiquidate) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg *MsgLiquidate) ValidateBasic() error {
	if err := validateSenderAndAsset(msg.Borrower, &msg.Repayment); err != nil {
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

func validateSenderAndAsset(sender string, asset *sdk.Coin) error {
	_, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrNilAsset
	}
	return asset.Validate()
}

func validateSenderAndDenom(sender string, denom string) error {
	_, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}
	return sdk.ValidateDenom(denom)
}
