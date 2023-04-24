package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/checkers"
	"gopkg.in/yaml.v3"
)

var _ sdk.Msg = &MsgGovUpdateRegistry{}

// NewMsgUpdateRegistry will create a new MsgUpdateRegistry instance
func NewMsgUpdateRegistry(authority, title, description string, updateTokens, addTokens []Token) *MsgGovUpdateRegistry {
	return &MsgGovUpdateRegistry{
		Title:        title,
		Description:  description,
		UpdateTokens: updateTokens,
		AddTokens:    addTokens,
		Authority:    authority,
	}
}

// Type implements Msg interface
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

	if len(msg.AddTokens) == 0 && len(msg.UpdateTokens) == 0 {
		return ErrEmptyAddAndUpdateTokens
	}

	if err := validateRegistryTokenDenoms(msg.AddTokens); err != nil {
		return err
	}

	for _, token := range msg.AddTokens {
		if err := token.Validate(); err != nil {
			return errors.Wrap(err, "token")
		}
	}

	if err := validateRegistryTokenDenoms(msg.UpdateTokens); err != nil {
		return err
	}

	for _, token := range msg.UpdateTokens {
		if err := token.Validate(); err != nil {
			return errors.Wrap(err, "token")
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
			return ErrDuplicateToken.Wrapf("duplicate token with baseDenom %s", token.BaseDenom)
		}
		tokenDenoms[token.BaseDenom] = true
	}
	return nil
}
