package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewMsgClaim(addr sdk.AccAddress) *MsgClaim {
	return &MsgClaim{
		Lender: addr.String(),
	}
}

func (msg MsgClaim) Route() string { return ModuleName }
func (msg MsgClaim) Type() string  { return EventTypeClaimReward }

func (msg *MsgClaim) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		return err
	}

	return nil
}

func (msg *MsgClaim) GetSigners() []sdk.AccAddress {
	lender, _ := sdk.AccAddressFromBech32(msg.GetLender())
	return []sdk.AccAddress{lender}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgClaim) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgLock(addr sdk.AccAddress, tier uint32, amount sdk.Coin) *MsgLock {
	return &MsgLock{
		Lender: addr.String(),
		Tier:   tier,
		Amount: amount,
	}
}

func (msg MsgLock) Route() string { return ModuleName }
func (msg MsgLock) Type() string  { return EventTypeLockCollateral }

func (msg *MsgLock) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		return err
	}

	if msg.Tier < 1 || msg.Tier > 3 {
		return ErrInvalidTier
	}

	if err := msg.Amount.Validate(); err != nil {
		return err
	}

	return nil
}

func (msg *MsgLock) GetSigners() []sdk.AccAddress {
	lender, _ := sdk.AccAddressFromBech32(msg.GetLender())
	return []sdk.AccAddress{lender}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgLock) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgUnlock(addr sdk.AccAddress, tier uint32, amount sdk.Coin) *MsgUnlock {
	return &MsgUnlock{
		Lender: addr.String(),
		Tier:   tier,
		Amount: amount,
	}
}

func (msg MsgUnlock) Route() string { return ModuleName }
func (msg MsgUnlock) Type() string  { return EventTypeUnlockCollateral }

func (msg *MsgUnlock) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		return err
	}

	if msg.Tier < 1 || msg.Tier > 3 {
		return ErrInvalidTier
	}

	if err := msg.Amount.Validate(); err != nil {
		return err
	}

	return nil
}

func (msg *MsgUnlock) GetSigners() []sdk.AccAddress {
	lender, _ := sdk.AccAddressFromBech32(msg.GetLender())
	return []sdk.AccAddress{lender}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgUnlock) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgSponsor(addr sdk.AccAddress, id uint32, amount sdk.Coin) *MsgSponsor {
	return &MsgSponsor{
		Sponsor:   addr.String(),
		ProgramId: id,
		Amount:    amount,
	}
}

func (msg MsgSponsor) Route() string { return ModuleName }
func (msg MsgSponsor) Type() string  { return EventTypeSponsorProgram }

func (msg *MsgSponsor) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetSponsor())
	if err != nil {
		return err
	}

	if msg.ProgramId < 1 {
		return ErrInvalidProgramID
	}

	if err := msg.Amount.Validate(); err != nil {
		return err
	}

	return nil
}

func (msg *MsgSponsor) GetSigners() []sdk.AccAddress {
	lender, _ := sdk.AccAddressFromBech32(msg.GetSponsor())
	return []sdk.AccAddress{lender}
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgSponsor) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}
