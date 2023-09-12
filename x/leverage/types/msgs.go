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
	_, _, _ sdk.Msg            = &MsgGovUpdateRegistry{}, &MsgGovUpdateSpecialAssets{}, &MsgGovSetParams{}
	_, _, _ legacytx.LegacyMsg = &MsgGovUpdateRegistry{}, &MsgGovUpdateSpecialAssets{}, &MsgGovSetParams{}
)

// NewMsgGovSetParams will create a new MsgGovSetParams instance.
// Authority must be a valid bech32 address.
func NewMsgGovSetParams(authority string, params Params) *MsgGovSetParams {
	return &MsgGovSetParams{
		Authority: authority,
		Params:    params,
	}
}

// String implements the Stringer interface.
func (msg MsgGovSetParams) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg MsgGovSetParams) ValidateBasic() error {
	if err := checkers.Proposal(msg.Authority, ""); err != nil {
		return err
	}

	return msg.Params.Validate()
}

// GetSigners implements Msg
func (msg MsgGovSetParams) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// NewMsgGovUpdateRegistry will create a new MsgUpdateRegistry instance.
// Authority must be a valid bech32 address.
func NewMsgGovUpdateRegistry(authority, description string, update, add []Token) *MsgGovUpdateRegistry {
	return &MsgGovUpdateRegistry{
		Authority:    authority,
		Description:  description,
		UpdateTokens: update,
		AddTokens:    add,
	}
}

// String implements the Stringer interface.
func (msg MsgGovUpdateRegistry) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg MsgGovUpdateRegistry) ValidateBasic() error {
	if err := checkers.Proposal(msg.Authority, msg.Description); err != nil {
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
	if err := checkers.Description(msg.Description); err != nil {
		return err
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
func (msg MsgGovSetParams) Type() string            { return sdk.MsgTypeURL(&msg) }
func (msg MsgGovUpdateRegistry) Route() string      { return "" }
func (msg MsgGovUpdateSpecialAssets) Route() string { return "" }
func (msg MsgGovSetParams) Route() string           { return "" }

func (msg MsgGovUpdateSpecialAssets) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgGovUpdateRegistry) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgGovSetParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
