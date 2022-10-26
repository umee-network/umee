package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"gopkg.in/yaml.v3"

	"github.com/umee-network/umee/v3/util/checkers"
)

var (
	_ sdk.Msg = &MsgClaim{}
	_ sdk.Msg = &MsgBeginUnbonding{}
	_ sdk.Msg = &MsgLock{}
	_ sdk.Msg = &MsgSponsor{}
	_ sdk.Msg = &MsgCreateProgram{}
)

func NewMsgClaim(supplier sdk.AccAddress) *MsgClaim {
	return &MsgClaim{
		Supplier: supplier.String(),
	}
}

func (msg MsgClaim) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgClaim) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg *MsgClaim) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Supplier)
	return err
}

func (msg *MsgClaim) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgClaim) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgLock(supplier sdk.AccAddress, tier uint32, asset sdk.Coin) *MsgLock {
	return &MsgLock{
		Supplier: supplier.String(),
		Tier:     tier,
		Asset:    asset,
	}
}

func (msg MsgLock) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgLock) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg *MsgLock) ValidateBasic() error {
	return validateSenderAssetTier(msg.Supplier, msg.Tier, &msg.Asset)
}

func (msg *MsgLock) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
}

// GetSignBytes get the bytes for the message signer to sign on
func (msg *MsgLock) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgBeginUnbonding(supplier sdk.AccAddress, tier uint32, asset sdk.Coin) *MsgBeginUnbonding {
	return &MsgBeginUnbonding{
		Supplier: supplier.String(),
		Tier:     tier,
		Asset:    asset,
	}
}

func (msg MsgBeginUnbonding) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgBeginUnbonding) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg *MsgBeginUnbonding) ValidateBasic() error {
	return validateSenderAssetTier(msg.Supplier, msg.Tier, &msg.Asset)
}

func (msg *MsgBeginUnbonding) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Supplier)
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

func (msg MsgSponsor) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgSponsor) Type() string  { return sdk.MsgTypeURL(&msg) }

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

func NewMsgCreateProgram(authority, title, description string, program IncentiveProgram) *MsgCreateProgram {
	return &MsgCreateProgram{
		Title:       title,
		Description: description,
		Program:     program,
		Authority:   authority,
	}
}

func (msg MsgCreateProgram) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgCreateProgram) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg MsgCreateProgram) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg *MsgCreateProgram) GetTitle() string { return msg.Title }

func (msg *MsgCreateProgram) GetDescription() string { return msg.Description }

func (msg MsgCreateProgram) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if err := msg.ValidateAbstract(); err != nil {
		return err
	}

	/*
		TODO: General validation
		if err := msg.Program.Validate(); err != nil {
			return err
		}
	*/

	// Additional rules apply to incentive programs which are still being proposed
	if msg.Program.Id != 0 {
		return ErrInvalidProgramID.Wrapf("%d", msg.Program.Id)
	}
	if !msg.Program.RemainingRewards.IsZero() {
		return ErrNonzeroRemainingRewards.Wrap(msg.Program.RemainingRewards.String())
	}
	if !msg.Program.FundedRewards.IsZero() {
		return ErrNonzeroFundedRewards.Wrap(msg.Program.FundedRewards.String())
	}

	return nil
}

func (c *MsgCreateProgram) ValidateAbstract() error {
	title := c.GetTitle()
	if len(strings.TrimSpace(title)) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(title) > gov1b1.MaxTitleLength {
		return sdkerrors.Wrapf(types.ErrInvalidProposalContent, "proposal title is longer than max length of %d",
			gov1b1.MaxTitleLength)
	}

	description := c.GetDescription()
	if len(description) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidProposalContent, "proposal description cannot be blank")
	}
	if len(description) > gov1b1.MaxDescriptionLength {
		return sdkerrors.Wrapf(types.ErrInvalidProposalContent, "proposal description is longer than max length of %d",
			gov1b1.MaxDescriptionLength)
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgCreateProgram) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgCreateProgram) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

func NewMsgCreateAndSponsorProgram(authority, title, description, sponsor string, program IncentiveProgram,
) *MsgCreateAndSponsorProgram {
	return &MsgCreateAndSponsorProgram{
		Title:       title,
		Description: description,
		Program:     program,
		Authority:   authority,
		Sponsor:     sponsor,
	}
}

func (msg MsgCreateAndSponsorProgram) Route() string { return sdk.MsgTypeURL(&msg) }
func (msg MsgCreateAndSponsorProgram) Type() string  { return sdk.MsgTypeURL(&msg) }

func (msg MsgCreateAndSponsorProgram) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg *MsgCreateAndSponsorProgram) GetTitle() string { return msg.Title }

func (msg *MsgCreateAndSponsorProgram) GetDescription() string { return msg.Description }

func (msg MsgCreateAndSponsorProgram) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if err := msg.ValidateAbstract(); err != nil {
		return err
	}

	/*
		TODO: General validation
		if err := msg.Program.Validate(); err != nil {
			return err
		}
	*/

	// Additional rules apply to incentive programs which are still being proposed
	if msg.Program.Id != 0 {
		return ErrInvalidProgramID.Wrapf("%d", msg.Program.Id)
	}
	if !msg.Program.RemainingRewards.IsZero() {
		return ErrNonzeroRemainingRewards.Wrap(msg.Program.RemainingRewards.String())
	}
	if !msg.Program.FundedRewards.IsZero() {
		return ErrNonzeroFundedRewards.Wrap(msg.Program.FundedRewards.String())
	}

	return nil
}

func (c *MsgCreateAndSponsorProgram) ValidateAbstract() error {
	title := c.GetTitle()
	if len(strings.TrimSpace(title)) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(title) > gov1b1.MaxTitleLength {
		return sdkerrors.Wrapf(types.ErrInvalidProposalContent, "proposal title is longer than max length of %d",
			gov1b1.MaxTitleLength)
	}

	description := c.GetDescription()
	if len(description) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidProposalContent, "proposal description cannot be blank")
	}
	if len(description) > gov1b1.MaxDescriptionLength {
		return sdkerrors.Wrapf(types.ErrInvalidProposalContent, "proposal description is longer than max length of %d",
			gov1b1.MaxDescriptionLength)
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgCreateAndSponsorProgram) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgCreateAndSponsorProgram) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
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
