package fixtures

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

const (
	UmeeDenom = "uumee"
	// AtomDenom is an ibc denom to be used as ATOM's BaseDenom during testing. Matches mainnet.
	AtomDenom = "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"
	// DaiDenom is an ibc denom to be used as DAI's BaseDenom during testing. Matches mainnet.
	DaiDenom = "ibc/C86651B4D30C1739BF8B061E36F4473A0C9D60380B52D01E56A6874037A5D060"
)

// Token returns a valid token
func Token(base, symbol string, exponent uint32) types.Token {
	return types.Token{
		BaseDenom:              base,
		SymbolDenom:            symbol,
		Exponent:               exponent,
		ReserveFactor:          sdk.MustNewDecFromStr("0.2"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.25"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.26"),
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
		HistoricMedians:        24,
	}
}
