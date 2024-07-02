package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/checkers"
)

func NewMsgSupply(supplier sdk.AccAddress, asset sdk.Coin) *MsgSupply {
	return &MsgSupply{
		Supplier: supplier.String(),
		Asset:    asset,
	}
}

func (msg *MsgSupply) ValidateBasic() error {
	return validateSenderAndAsset(msg.Supplier, &msg.Asset)
}

func (msg *MsgSupply) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// LegacyMsg.Type implementations

func (msg MsgSupply) Route() string { return "" }
func (msg MsgSupply) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgSupply) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgWithdraw(supplier sdk.AccAddress, asset sdk.Coin) *MsgWithdraw {
	return &MsgWithdraw{
		Supplier: supplier.String(),
		Asset:    asset,
	}
}

func (msg *MsgWithdraw) ValidateBasic() error {
	return validateSenderAndAsset(msg.Supplier, &msg.Asset)
}

func (msg *MsgWithdraw) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// LegacyMsg.Type implementations

func (msg MsgWithdraw) Route() string { return "" }
func (msg MsgWithdraw) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgWithdraw) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgMaxWithdraw(supplier sdk.AccAddress, denom string) *MsgMaxWithdraw {
	return &MsgMaxWithdraw{
		Supplier: supplier.String(),
		Denom:    denom,
	}
}

func (msg *MsgMaxWithdraw) ValidateBasic() error {
	return validateSenderAndDenom(msg.Supplier, msg.Denom)
}

func (msg *MsgMaxWithdraw) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// LegacyMsg.Type implementations

func (msg MsgMaxWithdraw) Route() string { return "" }
func (msg MsgMaxWithdraw) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgMaxWithdraw) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgCollateralize(borrower sdk.AccAddress, asset sdk.Coin) *MsgCollateralize {
	return &MsgCollateralize{
		Borrower: borrower.String(),
		Asset:    asset,
	}
}

func (msg *MsgCollateralize) ValidateBasic() error {
	return validateSenderAndAsset(msg.Borrower, &msg.Asset)
}

func (msg *MsgCollateralize) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// LegacyMsg.Type implementations

func (msg MsgCollateralize) Route() string { return "" }
func (msg MsgCollateralize) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgCollateralize) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgSupplyCollateral(supplier sdk.AccAddress, asset sdk.Coin) *MsgSupplyCollateral {
	return &MsgSupplyCollateral{
		Supplier: supplier.String(),
		Asset:    asset,
	}
}

func (msg *MsgSupplyCollateral) ValidateBasic() error {
	return validateSenderAndAsset(msg.Supplier, &msg.Asset)
}

func (msg *MsgSupplyCollateral) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// LegacyMsg.Type implementations

func (msg MsgSupplyCollateral) Route() string { return "" }
func (msg MsgSupplyCollateral) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgSupplyCollateral) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgDecollateralize(borrower sdk.AccAddress, asset sdk.Coin) *MsgDecollateralize {
	return &MsgDecollateralize{
		Borrower: borrower.String(),
		Asset:    asset,
	}
}

func (msg *MsgDecollateralize) ValidateBasic() error {
	return validateSenderAndAsset(msg.Borrower, &msg.Asset)
}

func (msg *MsgDecollateralize) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// LegacyMsg.Type implementations
func (msg MsgDecollateralize) Route() string { return "" }
func (msg MsgDecollateralize) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgDecollateralize) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgBorrow(borrower sdk.AccAddress, asset sdk.Coin) *MsgBorrow {
	return &MsgBorrow{
		Borrower: borrower.String(),
		Asset:    asset,
	}
}

func (msg *MsgBorrow) ValidateBasic() error {
	return validateSenderAndAsset(msg.Borrower, &msg.Asset)
}

func (msg *MsgBorrow) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// LegacyMsg.Type implementations
func (msg MsgBorrow) Route() string { return "" }
func (msg MsgBorrow) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgBorrow) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgMaxBorrow(borrower sdk.AccAddress, denom string) *MsgMaxBorrow {
	return &MsgMaxBorrow{
		Borrower: borrower.String(),
		Denom:    denom,
	}
}

func (msg *MsgMaxBorrow) ValidateBasic() error {
	return validateSenderAndDenom(msg.Borrower, msg.Denom)
}

func (msg *MsgMaxBorrow) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// LegacyMsg.Type implementations
func (msg MsgMaxBorrow) Route() string { return "" }
func (msg MsgMaxBorrow) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgMaxBorrow) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgRepay(borrower sdk.AccAddress, asset sdk.Coin) *MsgRepay {
	return &MsgRepay{
		Borrower: borrower.String(),
		Asset:    asset,
	}
}

func (msg *MsgRepay) ValidateBasic() error {
	return validateSenderAndAsset(msg.Borrower, &msg.Asset)
}

func (msg *MsgRepay) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Borrower)
}

// LegacyMsg.Type implementations
func (msg MsgRepay) Route() string { return "" }
func (msg MsgRepay) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgRepay) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgLiquidate(liquidator, borrower sdk.AccAddress, repayment sdk.Coin, rewardDenom string) *MsgLiquidate {
	return &MsgLiquidate{
		Liquidator:  liquidator.String(),
		Borrower:    borrower.String(),
		Repayment:   repayment,
		RewardDenom: rewardDenom,
	}
}

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

// LegacyMsg.Type implementations
func (msg MsgLiquidate) Route() string { return "" }
func (msg MsgLiquidate) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgLiquidate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func NewMsgLeveragedLiquidate(
	liquidator, borrower sdk.AccAddress,
	repayDenom, rewardDenom string,
	maxRepay sdkmath.LegacyDec,
) *MsgLeveragedLiquidate {
	return &MsgLeveragedLiquidate{
		Liquidator:  liquidator.String(),
		Borrower:    borrower.String(),
		RepayDenom:  repayDenom,
		RewardDenom: rewardDenom,
		MaxRepay:    maxRepay,
	}
}

func (msg *MsgLeveragedLiquidate) ValidateBasic() error {
	if msg.RepayDenom != "" {
		if err := sdk.ValidateDenom(msg.RepayDenom); err != nil {
			return err
		}
	}
	if msg.RewardDenom != "" {
		if err := sdk.ValidateDenom(msg.RewardDenom); err != nil {
			return err
		}
	}
	if !msg.MaxRepay.IsZero() && msg.MaxRepay.LT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("nonzero max repay %s is less than one", msg.MaxRepay)
	}
	_, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return err
	}
	_, err = sdk.AccAddressFromBech32(msg.Liquidator)
	return err
}

func (msg *MsgLeveragedLiquidate) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Liquidator)
}

// LegacyMsg.Type implementations
func (msg MsgLeveragedLiquidate) Route() string { return "" }
func (msg MsgLeveragedLiquidate) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgLeveragedLiquidate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// -- helper methods -- //

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
