package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/umee-network/umee/v3/util/checkers"
	"gopkg.in/yaml.v3"
)

var _ sdk.Msg = &MsgGovUpdateRegistry{}

// NewMsgUpdateRegistry will creates a new MsgUpdateRegistry instance
func NewMsgUpdateRegistry(authority, title, description string, updateTokens, addTokens []Token) *MsgGovUpdateRegistry {
	return &MsgGovUpdateRegistry{
		Title:        title,
		Description:  description,
		UpdateTokens: updateTokens,
		AddTokens:    addTokens,
		Authority:    authority,
	}
}

// GetTitle returns the title of the proposal.
func (msg *MsgGovUpdateRegistry) GetTitle() string { return msg.Title }

// GetDescription returns the description of the proposal.
func (msg *MsgGovUpdateRegistry) GetDescription() string { return msg.Description }

// Type implements Msg
func (msg MsgGovUpdateRegistry) Type() string { return sdk.MsgTypeURL(&msg) }

// String implements the Stringer interface.
func (msg MsgGovUpdateRegistry) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg MsgGovUpdateRegistry) ValidateBasic() error {
	if err := checkers.ValidateProposal(msg.Title, msg.Description, msg.Authority); err != nil {
		return err
	}

	if err := validateRegistryTokenDenoms(msg.AddTokens); err != nil {
		return err
	}

	for _, token := range msg.AddTokens {
		if err := token.Validate(); err != nil {
			return sdkerrors.Wrap(err, "token")
		}
	}

	if err := validateRegistryTokenDenoms(msg.UpdateTokens); err != nil {
		return err
	}

	for _, token := range msg.UpdateTokens {
		if err := token.Validate(); err != nil {
			return sdkerrors.Wrap(err, "token")
		}
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgGovUpdateRegistry) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgGovUpdateRegistry) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// validateRegistryTokenDenoms returns error if duplicate baseDenom exists.
func validateRegistryTokenDenoms(tokens []Token) error {
	tokenDenoms := map[string]bool{}
	for _, token := range tokens {
		if _, ok := tokenDenoms[token.BaseDenom]; ok {
			return sdkerrors.Wrapf(ErrDuplicateToken, "duplicate token with baseDenom %s", token.BaseDenom)
		}
		tokenDenoms[token.BaseDenom] = true
	}
	return nil
}

func validateProposal(title, description string) error {
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

	return nil
}
