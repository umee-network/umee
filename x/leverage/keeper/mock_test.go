package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v5/x/oracle/types"
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

// MintCoins adds coins to a module account
func (m *mockBankKeeper) MintCoins(
	_ sdk.Context, toModule string, coins sdk.Coins,
) error {
	moduleAddr := authtypes.NewModuleAddress(toModule)
	moduleBalance, ok := m.spendableCoins[moduleAddr.String()]
	if !ok {
		moduleBalance = sdk.NewCoins()
	}
	m.spendableCoins[moduleAddr.String()] = moduleBalance.Add(coins...)
	return nil
}

// BurnCoins removes coins from a module account
func (m *mockBankKeeper) BurnCoins(
	_ sdk.Context, fromModule string, coins sdk.Coins,
) error {
	moduleAddr := authtypes.NewModuleAddress(fromModule)
	moduleBalance, ok := m.spendableCoins[moduleAddr.String()]
	if !ok {
		moduleBalance = sdk.NewCoins()
	}
	if coins.IsAnyGT(moduleBalance) {
		return errors.New("mock bank: insufficient module balance to burn")
	}
	m.spendableCoins[moduleAddr.String()] = moduleBalance.Sub(coins...)
	return nil
}

// SendCoins sends coins from one account's spendable coins to another.
// Error on insufficient balance.
func (m *mockBankKeeper) SendCoins(
	_ sdk.Context, fromAddr, toAddr sdk.AccAddress, coins sdk.Coins,
) error {
	toBalance, ok := m.spendableCoins[toAddr.String()]
	if !ok {
		toBalance = sdk.NewCoins()
	}
	fromBalance, ok := m.spendableCoins[fromAddr.String()]
	if !ok {
		fromBalance = sdk.NewCoins()
	}
	if coins.IsAnyGT(fromBalance) {
		return errors.New("mock bank: insufficient from balance")
	}
	m.spendableCoins[fromAddr.String()] = fromBalance.Sub(coins...)
	m.spendableCoins[toAddr.String()] = toBalance.Add(coins...)
	return nil
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

// GetAllBalances returns an account's spendable coins
func (m *mockBankKeeper) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return m.SpendableCoins(ctx, addr)
}

// GetBalance returns an element from an account's spendable coins
func (m *mockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, m.SpendableCoins(ctx, addr).AmountOf(denom))
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

type mockOracleKeeper struct {
	baseExchangeRates     map[string]sdk.Dec
	symbolExchangeRates   map[string]sdk.Dec
	historicExchangeRates map[string]sdk.Dec
}

func newMockOracleKeeper() *mockOracleKeeper {
	m := &mockOracleKeeper{
		baseExchangeRates:     make(map[string]sdk.Dec),
		symbolExchangeRates:   make(map[string]sdk.Dec),
		historicExchangeRates: make(map[string]sdk.Dec),
	}
	m.Reset()

	return m
}

func (m *mockOracleKeeper) MedianOfHistoricMedians(ctx sdk.Context, denom string, numStamps uint64,
) (sdk.Dec, uint32, error) {
	p, ok := m.historicExchangeRates[denom]
	if !ok {
		// This error matches oracle behavior on zero historic medians
		return sdk.ZeroDec(), 0, types.ErrNoHistoricMedians.Wrapf(
			"requested %d, got %d",
			numStamps,
			0,
		)
	}

	return p, uint32(numStamps), nil
}

func (m *mockOracleKeeper) GetExchangeRate(_ sdk.Context, denom string) (sdk.Dec, error) {
	p, ok := m.symbolExchangeRates[denom]
	if !ok {
		// This error matches oracle behavior on missing asset price
		return sdk.ZeroDec(), oracletypes.ErrUnknownDenom.Wrap(denom)
	}

	return p, nil
}

// Clear clears a denom from the mock oracle, simulating an outage.
func (m *mockOracleKeeper) Clear(denom string) {
	delete(m.symbolExchangeRates, denom)
	delete(m.historicExchangeRates, denom)
}

// Reset restores the mock oracle's prices to its default values.
func (m *mockOracleKeeper) Reset() {
	m.symbolExchangeRates = map[string]sdk.Dec{
		"UMEE": sdk.MustNewDecFromStr("4.21"),
		"ATOM": sdk.MustNewDecFromStr("39.38"),
		"DAI":  sdk.MustNewDecFromStr("1.00"),
		"DUMP": sdk.MustNewDecFromStr("0.50"), // A token which has recently halved in price
		"PUMP": sdk.MustNewDecFromStr("2.00"), // A token which has recently doubled in price
	}
	m.historicExchangeRates = map[string]sdk.Dec{
		"UMEE": sdk.MustNewDecFromStr("4.21"),
		"ATOM": sdk.MustNewDecFromStr("39.38"),
		"DAI":  sdk.MustNewDecFromStr("1.00"),
		"DUMP": sdk.MustNewDecFromStr("1.00"),
		"PUMP": sdk.MustNewDecFromStr("1.00"),
	}
}
