package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/x/leverage/types"
)

const (
	routeExchangeRates    = "exchange-rates"
	routeReserveAmount    = "reserve-amount"
	routeCollateralAmount = "collateral-amount"
	routeBorrowAmount     = "borrow-amount"
	routeBorrowAPY        = "borrow-apy"
)

// RegisterInvariants registers the leverage module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, routeExchangeRates, ExchangeRatesInvariant(k))
	ir.RegisterRoute(types.ModuleName, routeReserveAmount, ReserveAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, routeCollateralAmount, CollateralAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, routeBorrowAmount, BorrowAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, routeBorrowAPY, BorrowAPYInvariant(k))
}

// AllInvariants runs all invariants of the x/leverage module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := ExchangeRatesInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		res, stop = ReserveAmountInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		res, stop = CollateralAmountInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		res, stop = BorrowAmountInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		return BorrowAPYInvariant(k)(ctx)
	}
}

// ExchangeRatesInvariant checks that all exchante rate denom are bigger or equal to 1
func ExchangeRatesInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		exchangeRatePrefix := types.CreateExchangeRateKeyNoDenom()

		// Iterate through all denoms which have an exchange rate stored
		// in the keeper. If a token is registered but its exchange rate is
		// lower than 1.0 or it has some error doing the unmarshal it
		// adds the denom invariant count and message description
		err := k.Iterate(ctx, exchangeRatePrefix, func(key, val []byte) error {
			// remove exchangeRatePrefix and null-terminator
			denom := types.DenomFromKey(key, exchangeRatePrefix)

			amount := sdk.ZeroDec()
			if err := amount.Unmarshal(val); err != nil {
				count++
				msg += fmt.Sprintf("\t%s received an error while Unmarshal the byte %+v\n", denom, val)
				return nil
			}

			if amount.LT(sdk.OneDec()) {
				count++
				msg += fmt.Sprintf("\t%s exchange rate %s is lower than one\n", denom, amount.String())
			}
			return nil
		})

		if err != nil {
			msg += fmt.Sprintf("\tSome error occurred while iterating through the exchange rates %+v\n", err)
		}

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, routeExchangeRates,
			fmt.Sprintf("amount of exchange rate lower than one %d\n%s", count, msg),
		), broken
	}
}

// ReserveAmountInvariant checks that reserve amounts have non-negative balances
func ReserveAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		reserveAmountPrefix := types.CreateReserveAmountKeyNoDenom()

		// Iterate through all denoms which have an reserve amount stored
		// in the keeper. If a token is registered but its reserve amount is
		// negative or it has some error doing the unmarshal it
		// adds the denom invariant count and message description
		err := k.Iterate(ctx, reserveAmountPrefix, func(key, val []byte) error {
			// remove reserveAmountPrefix and null-terminator
			denom := types.DenomFromKey(key, reserveAmountPrefix)

			amount := sdk.ZeroInt()
			if err := amount.Unmarshal(val); err != nil {
				count++
				msg += fmt.Sprintf("\t%s received an error while Unmarshal the byte %+v\n", denom, val)
				return nil
			}

			if amount.IsNegative() {
				count++
				msg += fmt.Sprintf("\t%s reserve amount %s is negative\n", denom, amount.String())
			}
			return nil
		})

		if err != nil {
			msg += fmt.Sprintf("\tSome error occurred while iterating through the reserve amount %+v\n", err)
		}

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, routeReserveAmount,
			fmt.Sprintf("number of negative reserve amount found %d\n%s", count, msg),
		), broken
	}
}

// CollateralAmountInvariant checks that collateral amounts have all positive values
func CollateralAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		collateralPrefix := types.CreateCollateralAmountKeyNoAddress()

		// Iterate through all denoms which have an borrow amount stored
		// in the keeper. If a token is registered but its borrow amount is
		// not positive or it has some error doing the unmarshal it
		// adds the denom and address invariant count and message description
		err := k.Iterate(ctx, collateralPrefix, func(key, val []byte) error {
			// remove prefix | lengthPrefixed(addr) and null-terminator
			denom := types.DenomFromKeyWithAddress(key, collateralPrefix)
			// remove prefix | denom and null-terminator
			address := types.AddressFromKey(key, collateralPrefix)

			amount := sdk.ZeroInt()
			if err := amount.Unmarshal(val); err != nil {
				count++
				msg += fmt.Sprintf("\t%s - %s received an error while Unmarshal the byte %+v\n", denom, address.String(), val)
				return nil
			}

			if !amount.IsPositive() {
				count++
				msg += fmt.Sprintf("\t%s - %s collateral amount %s is not positive\n", denom, address.String(), amount.String())
			}
			return nil
		})

		if err != nil {
			msg += fmt.Sprintf("\tSome error occurred while iterating through the collateral amount %+v\n", err)
		}

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, routeCollateralAmount,
			fmt.Sprintf("number of not positive collateral amount found %d\n%s", count, msg),
		), broken
	}
}

// BorrowAmountInvariant checks that borrow amounts have all positive values
func BorrowAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		loanKeyPrefix := types.CreateLoanKeyNoAddress()

		// Iterate through all denoms which have an borrow amount stored
		// in the keeper. If a token is registered but its borrow amount is
		// not positive or it has some error doing the unmarshal it
		// adds the denom and address invariant count and message description
		err := k.Iterate(ctx, loanKeyPrefix, func(key, val []byte) error {
			// remove prefix | lengthPrefixed(addr) and null-terminator
			denom := types.DenomFromKeyWithAddress(key, loanKeyPrefix)
			// remove prefix | denom and null-terminator
			address := types.AddressFromKey(key, loanKeyPrefix)

			amount := sdk.ZeroInt()
			if err := amount.Unmarshal(val); err != nil {
				count++
				msg += fmt.Sprintf("\t%s - %s received an error while Unmarshal the byte %+v\n", denom, address.String(), val)
				return nil
			}

			if !amount.IsPositive() {
				count++
				msg += fmt.Sprintf("\t%s - %s borrow amount %s is not positive\n", denom, address.String(), amount.String())
			}
			return nil
		})

		if err != nil {
			msg += fmt.Sprintf("\tSome error occurred while iterating through the borrow amount %+v\n", err)
		}

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, routeBorrowAmount,
			fmt.Sprintf("number of not positive borrow amount found %d\n%s", count, msg),
		), broken
	}
}

// BorrowAPYInvariant checks that Borrow APY have all positive values
func BorrowAPYInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		borrowAPYprefix := types.CreateBorrowAPYKeyNoDenom()

		// Iterate through all denoms which have an Borrow APY stored
		// in the keeper. If a token is registered but its borrow APY is
		// not positive or it has some error doing the unmarshal it
		// adds the denom and address invariant count and message description
		err := k.Iterate(ctx, borrowAPYprefix, func(key, val []byte) error {
			denom := types.DenomFromKey(key, borrowAPYprefix)

			var borrowAPY sdk.Dec
			if err := borrowAPY.Unmarshal(val); err != nil {
				count++
				msg += fmt.Sprintf("\t%s received an error while Unmarshal the byte %+v\n", denom, val)
				return nil
			}

			if !borrowAPY.IsPositive() {
				count++
				msg += fmt.Sprintf("\t%s borrow APY %s is not positive\n", denom, borrowAPY.String())
			}
			return nil
		})

		if err != nil {
			msg += fmt.Sprintf("\tSome error occurred while iterating through the borrow APY %+v\n", err)
		}

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, routeBorrowAPY,
			fmt.Sprintf("number of not positive borrow APY found %d\n%s", count, msg),
		), broken
	}
}
