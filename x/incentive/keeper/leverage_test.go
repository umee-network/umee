package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	leveragefixtures "github.com/umee-network/umee/v4/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// mockLeverageKeeper implements the methods called by the incentive module on the leverage module,
// but will not independently call any methods on the incentive module as the real leverage module does.
type mockLeverageKeeper struct {
	// collateral[address][uToken] = int64
	collateral map[string]map[string]int64
	// to test emergency unbondings
	donatedCollateral sdk.Coins
}

func newMockLeverageKeeper() *mockLeverageKeeper {
	m := &mockLeverageKeeper{
		collateral:        map[string]map[string]int64{},
		donatedCollateral: sdk.NewCoins(),
	}
	return m
}

// GetCollateral implements the expected leverage keeper
func (m *mockLeverageKeeper) GetCollateral(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	amount, ok := m.collateral[addr.String()][denom]
	if !ok {
		return coin.Zero(denom)
	}
	return sdk.NewCoin(denom, sdk.NewInt(amount))
}

// DonateCollateral implements the expected leverage keeper
func (m *mockLeverageKeeper) DonateCollateral(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error {
	return nil
}

// getDonatedCollateral is used to test the effects of emergency unbondings
func (m *mockLeverageKeeper) getDonatedCollateral(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, m.donatedCollateral.AmountOf(denom))
}

// setCollateral sets an account's collateral in the mock leverage keeper without requiring any supplying action
func (m *mockLeverageKeeper) setCollateral(addr sdk.AccAddress, denom string, amount int64) {
	if _, ok := m.collateral[addr.String()]; !ok {
		m.collateral[addr.String()] = map[string]int64{}
	}
	m.collateral[addr.String()][denom] = amount
}

// GetTokenSettings implements the expected leverage keeper, with UMEE, ATOM, and DAI registered.
func (m *mockLeverageKeeper) GetTokenSettings(ctx sdk.Context, denom string) (leveragetypes.Token, error) {
	switch denom {
	case umeeDenom:
		return leveragefixtures.Token(denom, "UMEE", 6), nil
	case leveragefixtures.AtomDenom:
		return leveragefixtures.Token(denom, "ATOM", 6), nil
	case leveragefixtures.DaiDenom:
		return leveragefixtures.Token(denom, "DAI", 18), nil
	}
	return leveragetypes.Token{}, leveragetypes.ErrNotRegisteredToken
}
