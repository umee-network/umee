package params

import (
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Name defines the application name of the Umee network.
	// NOTE: do not change this.
	Name = "umee"

	// IBC denom is the base denom for IBC and Metadata.Base.
	// NOTE: it is used by IBC, and must not change to avoid token migration in all IBC chains.
	IBCBaseDenom = "uumee"

	// BaseDenom defines the native staking an d fee token denom.
	BaseDenom = "uux"

	// DisplayDenom defines the name, symbol, and display value of the umee token.
	DisplayDenom = "UX"

	// Old dispaly name. We renamed UMEE to UX.
	LegacyDisplayDenom = "UMEE"

	// DefaultGasLimit - set to the same value as cosmos-sdk flags.DefaultGasLimit
	// this value is currently only used in tests.
	DefaultGasLimit = 200000
)

// ProtocolMinGasPrice is a consensus controlled gas price. Each validator must set his
// `minimum-gas-prices` in app.toml config to value above ProtocolMinGasPrice.
// Transactions with gas-price smaller than ProtocolMinGasPrice will fail during DeliverTx.
var ProtocolMinGasPrice = sdk.NewDecCoinFromDec(BaseDenom, sdk.MustNewDecFromStr("0.00"))

func init() {
	// XXX: If other upstream or external application's depend on any of Umee's
	// CLI or command functionality, then this would require us to move the
	// SetAddressConfig call to somewhere external such as the root command
	// constructor and anywhere else we contract the app.
	SetAddressConfig()

	if AccountAddressPrefix != Name {
		log.Fatal("AccountAddresPrefix must equal Name")
	}
}
