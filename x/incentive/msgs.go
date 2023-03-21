package incentive

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/checkers"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

var (
	_ sdk.Msg = &MsgClaim{}
	_ sdk.Msg = &MsgBeginUnbonding{}
	_ sdk.Msg = &MsgBond{}
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
func (msg MsgClaim) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgClaim) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgBond(account sdk.AccAddress, tier uint32, asset sdk.Coin) *MsgBond {
	return &MsgBond{
		Account: account.String(),
		Tier:    tier,
		Asset:   asset,
	}
}

func (msg MsgBond) ValidateBasic() error {
	return validateSenderAssetTier(msg.Account, msg.Tier, &msg.Asset)
}

func (msg MsgBond) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Account)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg MsgBond) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Route implements the sdk.Msg interface.
func (msg MsgBond) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgBond) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgBeginUnbonding(account sdk.AccAddress, tier uint32, asset sdk.Coin) *MsgBeginUnbonding {
	return &MsgBeginUnbonding{
		Account: account.String(),
		Tier:    tier,
		Asset:   asset,
	}
}

func (msg MsgBeginUnbonding) ValidateBasic() error {
	return validateSenderAssetTier(msg.Account, msg.Tier, &msg.Asset)
}

func (msg MsgBeginUnbonding) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Account)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg MsgBeginUnbonding) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Route implements the sdk.Msg interface.
func (msg MsgBeginUnbonding) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgBeginUnbonding) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgSponsor(sponsor sdk.AccAddress, programID uint32, asset sdk.Coin) *MsgSponsor {
	return &MsgSponsor{
		Sponsor: sponsor.String(),
		Program: programID,
		Asset:   asset,
	}
}

func (msg MsgSponsor) ValidateBasic() error {
	if msg.Program == 0 {
		return ErrInvalidProgramID.Wrapf("%d", msg.Program)
	}
	return validateSenderAsset(msg.Sponsor, &msg.Asset)
}

func (msg MsgSponsor) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Sponsor)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg MsgSponsor) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Route implements the sdk.Msg interface.
func (msg MsgSponsor) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgSponsor) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgGovSetParams(authority, title, description string, params Params) *MsgGovSetParams {
	return &MsgGovSetParams{
		Title:       title,
		Description: description,
		Params:      params,
		Authority:   authority,
	}
}

func (msg MsgGovSetParams) ValidateBasic() error {
	if err := checkers.ValidateProposal(msg.Title, msg.Description, msg.Authority); err != nil {
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

// Route implements the sdk.Msg interface.
func (msg MsgGovSetParams) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgGovSetParams) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgGovCreatePrograms(authority, title, description string, programs []IncentiveProgram) *MsgGovCreatePrograms {
	return &MsgGovCreatePrograms{
		Title:       title,
		Description: description,
		Programs:    programs,
		Authority:   authority,
	}
}

func (msg MsgGovCreatePrograms) ValidateBasic() error {
	if err := checkers.ValidateProposal(msg.Title, msg.Description, msg.Authority); err != nil {
		return err
	}
	if len(msg.Programs) == 0 {
		return ErrEmptyProposal
	}
	for _, p := range msg.Programs {
		if err := validateProposedIncentiveProgram(p); err != nil {
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

// Route implements the sdk.Msg interface.
func (msg MsgGovCreatePrograms) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgGovCreatePrograms) Type() string { return sdk.MsgTypeURL(&msg) }

// validateProposedIncentiveProgram runs IncentiveProgram.Validate and also checks additional requirements applying
// to incentive programs which have not yet been funded or passed by governance
func validateProposedIncentiveProgram(program IncentiveProgram) error {
	if program.ID != 0 {
		return ErrInvalidProgramID.Wrapf("%d", program.ID)
	}
	if !program.RemainingRewards.IsZero() {
		return ErrNonzeroRemainingRewards.Wrap(program.RemainingRewards.String())
	}
	if program.Funded {
		return ErrProposedFundedProgram
	}
	return program.Validate()
}

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

func validateSenderAssetTier(sender string, tier uint32, asset *sdk.Coin) error {
	if err := validateSenderAsset(sender, asset); err != nil {
		return err
	}
	if tier < 1 || tier > 3 {
		return ErrInvalidTier.Wrapf("%d", tier)
	}
	return nil
}

// Validate performs validation on an IncentiveProgram type returning an error
// if the program is invalid.
func (ip IncentiveProgram) Validate() error {
	if err := sdk.ValidateDenom(ip.UToken); err != nil {
		return err
	}
	if !leveragetypes.HasUTokenPrefix(ip.UToken) {
		// only allow uToken denoms as bonded denoms
		return errors.Wrap(leveragetypes.ErrNotUToken, ip.UToken)
	}

	if err := ip.TotalRewards.Validate(); err != nil {
		return err
	}
	if leveragetypes.HasUTokenPrefix(ip.TotalRewards.Denom) {
		// only allow base token denoms as rewards
		return errors.Wrap(leveragetypes.ErrUToken, ip.TotalRewards.Denom)
	}

	if err := ip.RemainingRewards.Validate(); err != nil {
		return err
	}

	// TODO #1749: Finish validate logic

	return nil
}
