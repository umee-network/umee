package fixtures

import (
	sdkmath "cosmossdk.io/math"

	"github.com/umee-network/umee/v6/x/leverage/types"
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
		ReserveFactor:          sdkmath.LegacyMustNewDecFromStr("0.2"),
		CollateralWeight:       sdkmath.LegacyMustNewDecFromStr("0.25"),
		LiquidationThreshold:   sdkmath.LegacyMustNewDecFromStr("0.26"),
		BaseBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.22"),
		MaxBorrowRate:          sdkmath.LegacyMustNewDecFromStr("1.52"),
		KinkUtilization:        sdkmath.LegacyMustNewDecFromStr("0.8"),
		LiquidationIncentive:   sdkmath.LegacyMustNewDecFromStr("0.1"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdkmath.LegacyMustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdkmath.LegacyMustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdkmath.LegacyMustNewDecFromStr("0"),
		MaxSupply:              sdkmath.NewInt(100_000_000000),
		HistoricMedians:        24,
	}
}
