package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

const (
	routeInterestScalars  = "interest-scalars"
	routeReserveAmount    = "reserve-amount"
	routeCollateralAmount = "collateral-amount"
	routeBorrowAmount     = "borrow-amount"
	routeBorrowAPY        = "borrow-apy"
	routeSupplyAPY        = "supply-apy"
)

// RegisterInvariants registers the leverage module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, routeReserveAmount, ReserveAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, routeCollateralAmount, CollateralAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, routeBorrowAmount, BorrowAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, routeBorrowAPY, BorrowAPYInvariant(k))
	ir.RegisterRoute(types.ModuleName, routeSupplyAPY, SupplyAPYInvariant(k))
	ir.RegisterRoute(types.ModuleName, routeInterestScalars, InterestScalarsInvariant(k))
}

// AllInvariants runs all invariants of the x/leverage module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := ReserveAmountInvariant(k)(ctx)
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

		res, stop = BorrowAPYInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		res, stop = SupplyAPYInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		return InterestScalarsInvariant(k)(ctx)
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
		err := k.iterate(ctx, reserveAmountPrefix, func(key, val []byte) error {
			// remove reserveAmountPrefix and null-terminator
			denom := types.DenomFromKey(key, reserveAmountPrefix)

			amount := sdk.ZeroInt()
			if err := amount.Unmarshal(val); err != nil {
				count++
				msg += fmt.Sprintf("\tfailed to unmarshal bytes for %s: %+v\n", denom, val)
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

		// Iterate through all collateral amounts stored in the keeper,
		// ensuring all successfully unmarshal to positive values.
		err := k.iterate(ctx, collateralPrefix, func(key, val []byte) error {
			// remove prefix | lengthPrefixed(addr) and null-terminator
			denom := types.DenomFromKeyWithAddress(key, collateralPrefix)
			// remove prefix | denom and null-terminator
			address := types.AddressFromKey(key, collateralPrefix)

			amount := sdk.ZeroInt()
			if err := amount.Unmarshal(val); err != nil {
				count++
				msg += fmt.Sprintf("\tfailed to unmarshal bytes for %s - %s: %+v\n", denom, address.String(), val)
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

		borrowPrefix := types.KeyPrefixAdjustedBorrow

		// Iterate through all adjusted borrow amounts stored in the keeper,
		// ensuring all successfully unmarshal to positive values.
		err := k.iterate(ctx, borrowPrefix, func(key, val []byte) error {
			// remove prefix | lengthPrefixed(addr) and null-terminator
			denom := types.DenomFromKeyWithAddress(key, borrowPrefix)
			// remove prefix | denom and null-terminator
			address := types.AddressFromKey(key, borrowPrefix)

			amount := sdk.ZeroDec()
			if err := amount.Unmarshal(val); err != nil {
				count++
				msg += fmt.Sprintf("\tfailed to unmarshal bytes for %s - %s: %+v\n", denom, address.String(), val)
				return nil
			}

			if !amount.IsPositive() {
				count++
				msg += fmt.Sprintf("\t%s - %s adjusted borrow %s is not positive\n", denom, address.String(), amount.String())
			}
			return nil
		})
		if err != nil {
			msg += fmt.Sprintf("\tSome error occurred while iterating through adjusted borrow amounts %+v\n", err)
		}

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, routeBorrowAmount,
			fmt.Sprintf("number of not positive adjusted borrow amounts found %d\n%s", count, msg),
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

		tokenPrefix := types.KeyPrefixRegisteredToken

		// Iterate through all denoms of registered tokens in the
		// keeper, ensuring none have a negative borrow APY.
		err := k.iterate(ctx, tokenPrefix, func(key, _ []byte) error {
			denom := types.DenomFromKey(key, tokenPrefix)

			borrowAPY := k.DeriveBorrowAPY(ctx, denom)

			if borrowAPY.IsNegative() {
				count++
				msg += fmt.Sprintf("\t%s borrow APY %s is negative\n", denom, borrowAPY.String())
			}
			return nil
		})
		if err != nil {
			msg += fmt.Sprintf("\tSome error occurred while iterating through the borrow APY %+v\n", err)
		}

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, routeBorrowAPY,
			fmt.Sprintf("number of negative borrow APY found %d\n%s", count, msg),
		), broken
	}
}

// SupplyAPYInvariant checks that Supply APY have all positive values
func SupplyAPYInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		tokenPrefix := types.KeyPrefixRegisteredToken

		// Iterate through all denoms of registered tokens in the
		// keeper, ensuring none have a negative supply APY.
		err := k.iterate(ctx, tokenPrefix, func(key, _ []byte) error {
			denom := types.DenomFromKey(key, tokenPrefix)

			supplyAPY := k.DeriveSupplyAPY(ctx, denom)

			if supplyAPY.IsNegative() {
				count++
				msg += fmt.Sprintf("\t%s supply APY %s is negative\n", denom, supplyAPY.String())
			}
			return nil
		})
		if err != nil {
			msg += fmt.Sprintf("\tSome error occurred while iterating through the supply APY %+v\n", err)
		}

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, routeSupplyAPY,
			fmt.Sprintf("number of negative supply APY found %d\n%s", count, msg),
		), broken
	}
}

// InterestScalarsInvariant checks that all denoms have an interest scalar >= 1
func InterestScalarsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		tokenPrefix := types.KeyPrefixRegisteredToken

		// Iterate through all denoms of registered tokens in the
		// keeper, ensuring none have an interest scalar less than one.
		err := k.iterate(ctx, tokenPrefix, func(key, _ []byte) error {
			denom := types.DenomFromKey(key, tokenPrefix)

			scalar := k.getInterestScalar(ctx, denom)

			if scalar.LT(sdk.OneDec()) {
				count++
				msg += fmt.Sprintf("\t%s interest scalar %s is less than one\n", denom, scalar.String())
			}
			return nil
		})
		if err != nil {
			msg += fmt.Sprintf("\tSome error occurred while iterating through the interest scalars %+v\n", err)
		}

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, routeInterestScalars,
			fmt.Sprintf("amount of interest scalars lower than one %d\n%s", count, msg),
		), broken
	}
}
