package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
)

type mockLeverageKeeper struct{}

func newMockLeverageKeeper() *mockLeverageKeeper {
	m := &mockLeverageKeeper{}
	return m
}

func (m *mockLeverageKeeper) GetCollateral(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	// TODO
	return coin.Zero(denom)
}
