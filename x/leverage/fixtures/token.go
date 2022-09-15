package fixtures

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

const (
	// AtomDenom is an ibc denom to be used as ATOM's BaseDenom during testing
	AtomDenom = "ibc/CDC4587874B85BEA4FCEC3CEA5A1195139799A1FEE711A07D972537E18FDA39D"
)

// Token returns a valid token
func Token(base, symbol string) types.Token {
	return types.Token{
		BaseDenom:              base,
		SymbolDenom:            symbol,
		Exponent:               6,
		ReserveFactor:          sdk.MustNewDecFromStr("0.2"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.25"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.25"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.22"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("1.52"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.1"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0"),
		MaxSupply:              sdk.NewInt(100_000_000000),
	}
}
