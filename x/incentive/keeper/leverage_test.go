package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// mockLeverageKeeper implements the methods called by the incentive module on the leverage module,
// but will not independently call any methods on the incentive module as the real leverage module does.
type mockLeverageKeeper struct {
	// collateral[address][uToken] = int
	collateral map[string]map[string]int64
}

func newMockLeverageKeeper() *mockLeverageKeeper {
	m := &mockLeverageKeeper{}
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

// setCollateral sets an account's collateral in the mock leverage keeper without requiring any supplying action
func (m *mockLeverageKeeper) setCollateral(addr sdk.AccAddress, denom string, amount int64) {
	m.collateral[addr.String()][denom] = amount
}

// GetTokenSettings implements the expected leverage keeper
func (m *mockLeverageKeeper) GetTokenSettings(ctx sdk.Context, denom string) (leveragetypes.Token, error) {
	return leveragetypes.Token{}, incentive.ErrNotImplemented
}
