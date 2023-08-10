package checkers

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	// imported to assure params are set before loading this package and we can correctly
	// initialize govModuleAddr
	_ "github.com/umee-network/umee/v6/app/params"
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

// WithEmergencyGroup is a copy of ugov.WithEmergencyGroup to avoid import cycle
type WithEmergencyGroup interface {
	EmergencyGroup() sdk.AccAddress
}

// EmergencyGroupAuthority returns true if the authority is EmergencyGroup. Returns false if
// authority is the x/gov address. Returns error otherwise.
// Note: we use WithEmergencyGroup rather than emergency group AccAddress to avoid storage read
// if it's not necessary.
func EmergencyGroupAuthority(authority string, eg WithEmergencyGroup) (bool, error) {
	if IsGovAuthority(authority) {
		return false, nil
	}
	a, err := sdk.AccAddressFromBech32(authority)
	if err != nil {
		return false, sdkerrors.ErrInvalidAddress.Wrapf("Authority: %v", err)
	}
	if !eg.EmergencyGroup().Equals(a) {
		return false, sdkerrors.ErrUnauthorized
	}
	return true, nil
}

// ValidateProposal checks the format of the title, description.
// If `requireGov=true` then authority must be a gov module address. Otherwise authority must be
// a correct bech32 address.
func ValidateProposal(title, description, authority string, requireGov bool) error {
	var err error
	if requireGov {
		err = AssertGovAuthority(authority)
	} else {
		_, err = sdk.AccAddressFromBech32(authority)
	}
	if err != nil {
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
