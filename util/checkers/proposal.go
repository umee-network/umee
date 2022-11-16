package checkers

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// ValidateProposal checks the format of the title, description, and authority of a gov message.
func ValidateProposal(title, description, authority string) error {
	_, err := sdk.AccAddressFromBech32(authority)
	if err != nil {
		return err
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
