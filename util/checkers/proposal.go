package checkers

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var govModuleAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()

// ValidateProposal checks the format of the title, description, and authority of a gov message.
func ValidateProposal(title, description, authority string) error {
	_, err := sdk.AccAddressFromBech32(authority)
	if err != nil {
		return err
	}

	if authority != govModuleAddr {
		return govtypes.ErrInvalidSigner.Wrapf(
			"invalid authority: expected %s, got %s",
			govModuleAddr, authority,
		)
	}

	if len(strings.TrimSpace(title)) == 0 {
		return errors.New("proposal title cannot be blank")
	}
	if len(title) > gov1b1.MaxTitleLength {
		return fmt.Errorf("proposal title is longer than max length of %d", gov1b1.MaxTitleLength)
	}

	if len(description) == 0 {
		return errors.New("proposal description cannot be blank")
	}
	if len(description) > gov1b1.MaxDescriptionLength {
		return fmt.Errorf(
			"proposal description is longer than max length of %d",
			gov1b1.MaxDescriptionLength,
		)
	}

	return nil
}
