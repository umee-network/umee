package checkers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

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

func ValidateAddr(addr, name string) error {
	if _, err := sdk.AccAddressFromBech32(addr); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid %s address: %s", name, err)
	}
	return nil
}
