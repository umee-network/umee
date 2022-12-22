package checkers

import (
	"errors"
	"fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// govModuleAddr is set during the first call of ValidateProposal
var govModuleAddr string

func SetGovModuleAddr(addr string) {
	if govModuleAddr != "" {
		panic("gov module address already set in the checkers package")
	}
	govModuleAddr = addr
}

const minProposalTitleLen = 3

// ValidateProposal checks the format of the title, description, and authority of a gov message.
func ValidateProposal(title, description, authority string) error {
	if govModuleAddr == "" {
		return sdkerrors.ErrLogic.Wrap("govModuleAddrs in the checkers package must be set before using this function")
	}
	if authority != govModuleAddr {
		return govtypes.ErrInvalidSigner.Wrapf(
			"invalid authority: expected %s, got %s",
			govModuleAddr, authority,
		)
	}

	if len(strings.TrimSpace(title)) < minProposalTitleLen {
		return fmt.Errorf("proposal title must be at least %d of non blank characters",
			minProposalTitleLen)
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
