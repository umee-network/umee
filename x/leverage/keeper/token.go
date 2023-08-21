package keeper

import (
	"errors"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

// CleanTokenRegistry deletes all blacklisted tokens in the leverage registry
// whose uToken supplies are zero. Called automatically on registry update.
func (k Keeper) CleanTokenRegistry(ctx sdk.Context) error {
	tokens := k.GetAllRegisteredTokens(ctx)
	for _, t := range tokens {
		if t.Blacklist {
			uDenom := coin.ToUTokenDenom(t.BaseDenom)
			uSupply := k.GetUTokenSupply(ctx, uDenom)
			if uSupply.IsZero() {
				err := k.deleteTokenSettings(ctx, t)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// deleteTokenSettings deletes a Token in the x/leverage module's KVStore.
// it should only be called by CleanTokenRegistry.
func (k Keeper) deleteTokenSettings(ctx sdk.Context, token types.Token) error {
	store := ctx.KVStore(k.storeKey)
	tokenKey := types.KeyRegisteredToken(token.BaseDenom)
	store.Delete(tokenKey)
	// call token hooks on deleted (not just blacklisted) token
	k.afterRegisteredTokenRemoved(ctx, token)
	return nil
}

// SetTokenSettings stores a Token into the x/leverage module's KVStore.
func (k Keeper) SetTokenSettings(ctx sdk.Context, token types.Token) error {
	if err := token.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	tokenKey := types.KeyRegisteredToken(token.BaseDenom)

	bz, err := k.cdc.Marshal(&token)
	if err != nil {
		return err
	}

	k.afterTokenRegistered(ctx, token)
	store.Set(tokenKey, bz)
	return nil
}

// GetTokenSettings gets a token from the x/leverage module's KVStore.
func (k Keeper) GetTokenSettings(ctx sdk.Context, denom string) (types.Token, error) {
	store := ctx.KVStore(k.storeKey)
	tokenKey := types.KeyRegisteredToken(denom)

	token := types.Token{}
	bz := store.Get(tokenKey)
	if len(bz) == 0 {
		return token, types.ErrNotRegisteredToken.Wrap(denom)
	}

	err := k.cdc.Unmarshal(bz, &token)
	return token, err
}

// SetSpecialAssetPair stores a SpecialAssetPair into the x/leverage module's KVStore.
// Deletes any existing special pairs between the assets instead if given zero
// collateral weight and zero liquidation threshold.
func (k Keeper) SetSpecialAssetPair(ctx sdk.Context, pair types.SpecialAssetPair) error {
	if err := pair.Validate(); err != nil {
		return err
	}
	if !pair.CollateralWeight.IsPositive() && !pair.LiquidationThreshold.IsPositive() {
		k.deleteSpecialAssetPair(ctx, pair.Collateral, pair.Borrow)
		return nil
	}

	key := types.KeySpecialAssetPair(pair.Collateral, pair.Borrow)
	return store.SetValue(ctx.KVStore(k.storeKey), key, &pair, "leverage-special-asset")
}

// deleteSpecialAssetPair removes a SpecialAssetPair from the x/leverage module's KVStore.
func (k Keeper) deleteSpecialAssetPair(ctx sdk.Context, collateralDenom, borrowDenom string) {
	key := types.KeySpecialAssetPair(collateralDenom, borrowDenom)
	ctx.KVStore(k.storeKey).Delete(key)
}

// UpdateTokenRegistry adds new tokens or updates the new tokens settings to registry.
// It requires maps of the currently registered base and symbol denoms, so it can prevent duplicates of either.
func (k Keeper) UpdateTokenRegistry(
	ctx sdk.Context, toUpdate, toAdd []types.Token,
	regDenoms map[string]types.Token, byEmergencyGroup bool,
) error {
	// NOTE: validation is skipped here because it's done in MsgGovUpdateRegistry.ValidateBasic()
	// and in k.SetTokenSettings

	errs := assertTokensRegistered(toUpdate, regDenoms)

	if byEmergencyGroup {
		if len(toAdd) != 0 {
			errs = append(errs, sdkerrors.ErrInvalidRequest.Wrap("Emergency Group can't register new tokens"))
		}
		if errs2 := validateEmergencyTokenSettingsUpdate(regDenoms, toUpdate); errs2 != nil {
			errs = append(errs, errs2...)
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	for _, token := range toUpdate {
		if err := k.SetTokenSettings(ctx, token); err != nil {
			return err
		}
	}

	regSymbols := map[string]bool{}
	for i := range regDenoms {
		regSymbols[strings.ToUpper(regDenoms[i].SymbolDenom)] = true
	}
	for _, token := range toAdd {
		// Note: we are allowing duplicate symbols (Kava USDT, axelar USDT both have same USDT symbol )
		if _, ok := regDenoms[token.BaseDenom]; ok {
			return types.ErrDuplicateToken.Wrapf("token %s is already registered", token.BaseDenom)
		}
		if err := k.SetTokenSettings(ctx, token); err != nil {
			return err
		}
	}

	return nil
}

var maxEmergencyActionNumericDiff = sdk.MustNewDecFromStr("0.2")

func validateEmergencyTokenSettingsUpdate(regTokens map[string]types.Token, updates []types.Token) []error {
	var errs []error
	for _, ut := range updates {
		t := regTokens[ut.BaseDenom]
		if !t.ReserveFactor.Equal(ut.ReserveFactor) {
			errs = append(errs, errors.New("can't change ReserveFactor"))
		}
		if !t.CollateralWeight.Equal(ut.CollateralWeight) {
			errs = append(errs, errors.New("can't change CollateralWeight"))
		}
		if !t.LiquidationThreshold.Equal(ut.LiquidationThreshold) {
			errs = append(errs, errors.New("can't change LiquidationThreshold"))
		}
		if !t.BaseBorrowRate.Equal(ut.BaseBorrowRate) {
			errs = append(errs, errors.New("can't change BaseBorrowRate"))
		}
		if !t.KinkBorrowRate.Equal(ut.KinkBorrowRate) {
			errs = append(errs, errors.New("can't change KinkBorrowRate"))
		}
		if !t.MaxBorrowRate.Equal(ut.MaxBorrowRate) {
			errs = append(errs, errors.New("can't change MaxBorrowRate"))
		}
		if !t.KinkUtilization.Equal(ut.KinkUtilization) {
			errs = append(errs, errors.New("can't change KinkUtilization"))
		}
		if !t.LiquidationIncentive.Equal(ut.LiquidationIncentive) {
			errs = append(errs, errors.New("can't change LiquidationIncentive"))
		}
		if t.SymbolDenom != ut.SymbolDenom {
			errs = append(errs, errors.New("can't change SymbolDenom"))
		}
		if t.Exponent != ut.Exponent {
			errs = append(errs, errors.New("can't change Exponent"))
		}
		if t.Blacklist != ut.Blacklist {
			errs = append(errs, errors.New("can't change Blacklist"))
		}
		if t.HistoricMedians != ut.HistoricMedians {
			errs = append(errs, errors.New("can't change HistoricMedians"))
		}

		// EnableMsgSupply, EnableMsgBorrow
		// we only allow switch to disable
		if !t.EnableMsgSupply && ut.EnableMsgSupply {
			errs = append(errs, errors.New("can't switch EnableMsgSupply to true"))
		}
		if !t.EnableMsgBorrow && ut.EnableMsgBorrow {
			errs = append(errs, errors.New("can't switch EnableMsgBorrow to true"))
		}

		// MaxCollateralShare
		// allow limited numeric change
		err := checkers.DecMaxDiff(
			t.MaxCollateralShare, ut.MaxCollateralShare, maxEmergencyActionNumericDiff, "MaxCollateralShare")
		if err != nil {
			errs = append(errs, err)
		}

		// MaxCollateralShare, MaxSupplyUtilization, MinCollateralLiquidity, MaxSupply
		// allow any change
	}

	return errs
}

func assertTokensRegistered(tokens []types.Token, regTokens map[string]types.Token) []error {
	errs := []error{}
	for i := range tokens {
		d := tokens[i].BaseDenom
		if _, ok := regTokens[d]; !ok {
			errs = append(errs, types.ErrNotRegisteredToken.Wrap(d))
		}
	}
	return errs
}
