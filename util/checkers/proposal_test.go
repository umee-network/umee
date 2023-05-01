package checkers

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
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
		{"invalid: empty addr", bankAddr, true},
	}

	for i, tc := range tcs {
		err := IsGovAuthority(tc.auth)
		if tc.isErr {
			require.ErrorContains(err, "invalid authority", "[test: %d] expected error", i)
		} else {
			require.NoError(err, "[test: %d] expected error", i)
		}
	}
}
