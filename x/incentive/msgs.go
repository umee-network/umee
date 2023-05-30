package incentive

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/checkers"
)

var (
	_ sdk.Msg = &MsgClaim{}
	_ sdk.Msg = &MsgBond{}
	_ sdk.Msg = &MsgBeginUnbonding{}
	_ sdk.Msg = &MsgEmergencyUnbond{}
	_ sdk.Msg = &MsgSponsor{}
	_ sdk.Msg = &MsgGovSetParams{}
	_ sdk.Msg = &MsgGovCreatePrograms{}
)

func NewMsgClaim(account sdk.AccAddress) *MsgClaim {
	return &MsgClaim{
		Account: account.String(),
	}
}

func (msg MsgClaim) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Account)
	return err
}

func (msg MsgClaim) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Account)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg MsgClaim) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Route implements the sdk.Msg interface.
func (msg MsgClaim) Route() string { return "" }

// Type implements the LegacyMsg interface.
func (msg MsgClaim) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgBond(account sdk.AccAddress, asset sdk.Coin) *MsgBond {
	return &MsgBond{
		Account: account.String(),
		UToken:  asset,
	}
}

func (msg MsgBond) ValidateBasic() error {
	return validateSenderAsset(msg.Account, &msg.UToken)
}

func (msg MsgBond) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Account)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg MsgBond) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Route implements the LegacyMsg interface.
func (msg MsgBond) Route() string { return "" }

// Type implements the sdk.Msg interface.
func (msg MsgBond) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgBeginUnbonding(account sdk.AccAddress, asset sdk.Coin) *MsgBeginUnbonding {
	return &MsgBeginUnbonding{
		Account: account.String(),
		UToken:  asset,
	}
}

func (msg MsgBeginUnbonding) ValidateBasic() error {
	return validateSenderAsset(msg.Account, &msg.UToken)
}

func (msg MsgBeginUnbonding) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Account)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg MsgBeginUnbonding) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Route implements the LegacyMsg interface.
func (msg MsgBeginUnbonding) Route() string { return "" }

// Type implements the sdk.Msg interface.
func (msg MsgBeginUnbonding) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgEmergencyUnbond(account sdk.AccAddress, asset sdk.Coin) *MsgEmergencyUnbond {
	return &MsgEmergencyUnbond{
		Account: account.String(),
		UToken:  asset,
	}
}

func (msg MsgEmergencyUnbond) ValidateBasic() error {
	return validateSenderAsset(msg.Account, &msg.UToken)
}

func (msg MsgEmergencyUnbond) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Account)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg MsgEmergencyUnbond) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Route implements the LegacyMsg interface.
func (msg MsgEmergencyUnbond) Route() string { return "" }

// Type implements the sdk.Msg interface.
func (msg MsgEmergencyUnbond) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgSponsor(sponsor sdk.AccAddress, programID uint32) *MsgSponsor {
	return &MsgSponsor{
		Sponsor: sponsor.String(),
		Program: programID,
	}
}

func (msg MsgSponsor) ValidateBasic() error {
	if msg.Program == 0 {
		return ErrInvalidProgramID.Wrapf("%d", msg.Program)
	}
	_, err := sdk.AccAddressFromBech32(msg.Sponsor)
	return err
}

func (msg MsgSponsor) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Sponsor)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg MsgSponsor) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Route implements the LegacyMsg interface.
func (msg MsgSponsor) Route() string { return "" }

// Type implements the sdk.Msg interface.
func (msg MsgSponsor) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgGovSetParams(authority string, params Params) *MsgGovSetParams {
	return &MsgGovSetParams{
		Params:    params,
		Authority: authority,
	}
}

func (msg MsgGovSetParams) ValidateBasic() error {
	if err := checkers.IsGovAuthority(msg.Authority); err != nil {
		return err
	}
	return msg.Params.Validate()
}

// GetSignBytes implements Msg
func (msg MsgGovSetParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgGovSetParams) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// Route implements the LegacyMsg interface.
func (msg MsgGovSetParams) Route() string { return "" }

// Type implements the sdk.Msg interface.
func (msg MsgGovSetParams) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgGovCreatePrograms(authority string, programs []IncentiveProgram) *MsgGovCreatePrograms {
	return &MsgGovCreatePrograms{
		Authority: authority,
		Programs:  programs,
	}
}

func (msg MsgGovCreatePrograms) ValidateBasic() error {
	if err := checkers.IsGovAuthority(msg.Authority); err != nil {
		return err
	}
	if len(msg.Programs) == 0 {
		return ErrEmptyProposal
	}
	for _, p := range msg.Programs {
		if err := p.ValidateProposed(); err != nil {
			return err
		}
	}
	return nil
}

// GetSignBytes implements Msg
func (msg MsgGovCreatePrograms) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgGovCreatePrograms) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// Route implements the LegacyMsg interface.
func (msg MsgGovCreatePrograms) Route() string { return "" }

// Type implements the sdk.Msg interface.
func (msg MsgGovCreatePrograms) Type() string { return sdk.MsgTypeURL(&msg) }

func validateSenderAsset(sender string, asset *sdk.Coin) error {
	_, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrNilAsset
	}
	return asset.Validate()
}
