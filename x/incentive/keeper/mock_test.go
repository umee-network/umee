package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/umee-network/umee/v4/util/coin"
	leveragefixtures "github.com/umee-network/umee/v4/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// mockBankKeeper mocks the bank keeper
type mockBankKeeper struct {
	spendableCoins map[string]sdk.Coins
}

func newMockBankKeeper() mockBankKeeper {
	m := mockBankKeeper{
		spendableCoins: map[string]sdk.Coins{},
	}
	return m
}

// SendCoinsFromModuleToAccount sends coins from a module balance to an account's spendable coins.
// Error on insufficient module balance.
func (m *mockBankKeeper) SendCoinsFromModuleToAccount(
	_ sdk.Context, fromModule string, toAddr sdk.AccAddress, coins sdk.Coins,
) error {
	moduleAddr := authtypes.NewModuleAddress(fromModule)
	spendable, ok := m.spendableCoins[toAddr.String()]
	if !ok {
		spendable = sdk.NewCoins()
	}
	moduleBalance, ok := m.spendableCoins[moduleAddr.String()]
	if !ok {
		moduleBalance = sdk.NewCoins()
	}
	if coins.IsAnyGT(moduleBalance) {
		return errors.New("mock bank: insufficient module balance")
	}
	m.spendableCoins[moduleAddr.String()] = moduleBalance.Sub(coins...)
	m.spendableCoins[toAddr.String()] = spendable.Add(coins...)
	return nil
}

// SendCoinsFromAccountToModule sends coins from an account's spendable balance to a module balance.
// Error on insufficient spendable coins.
func (m *mockBankKeeper) SendCoinsFromAccountToModule(
	_ sdk.Context, fromAddr sdk.AccAddress, toModule string, coins sdk.Coins,
) error {
	moduleAddr := authtypes.NewModuleAddress(toModule)
	spendable, ok := m.spendableCoins[fromAddr.String()]
	if !ok {
		spendable = sdk.NewCoins()
	}
	moduleBalance, ok := m.spendableCoins[moduleAddr.String()]
	if !ok {
		moduleBalance = sdk.NewCoins()
	}
	if coins.IsAnyGT(spendable) {
		return errors.New("mock bank: insufficient account balance")
	}
	m.spendableCoins[fromAddr.String()] = spendable.Sub(coins...)
	m.spendableCoins[moduleAddr.String()] = moduleBalance.Add(coins...)
	return nil
}

// SendCoinsFromModuleToModule sends coins from one module balance to another.
// Error on insufficient module balance.
func (m *mockBankKeeper) SendCoinsFromModuleToModule(_ sdk.Context, fromModule, toModule string, coins sdk.Coins) error {
	fromAddr := authtypes.NewModuleAddress(fromModule)
	fromBalance, ok := m.spendableCoins[fromAddr.String()]
	if !ok {
		fromBalance = sdk.NewCoins()
	}
	toAddr := authtypes.NewModuleAddress(toModule)
	toBalance, ok := m.spendableCoins[toAddr.String()]
	if !ok {
		toBalance = sdk.NewCoins()
	}
	if coins.IsAnyGT(fromBalance) {
		return errors.New("mock bank: insufficient module balance")
	}
	m.spendableCoins[fromAddr.String()] = fromBalance.Sub(coins...)
	m.spendableCoins[toAddr.String()] = toBalance.Add(coins...)
	return nil
}

// SpendableCoins returns an account's spendable coins, without validating the address
func (m *mockBankKeeper) SpendableCoins(_ sdk.Context, addr sdk.AccAddress) sdk.Coins {
	spendable, ok := m.spendableCoins[addr.String()]
	if !ok {
		return sdk.NewCoins()
	}
	return spendable
}

// FundAccount mints new coins and sends them to an address.
func (m *mockBankKeeper) FundAccount(addr sdk.AccAddress, coins sdk.Coins) {
	coins = sdk.NewCoins(coins...) // prevents panic: Wrong argument: coins must be sorted
	spendable, ok := m.spendableCoins[addr.String()]
	if !ok {
		spendable = sdk.NewCoins()
	}
	m.spendableCoins[addr.String()] = spendable.Add(coins...)
}

// FundModule mints new coins and adds them to a module balance.
func (m *mockBankKeeper) FundModule(module string, coins sdk.Coins) {
	coins = sdk.NewCoins(coins...) // prevents panic: Wrong argument: coins must be sorted
	moduleAddr := authtypes.NewModuleAddress(module)
	balance, ok := m.spendableCoins[moduleAddr.String()]
	if !ok {
		balance = sdk.NewCoins()
	}
	m.spendableCoins[moduleAddr.String()] = balance.Add(coins...)
}

// mockLeverageKeeper implements the methods called by the incentive module on the leverage module,
// but will not independently call any methods on the incentive module as the real leverage module does.
type mockLeverageKeeper struct {
	// collateral[address] = coins
	collateral map[string]sdk.Coins
	// to test emergency unbondings
	donatedCollateral sdk.Coins
}

func newMockLeverageKeeper() mockLeverageKeeper {
	m := mockLeverageKeeper{
		collateral:        map[string]sdk.Coins{},
		donatedCollateral: sdk.NewCoins(),
	}
	return m
}

// GetCollateral implements the expected leverage keeper
func (m *mockLeverageKeeper) GetCollateral(_ sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	collateral, ok := m.collateral[addr.String()]
	if !ok {
		return coin.Zero(denom)
	}
	return sdk.NewCoin(denom, collateral.AmountOf(denom))
}

// DonateCollateral implements the expected leverage keeper
func (m *mockLeverageKeeper) DonateCollateral(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error {
	newCollateral := m.GetCollateral(ctx, addr, uToken.Denom).Sub(uToken).Amount.Int64()
	m.setCollateral(addr, uToken.Denom, newCollateral)
	m.donatedCollateral = m.donatedCollateral.Add(uToken)
	return nil
}

// getDonatedCollateral is used to test the effects of emergency unbondings
func (m *mockLeverageKeeper) getDonatedCollateral(_ sdk.Context, denom string) sdk.Coin {
	return sdk.NewCoin(denom, m.donatedCollateral.AmountOf(denom))
}

// setCollateral sets an account's collateral in the mock leverage keeper without requiring any supplying action
func (m *mockLeverageKeeper) setCollateral(addr sdk.AccAddress, denom string, amount int64) {
	collateral, ok := m.collateral[addr.String()]
	if !ok {
		// initialize
		collateral = sdk.NewCoins()
	}
	if collateral.AmountOf(denom).IsZero() {
		// no existing collateral of this denom
		collateral = collateral.Add(sdk.NewInt64Coin(denom, amount))
	} else {
		// overwrite existing collateral of this denom
		for i := range collateral {
			if collateral[i].Denom == denom {
				collateral[i].Amount = sdk.NewInt(amount)
			}
		}
	}
	// set collateral
	m.collateral[addr.String()] = collateral
}

// GetTokenSettings implements the expected leverage keeper, with UMEE, ATOM, and DAI registered.
func (m *mockLeverageKeeper) GetTokenSettings(_ sdk.Context, denom string) (leveragetypes.Token, error) {
	switch denom {
	case leveragefixtures.UmeeDenom:
		return leveragefixtures.Token(denom, "UMEE", 6), nil
	case leveragefixtures.AtomDenom:
		return leveragefixtures.Token(denom, "ATOM", 6), nil
	case leveragefixtures.DaiDenom:
		return leveragefixtures.Token(denom, "DAI", 18), nil
	}
	return leveragetypes.Token{}, leveragetypes.ErrNotRegisteredToken
}

// TotalTokenValue implements the expected leverage keeper, with UMEE, ATOM, and DAI registered.
func (m *mockLeverageKeeper) TotalTokenValue(_ sdk.Context, coins sdk.Coins, _ leveragetypes.PriceMode) (sdk.Dec, error) {
	total := sdk.ZeroDec()

	var (
		umeePrice = sdk.MustNewDecFromStr("4.21")
		atomPrice = sdk.MustNewDecFromStr("39.38")
		daiPrice  = sdk.MustNewDecFromStr("1.00")
	)

	for _, coin := range coins {
		switch coin.Denom {
		case leveragefixtures.UmeeDenom:
			total = total.Add(
				umeePrice.MulInt(coin.Amount).QuoInt64(1_000000))
		case leveragefixtures.AtomDenom:
			total = total.Add(
				atomPrice.MulInt(coin.Amount).QuoInt64(1_000000))
		case leveragefixtures.DaiDenom:
			total = total.Add(
				daiPrice.MulInt(coin.Amount).QuoInt64(1_000000_000000_000000))
		}
	}
	return total, nil
}

// ExchangeUToken implements the expected leverage keeper, with uToken exchange rates always equal to 1
func (m *mockLeverageKeeper) ExchangeUToken(_ sdk.Context, uToken sdk.Coin) (sdk.Coin, error) {
	if !leveragetypes.HasUTokenPrefix(uToken.Denom) {
		return uToken, leveragetypes.ErrUToken
	}
	return sdk.NewCoin(
		leveragetypes.ToTokenDenom(uToken.Denom),
		uToken.Amount,
	), nil
}
