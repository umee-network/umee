package checkers

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	// imported to assure params are set before loading this package and we can correctly
	// initialize govModuleAddr
	_ "github.com/umee-network/umee/v5/app/params"
)

// govModuleAddr is set during the first call of ValidateProposal
var govModuleAddr string

func init() {
	govModuleAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()
}

const minProposalTitleLen = 3

// AssertGovAuthority errors is the authority is not the gov module address. Panics if
// the gov module address is not set during the package initialization.
func AssertGovAuthority(authority string) error {
	if !IsGovAuthority(authority) {
		return govtypes.ErrInvalidSigner.Wrapf(
			"expected %s, got %s", govModuleAddr, authority)
	}
	return nil
}

// IsGovAuthority returns true if the authority is the gov module address. Panics if
// the gov module address is not set during the package initialization.
func IsGovAuthority(authority string) bool {
	if govModuleAddr == "" {
		panic("govModuleAddrs in the checkers package must be set before using this function")
	}
	return authority == govModuleAddr
}

// ValidateProposal checks the format of the title, description, and authority of a gov message.
func ValidateProposal(title, description, authority string) error {
	if err := AssertGovAuthority(authority); err != nil {
		return err
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

func ValidateAddr(addr, name string) error {
	if _, err := sdk.AccAddressFromBech32(addr); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid %s address: %s", name, err)
	}
	return nil
}
