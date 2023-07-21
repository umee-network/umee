package metoken

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/v5/util/checkers"
)

var (
	_ sdk.Msg = &MsgSwap{}
	_ sdk.Msg = &MsgRedeem{}
	_ sdk.Msg = &MsgGovSetParams{}
	_ sdk.Msg = &MsgGovUpdateRegistry{}
)

func NewMsgSwap(user sdk.AccAddress, asset sdk.Coin, metokenDenom string) *MsgSwap {
	return &MsgSwap{
		User:         user.String(),
		Asset:        asset,
		MetokenDenom: metokenDenom,
	}
}

// ValidateBasic implements Msg
func (msg *MsgSwap) ValidateBasic() error {
	return validateUserAndAssetAndDenom(msg.User, &msg.Asset, msg.MetokenDenom)
}

// GetSigners implements Msg
func (msg *MsgSwap) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.User)
}

// LegacyMsg.Type implementations
func (msg MsgSwap) Route() string { return "" }

func (msg MsgSwap) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgSwap) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgRedeem(user sdk.AccAddress, metoken sdk.Coin, assetDenom string) *MsgRedeem {
	return &MsgRedeem{
		User:       user.String(),
		Metoken:    metoken,
		AssetDenom: assetDenom,
	}
}

// ValidateBasic implements Msg
func (msg *MsgRedeem) ValidateBasic() error {
	return validateUserAndAssetAndDenom(msg.User, &msg.Metoken, msg.AssetDenom)
}

// GetSigners implements Msg
func (msg *MsgRedeem) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.User)
}

// LegacyMsg.Type implementations
func (msg MsgRedeem) Route() string { return "" }

func (msg MsgRedeem) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgRedeem) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgGovSetParams(authority string, params Params) *MsgGovSetParams {
	return &MsgGovSetParams{
		Authority: authority,
		Params:    params,
	}
}

// ValidateBasic implements Msg
func (msg *MsgGovSetParams) ValidateBasic() error {
	return checkers.IsGovAuthority(msg.Authority)
}

// GetSigners implements Msg
func (msg *MsgGovSetParams) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// LegacyMsg.Type implementations
func (msg MsgGovSetParams) Route() string { return "" }

func (msg MsgGovSetParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgGovSetParams) Type() string { return sdk.MsgTypeURL(&msg) }

func NewMsgGovUpdateRegistry(authority string, addIndex, updateIndex []Index) *MsgGovUpdateRegistry {
	return &MsgGovUpdateRegistry{
		Authority:   authority,
		AddIndex:    addIndex,
		UpdateIndex: updateIndex,
	}
}

// ValidateBasic implements Msg
func (msg *MsgGovUpdateRegistry) ValidateBasic() error {
	if err := checkers.IsGovAuthority(msg.Authority); err != nil {
		return err
	}

	if len(msg.AddIndex) == 0 && len(msg.UpdateIndex) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("empty add and update indexes")
	}

	if err := validateDuplicates(msg.AddIndex, msg.UpdateIndex); err != nil {
		return err
	}

	for _, index := range msg.AddIndex {
		if err := index.Validate(); err != nil {
			return errors.Wrapf(err, "addIndex: %s", index.Denom)
		}
	}

	for _, index := range msg.UpdateIndex {
		if err := index.Validate(); err != nil {
			return errors.Wrapf(err, "updateIndex: %s", index.Denom)
		}
	}

	return nil
}

// GetSigners implements Msg
func (msg *MsgGovUpdateRegistry) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// LegacyMsg.Type implementations
func (msg MsgGovUpdateRegistry) Route() string { return "" }

func (msg MsgGovUpdateRegistry) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgGovUpdateRegistry) Type() string { return sdk.MsgTypeURL(&msg) }

func validateUserAndAssetAndDenom(sender string, asset *sdk.Coin, denom string) error {
	if _, err := sdk.AccAddressFromBech32(sender); err != nil {
		return err
	}
	if asset == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("nil asset")
	}
	if err := asset.Validate(); err != nil {
		return err
	}

	return sdk.ValidateDenom(denom)
}

func validateDuplicates(addIndex, updateIndex []Index) error {
	indexes := make(map[string]struct{})
	for _, index := range addIndex {
		if _, ok := indexes[index.Denom]; ok {
			return sdkerrors.ErrInvalidRequest.Wrapf(
				"duplicate addIndex metoken denom %s",
				index.Denom,
			)
		}
		indexes[index.Denom] = struct{}{}
	}

	for _, index := range updateIndex {
		if _, ok := indexes[index.Denom]; ok {
			return sdkerrors.ErrInvalidRequest.Wrapf(
				"duplicate updateIndex metoken denom %s",
				index.Denom,
			)
		}
		indexes[index.Denom] = struct{}{}
	}

	return nil
}
