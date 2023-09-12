package checkers

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/v6/tests/accs"
)

func TestIsGovAuthority(t *testing.T) {
	require := require.New(t)
	expectedGovAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	bankAddr := authtypes.NewModuleAddress(banktypes.ModuleName).String()

	tcs := []struct {
		name  string
		auth  string
		isErr bool
	}{
		{"validAddr", expectedGovAddr, false},
		{"invalid: empty addr", "", true},
		{"invalid: addr", bankAddr, true},
		{"invalid: addr", accs.Bob.String(), true},
	}

	for i, tc := range tcs {
		err := AssertGovAuthority(tc.auth)
		if tc.isErr {
			require.ErrorIs(err, govtypes.ErrInvalidSigner, "[test: %d] expected error", i)
		} else {
			require.NoError(err, "[test: %d] expected error", i)
		}
	}
}

func TestProposal(t *testing.T) {
	require := require.New(t)
	expectedGovAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	tcs := []struct {
		name  string
		auth  string
		descr string
		isErr bool
	}{
		{"x/gov", expectedGovAddr, "", false},
		{"not gov, good descr", accs.Bob.String(), "some description", false},

		{"invalid: empty addr", "", "", true},
		{"invalid: addr", "XYZ", "", true},
		{"x/gov, non empty descr", expectedGovAddr, "", false},
		{"not gov, empty descr", accs.Bob.String(), "", true},
	}

	for i, tc := range tcs {
		err := Proposal(tc.auth, tc.descr)
		if tc.isErr {
			require.Error(err, "test %d: %s", i, tc.name)
		} else {
			require.NoError(err, "test %d: %s", i, tc.name)
		}
	}
}
