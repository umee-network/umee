package fixtures

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

const (
	// AtomDenom is an ibc denom to be used as ATOM's BaseDenom during testing. Matches mainnet.
	AtomDenom = "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"
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
