// package accs provides test accounts for testing purposes
package accs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto"
)

// Test user accounts
//
//nolint:lll
var (
	Alice         = sdk.MustAccAddressFromBech32("umee1yesmdu06f7strl67kjvg2w7t5kacc97taczr47")
	AliceMenmonic = "paper intact wine brother wrist sniff cheese garbage differ save chase hospital wine sock lobster scene border height gas dad tornado wrist tone pause"

	Bob         = sdk.MustAccAddressFromBech32("umee186tgjft0gqqafxmyf88zh27wxs8jexw0hkycps")
	BobMenmonic = "desert tube fan laugh gold beyond urban bicycle above sunny circle lake shaft space demise pony betray party benefit climb start ordinary attack input"
)

// Test module accounts
var (
	FooModule = authtypes.NewModuleAddress("foomodule")
)

// MkAddress creats an address from the given seed string
func MkAddress(seed string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(seed)))
}
