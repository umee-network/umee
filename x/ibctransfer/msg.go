package ibctransfer

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
	_ sdk.Msg = &MsgUpdateIBCDenomsRateLimit{}
	_ sdk.Msg = &MsgUpdateIBCTransferPauseStatus{}
)

func NewIbcDenomsRateLimits(authority, title, description string,
	newIBCDenomsRateLimits, updateIBCDenomsRateLimits []MsgRateLimit,
) *MsgUpdateIBCDenomsRateLimit {
	return &MsgUpdateIBCDenomsRateLimit{
		Title:                     title,
		Description:               description,
		Authority:                 authority,
		NewIbcDenomsRateLimits:    newIBCDenomsRateLimits,
		UpdateIbcDenomsRateLimits: updateIBCDenomsRateLimits,
	}
}

// GetTitle returns the title of the proposal.
func (msg *MsgUpdateIBCDenomsRateLimit) GetTitle() string { return msg.Title }

// GetDescription returns the description of the proposal.
func (msg *MsgUpdateIBCDenomsRateLimit) GetDescription() string { return msg.Description }

// Route implements Msg
func (msg MsgUpdateIBCDenomsRateLimit) Route() string { return sdk.MsgTypeURL(&msg) }

// Type implements Msg
func (msg MsgUpdateIBCDenomsRateLimit) Type() string { return sdk.MsgTypeURL(&msg) }

// String implements the Stringer interface.
func (msg *MsgUpdateIBCDenomsRateLimit) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgUpdateIBCDenomsRateLimit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if err := validateAbstract(msg.Title, msg.Description); err != nil {
		return err
	}

	if err := validateRateLimitsOfIBCDenom(msg.NewIbcDenomsRateLimits); err != nil {
		return err
	}

	return validateRateLimitsOfIBCDenom(msg.UpdateIbcDenomsRateLimits)
}

// GetSignBytes implements Msg
func (msg *MsgUpdateIBCDenomsRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg *MsgUpdateIBCDenomsRateLimit) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

func validateAbstract(title, description string) error {
	if len(strings.TrimSpace(title)) == 0 {
		return types.ErrInvalidProposalContent.Wrap("proposal title cannot be blank")
	}
	if len(title) > gov1b1.MaxTitleLength {
		return types.ErrInvalidProposalContent.Wrapf("proposal title is longer than max length of %d",
			gov1b1.MaxTitleLength)
	}

	if len(description) == 0 {
		return types.ErrInvalidProposalContent.Wrap("proposal description cannot be blank")
	}
	if len(description) > gov1b1.MaxDescriptionLength {
		return types.ErrInvalidProposalContent.Wrapf("proposal description is longer than max length of %d",
			gov1b1.MaxDescriptionLength)
	}

	return nil
}

func validateRateLimitsOfIBCDenom(rateLimits []MsgRateLimit) error {
	for _, rateLimit := range rateLimits {
		if rateLimit.OutflowLimit.IsNil() {
			return ErrInvalidIBCDenom.Wrap("outflow limit shouldn't empty")
		}

		if rateLimit.OutflowLimit.IsNegative() {
			return ErrInvalidIBCDenom.Wrapf("outflow limit shouldn't be negative %s", rateLimit.OutflowLimit.String())
		}

		if len(rateLimit.IbcDenom) == 0 {
			return ErrInvalidIBCDenom.Wrap("ibc denom shouldn't empty")
		}
	}

	return nil
}

func NewUpdateIBCTransferPauseStatus(authority, title, description string,
	ibcPauseStatus bool,
) *MsgUpdateIBCTransferPauseStatus {
	return &MsgUpdateIBCTransferPauseStatus{
		Title:          title,
		Description:    description,
		Authority:      authority,
		IbcPauseStatus: ibcPauseStatus,
	}
}

// GetTitle returns the title of the proposal.
func (msg *MsgUpdateIBCTransferPauseStatus) GetTitle() string { return msg.Title }

// GetDescription returns the description of the proposal.
func (msg *MsgUpdateIBCTransferPauseStatus) GetDescription() string { return msg.Description }

// Route implements Msg
func (msg MsgUpdateIBCTransferPauseStatus) Route() string { return sdk.MsgTypeURL(&msg) }

// Type implements Msg
func (msg MsgUpdateIBCTransferPauseStatus) Type() string { return sdk.MsgTypeURL(&msg) }

// String implements the Stringer interface.
func (msg *MsgUpdateIBCTransferPauseStatus) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgUpdateIBCTransferPauseStatus) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	return validateAbstract(msg.Title, msg.Description)
}

// GetSignBytes implements Msg
func (msg *MsgUpdateIBCTransferPauseStatus) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg *MsgUpdateIBCTransferPauseStatus) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

func (r *RateLimit) Validate() error {
	if len(r.IbcDenom) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("ibcDenom shouldn't be empty")
	}

	return nil
}
