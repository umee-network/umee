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

// TODO: Other transaction types
