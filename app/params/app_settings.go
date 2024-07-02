package params

import (
	"log"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	// Name defines the application name of the Umee network.
	Name = "umee"

	// BondDenom defines the native staking token denomination.
	// NOTE: it is used by IBC, and must not change to avoid token migration in all IBC chains.
	BondDenom      = "uumee"
	BaseExtraDenom = "uux"

	// DisplayDenom defines the name, symbol, and display value of the umee token.
	DisplayDenom = "UX"
	// Old display name. We renamed UMEE to UX.
	LegacyDisplayDenom = "UMEE"

	// DefaultGasLimit - set to the same value as cosmos-sdk flags.DefaultGasLimit
	// this value is currently only used in tests.
	DefaultGasLimit = 200000
)

// ProtocolMinGasPrice is a consensus controlled gas price. Each validator must set his
// `minimum-gas-prices` in app.toml config to value above ProtocolMinGasPrice.
// Transactions with gas-price smaller than ProtocolMinGasPrice will fail during DeliverTx.
var ProtocolMinGasPrice = sdk.NewDecCoinFromDec(BondDenom, sdkmath.LegacyMustNewDecFromStr("0.00"))

func init() {
	// XXX: If other upstream or external application's depend on any of Umee's
	// CLI or command functionality, then this would require us to move the
	// SetAddressConfig call to somewhere external such as the root command
	// constructor and anywhere else we contract the app.
	SetAddressConfig()

	if AccountAddressPrefix != Name {
		log.Fatal("AccountAddresPrefix must equal Name")
	}

	sdk.SetCoinDenomRegex(func() string {
		// allow "ux" token. Modify the default regexp to allow 2 character long denoms.
		return `[a-zA-Z][a-zA-Z0-9/:._-]{1,127}`
	})
}

// UmeeTokenMetadata creates bank Metadata for the UX token
func UmeeTokenMetadata() banktypes.Metadata {
	return banktypes.Metadata{
		Description: "The native staking token of the Umee network.",
		Base:        BondDenom, // NOTE: must not change
		Name:        DisplayDenom,
		Display:     DisplayDenom,
		Symbol:      LegacyDisplayDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    BondDenom,
				Exponent: 0,
				Aliases: []string{
					"microumee", BaseExtraDenom,
				},
			}, {
				Denom:    DisplayDenom,
				Exponent: 6,
				Aliases:  []string{LegacyDisplayDenom},
			},
		},
	}
}
