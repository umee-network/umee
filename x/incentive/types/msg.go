package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/umee-network/umee/v3/util/checkers"
)

var (
	_ sdk.Msg = &MsgClaim{}
	_ sdk.Msg = &MsgBeginUnbonding{}
	_ sdk.Msg = &MsgBond{}
	_ sdk.Msg = &MsgSponsor{}
	_ sdk.Msg = &MsgGovCreateProgram{}
	_ sdk.Msg = &MsgGovCreateAndSponsorProgram{}
)

func NewMsgClaim(account sdk.AccAddress) *MsgClaim {
	return &MsgClaim{
		Account: account.String(),
	}
}

func (msg *MsgClaim) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Account)
	return err
}

func (msg *MsgClaim) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Account)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgClaim) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgBond(account sdk.AccAddress, tier uint32, asset sdk.Coin) *MsgBond {
	return &MsgBond{
		Account: account.String(),
		Tier:    tier,
		Asset:   asset,
	}
}

func (msg *MsgBond) ValidateBasic() error {
	return validateSenderAssetTier(msg.Account, msg.Tier, &msg.Asset)
}

func (msg *MsgBond) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Account)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgBond) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgBeginUnbonding(account sdk.AccAddress, tier uint32, asset sdk.Coin) *MsgBeginUnbonding {
	return &MsgBeginUnbonding{
		Account: account.String(),
		Tier:    tier,
		Asset:   asset,
	}
}

func (msg *MsgBeginUnbonding) ValidateBasic() error {
	return validateSenderAssetTier(msg.Account, msg.Tier, &msg.Asset)
}

func (msg *MsgBeginUnbonding) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Account)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgBeginUnbonding) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgSponsor(sponsor sdk.AccAddress, programID uint64, asset sdk.Coin) *MsgSponsor {
	return &MsgSponsor{
		Sponsor: sponsor.String(),
		Program: programID,
		Asset:   asset,
	}
}

func (msg *MsgSponsor) ValidateBasic() error {
	if msg.Program == 0 {
		return ErrInvalidProgramID.Wrapf("%d", msg.Program)
	}
	return validateSenderAsset(msg.Sponsor, &msg.Asset)
}

func (msg *MsgSponsor) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Sponsor)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgSponsor) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgCreateProgram(authority, title, description string, program IncentiveProgram) *MsgGovCreateProgram {
	return &MsgGovCreateProgram{
		Title:       title,
		Description: description,
		Program:     program,
		Authority:   authority,
	}
}

func (msg MsgGovCreateProgram) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}
	if err := validateIncentiveProposal(msg.Title, msg.Description, msg.Program); err != nil {
		return err
	}
	return nil
}

// GetSignBytes implements Msg
func (msg MsgGovCreateProgram) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgGovCreateProgram) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

func NewMsgCreateAndSponsorProgram(authority, title, description, sponsor string, program IncentiveProgram,
) *MsgGovCreateAndSponsorProgram {
	return &MsgGovCreateAndSponsorProgram{
		Title:       title,
		Description: description,
		Program:     program,
		Authority:   authority,
		Sponsor:     sponsor,
	}
}

func (msg MsgGovCreateAndSponsorProgram) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Sponsor); err != nil {
		return sdkerrors.Wrap(err, "invalid sponsor address")
	}
	if err := validateIncentiveProposal(msg.Title, msg.Description, msg.Program); err != nil {
		return err
	}
	return nil
}

// GetSignBytes implements Msg
func (msg MsgGovCreateAndSponsorProgram) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgGovCreateAndSponsorProgram) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

func validateIncentiveProposal(title, description string, program IncentiveProgram) error {
	if len(strings.TrimSpace(title)) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(title) > gov1b1.MaxTitleLength {
		return sdkerrors.Wrapf(types.ErrInvalidProposalContent, "proposal title is longer than max length of %d",
			gov1b1.MaxTitleLength)
	}

	if len(description) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidProposalContent, "proposal description cannot be blank")
	}
	if len(description) > gov1b1.MaxDescriptionLength {
		return sdkerrors.Wrapf(types.ErrInvalidProposalContent, "proposal description is longer than max length of %d",
			gov1b1.MaxDescriptionLength)
	}

	if program.Id != 0 {
		return ErrInvalidProgramID.Wrapf("%d", program.Id)
	}
	if !program.RemainingRewards.IsZero() {
		return ErrNonzeroRemainingRewards.Wrap(program.RemainingRewards.String())
	}
	if !program.FundedRewards.IsZero() {
		return ErrNonzeroFundedRewards.Wrap(program.FundedRewards.String())
	}
	if err := program.Validate(); err != nil {
		return err
	}
	return nil
}

func validateSenderAsset(sender string, asset *sdk.Coin) error {
	_, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrNilAsset
	}
	if err := asset.Validate(); err != nil {
		return err
	}
	return nil
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
