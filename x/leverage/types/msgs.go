package types

import (
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/umee-network/umee/v6/util/checkers"
	"gopkg.in/yaml.v3"
)

var (
	_, _ sdk.Msg            = &MsgGovUpdateRegistry{}, &MsgGovUpdateSpecialAssets{}
	_, _ legacytx.LegacyMsg = &MsgGovUpdateRegistry{}, &MsgGovUpdateSpecialAssets{}
)

// NewMsgGovUpdateRegistry will create a new MsgUpdateRegistry instance.
// Authority must be a valid bech32 address.
func NewMsgGovUpdateRegistry(authority string, update, add []Token) *MsgGovUpdateRegistry {
	return &MsgGovUpdateRegistry{
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
	if err := checkers.ValidateProposal(msg.Description, msg.Authority); err != nil {
		return err
	}

	if len(msg.AddTokens) == 0 && len(msg.UpdateTokens) == 0 {
		return ErrEmptyAddAndUpdateTokens
	}

	if err := validateRegistryToken(msg.AddTokens); err != nil {
		return err
	}
	return validateRegistryToken(msg.UpdateTokens)
}

// GetSigners implements Msg
func (msg MsgGovUpdateRegistry) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// validateRegistryToken returns error if duplicate baseDenom exists.
func validateRegistryToken(tokens []Token) error {
	tokenDenoms := map[string]bool{}
	for _, token := range tokens {
		if _, ok := tokenDenoms[token.BaseDenom]; ok {
			return ErrDuplicateToken.Wrapf("with baseDenom %s", token.BaseDenom)
		}
		tokenDenoms[token.BaseDenom] = true
		if err := token.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// NewMsgGovUpdateSpecialAssets will create a new MsgGovUpdateSpecialAssets instance
func NewMsgGovUpdateSpecialAssets(authority string, sets []SpecialAssetSet, pairs []SpecialAssetPair,
) *MsgGovUpdateSpecialAssets {
	return &MsgGovUpdateSpecialAssets{
		Authority: authority,
		Sets:      sets,
		Pairs:     pairs,
	}
}

// GetSigners implements Msg
func (msg MsgGovUpdateSpecialAssets) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// String implements the Stringer interface.
func (msg MsgGovUpdateSpecialAssets) String() string {
	// return fmt.Sprintf("<authority: %s, min_gas_price: %s>", msg.Authority, msg.MinGasPrice.String())
	return fmt.Sprintf(
		"authority: %s, sets: %s, pairs: %s",
		msg.Authority,
		msg.Sets,
		msg.Pairs,
	)
}

// ValidateBasic implements Msg
func (msg MsgGovUpdateSpecialAssets) ValidateBasic() error {
	// Today we only accept x/gov
	if err := checkers.AssertGovAuthority(msg.Authority); err != nil {
		return err
	}
	if len(msg.Description) != 0 {
		return fmt.Errorf("for x/gov proposals, description must be empty, and the x/gov proposal metadata should be used instead")
	}

	if len(msg.Pairs) == 0 && len(msg.Sets) == 0 {
		return fmt.Errorf("empty special asset pairs update")
	}

	if err := validateSpecialAssetPairs(msg.Pairs); err != nil {
		return err
	}

	ascendingWeight := sdk.ZeroDec()
	for _, set := range msg.Sets {
		// ensures sets are sorted from lowest to highest collateral weight
		// to ensure overlapping sets cause the higher collateral weight to
		// be stored in state
		if set.CollateralWeight.IsPositive() {
			if set.CollateralWeight.LT(ascendingWeight) {
				return fmt.Errorf("asset sets not in ascending (weight) order")
			}
			ascendingWeight = set.CollateralWeight
		}
		if err := set.Validate(); err != nil {
			return errors.Wrapf(err, "special asset set [%s]", set.String())
		}
	}

	return nil
}

// validateSpecialAssetPairs returns error if duplicate special asset pairs exist or
// if any individual pairs are invalid.
func validateSpecialAssetPairs(pairs []SpecialAssetPair) error {
	for _, pair := range pairs {
		if err := pair.Validate(); err != nil {
			return err
		}
	}
	assetPairs := map[string]bool{}
	for _, pair := range pairs {
		s := pair.Collateral + "," + pair.Borrow
		if _, ok := assetPairs[s]; ok {
			return fmt.Errorf("duplicate special asset pair: %s", s)
		}
		assetPairs[s] = true
	}
	return nil
}

// LegacyMsg.Type implementations

func (msg MsgGovUpdateRegistry) Type() string       { return sdk.MsgTypeURL(&msg) }
func (msg MsgGovUpdateSpecialAssets) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg MsgGovUpdateRegistry) Route() string      { return "" }
func (msg MsgGovUpdateSpecialAssets) Route() string { return "" }

func (msg MsgGovUpdateSpecialAssets) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgGovUpdateRegistry) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
