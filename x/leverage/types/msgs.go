package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/umee-network/umee/v5/util/checkers"
	"gopkg.in/yaml.v3"
)

var (
	_, _ sdk.Msg            = &MsgGovUpdateRegistry{}, &MsgGovUpdateSpecialAssetPairs{}
	_, _ legacytx.LegacyMsg = &MsgGovUpdateRegistry{}, &MsgGovUpdateSpecialAssetPairs{}
)

// NewMsgGovUpdateRegistry will create a new MsgUpdateRegistry instance
func NewMsgGovUpdateRegistry(authority, title, description string, update, add []Token) *MsgGovUpdateRegistry {
	return &MsgGovUpdateRegistry{
		Title:        title,
		Description:  description,
		UpdateTokens: update,
		AddTokens:    add,
		Authority:    authority,
	}
}

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

// NewMsgGovUpdateSpecialAssetPairs will create a new MsgGovUpdateSpecialAssetPairs instance
func NewMsgGovUpdateSpecialAssetPairs(authority string, sets []SpecialAssetSet, pairs []SpecialAssetPair,
) *MsgGovUpdateSpecialAssetPairs {
	return &MsgGovUpdateSpecialAssetPairs{
		Authority: authority,
		Sets:      sets,
		Pairs:     pairs,
	}
}

// GetSigners implements Msg
func (msg MsgGovUpdateSpecialAssetPairs) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// String implements the Stringer interface.
func (msg MsgGovUpdateSpecialAssetPairs) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg MsgGovUpdateSpecialAssetPairs) ValidateBasic() error {
	if err := checkers.IsGovAuthority(msg.Authority); err != nil {
		return err
	}

	if len(msg.Pairs) == 0 {
		return ErrEmptyUpdateSpecialAssetPairs
	}

	if err := validateSpecialAssetPairDenoms(msg.Pairs); err != nil {
		return err
	}

	ascendingWeight := sdk.ZeroDec()
	for _, set := range msg.Sets {
		// ensures sets are sorted from lowest to highest collateral weight
		// to ensure overlapping sets cause the higher collateral weight to
		// be stored in state
		if set.CollateralWeight.IsPositive() {
			if set.CollateralWeight.LT(ascendingWeight) {
				return ErrProposedSetOrder
			}
			ascendingWeight = set.CollateralWeight
		}
		if err := set.Validate(); err != nil {
			return errors.Wrapf(err, "special asset set [%s]", set.String())
		}
	}

	for _, pair := range msg.Pairs {
		if err := pair.Validate(); err != nil {
			return errors.Wrapf(err, "special asset pair [%s, %s]", pair.Collateral, pair.Borrow)
		}
	}

	return nil
}

// validateSpecialAssetPairDenoms returns error if duplicate special asset pairs exist.
func validateSpecialAssetPairDenoms(pairs []SpecialAssetPair) error {
	assetPairs := map[string]bool{}
	for _, pair := range pairs {
		s := pair.Collateral + "," + pair.Borrow
		if _, ok := assetPairs[s]; ok {
			return ErrDuplicatePair.Wrapf("[%s, %s]", pair.Collateral, pair.Borrow)
		}
		assetPairs[s] = true
	}
	return nil
}

// LegacyMsg.Type implementations

func (msg MsgGovUpdateRegistry) Type() string           { return sdk.MsgTypeURL(&msg) }
func (msg MsgGovUpdateSpecialAssetPairs) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgGovUpdateRegistry) Route() string          { return "" }
func (msg MsgGovUpdateSpecialAssetPairs) Route() string { return "" }

func (msg MsgGovUpdateSpecialAssetPairs) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgGovUpdateRegistry) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
