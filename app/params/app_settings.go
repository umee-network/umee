package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Name defines the application name of the Umee network.
	Name = "umee"

	// BondDenom defines the native staking token denomination.
	BondDenom = "uumee"

	// DisplayDenom defines the name, symbol, and display value of the umee token.
	DisplayDenom = "UMEE"

	// MaxAddrLen is the maximum allowed length (in bytes) for an address.
	//
	// NOTE: In the SDK, the default value is 255.
	MaxAddrLen = 20
)

var (
	// MinMinGasPrice is the minimum value a validator can set for `minimum-gas-prices` his app.toml config
	MinMinGasPrice = sdk.NewDecCoinFromDec(BondDenom, sdk.MustNewDecFromStr("0.05"))
)
