package tests

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/x/leverage/client/cli"
	"github.com/umee-network/umee/v4/x/leverage/fixtures"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

func (s *IntegrationTestSuite) TestInvalidQueries() {
	invalidQueries := []testQuery{
		{
			"query market summary - denom not registered",
			cli.GetCmdQueryMarketSummary(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		{
			"query account balances - invalid address",
			cli.GetCmdQueryAccountBalances(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query account summary - invalid address",
			cli.GetCmdQueryAccountSummary(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query max withdraw - invalid address",
			cli.GetCmdQueryMaxWithdraw(),
			[]string{
				"xyz",
				"uumee",
			},
			true,
			nil,
			nil,
		},
		{
			"query registered token - denom not registered",
			cli.GetCmdQueryRegisteredTokens(),
			[]string{"umm"},
			true,
			nil,
			nil,
		},
	}

	// These queries do not require any borrower setup because they contain invalid arguments
	s.runTestQueries(invalidQueries...)
}

func (s *IntegrationTestSuite) TestLeverageScenario() {
	val := s.network.Validators[0]

	oracleSymbolPrice := sdk.MustNewDecFromStr("34.21")

	initialQueries := []testQuery{
		{
			"query params",
			cli.GetCmdQueryParams(),
			[]string{},
			false,
			&types.QueryParamsResponse{},
			&types.QueryParamsResponse{
				Params: types.DefaultParams(),
			},
		},
		{
			"query registered tokens",
			cli.GetCmdQueryRegisteredTokens(),
			[]string{},
			false,
			&types.QueryRegisteredTokensResponse{},
			&types.QueryRegisteredTokensResponse{
				Registry: []types.Token{
					fixtures.Token(appparams.BondDenom, appparams.DisplayDenom, 6),
				},
			},
		},
		{
			"query registered token info by base_denom",
			cli.GetCmdQueryRegisteredTokens(),
			[]string{appparams.BondDenom},
			false,
			&types.QueryRegisteredTokensResponse{},
			&types.QueryRegisteredTokensResponse{
				Registry: []types.Token{
					fixtures.Token(appparams.BondDenom, appparams.DisplayDenom, 6),
				},
			},
		},
		{
			"query market summary - zero supply",
			cli.GetCmdQueryMarketSummary(),
			[]string{
				appparams.BondDenom,
			},
			false,
			&types.QueryMarketSummaryResponse{},
			&types.QueryMarketSummaryResponse{
				SymbolDenom:         "UMEE",
				Exponent:            6,
				OraclePrice:         &oracleSymbolPrice,
				OracleHistoricPrice: &oracleSymbolPrice,
				UTokenExchangeRate:  sdk.OneDec(),
				// Borrow rate * (1.52 - ReserveFactor - OracleRewardFactor)
				// 1.52 * (1 - 0.2 - 0.01) = 1.2008
				Supply_APY: sdk.MustNewDecFromStr("1.2008"),
				// This is an edge case technically - when effective supply, meaning
				// module balance + total borrows, is zero, utilization (0/0) is
				// interpreted as 100% so max borrow rate (152% APY) is used.
				Borrow_APY:             sdk.MustNewDecFromStr("1.52"),
				Supplied:               sdk.ZeroInt(),
				Reserved:               sdk.ZeroInt(),
				Collateral:             sdk.ZeroInt(),
				Borrowed:               sdk.ZeroInt(),
				Liquidity:              sdk.ZeroInt(),
				MaximumBorrow:          sdk.ZeroInt(),
				MaximumCollateral:      sdk.ZeroInt(),
				MinimumLiquidity:       sdk.ZeroInt(),
				UTokenSupply:           sdk.ZeroInt(),
				AvailableBorrow:        sdk.ZeroInt(),
				AvailableWithdraw:      sdk.ZeroInt(),
				AvailableCollateralize: sdk.ZeroInt(),
			},
		},
		{
			"query bad debts",
			cli.GetCmdQueryBadDebts(),
			[]string{},
			false,
			&types.QueryBadDebtsResponse{},
			&types.QueryBadDebtsResponse{
				Targets: []types.BadDebt{},
			},
		},
		{
			"query max withdraw (zero)",
			cli.GetCmdQueryMaxWithdraw(),
			[]string{
				val.Address.String(),
				"uumee",
			},
			false,
			&types.QueryMaxWithdrawResponse{},
			&types.QueryMaxWithdrawResponse{
				Tokens:  sdk.NewCoins(),
				UTokens: sdk.NewCoins(),
			},
		},
		{
			"query max borrow (zero)",
			cli.GetCmdQueryMaxBorrow(),
			[]string{
				val.Address.String(),
				"uumee",
			},
			false,
			&types.QueryMaxBorrowResponse{},
			&types.QueryMaxBorrowResponse{
				Tokens: sdk.NewCoins(),
			},
		},
	}

	supply := testTransaction{
		"supply",
		cli.GetCmdSupply(),
		[]string{
			"700uumee",
		},
		nil,
	}

	addCollateral := testTransaction{
		"add collateral",
		cli.GetCmdCollateralize(),
		[]string{
			"700u/uumee",
		},
		nil,
	}

	supplyCollateral := testTransaction{
		"supply collateral",
		cli.GetCmdSupplyCollateral(),
		[]string{
			"300uumee",
		},
		nil,
	}

	borrow := testTransaction{
		"borrow",
		cli.GetCmdBorrow(),
		[]string{
			"150uumee",
		},
		nil,
	}

	maxborrow := testTransaction{
		"max-borrow",
		cli.GetCmdMaxBorrow(),
		[]string{
			"uumee", // should borrow up to the max of 250 uumee, which will become 251 due to rounding
		},
		nil,
	}

	liquidate := testTransaction{
		"liquidate",
		cli.GetCmdLiquidate(),
		[]string{
			val.Address.String(),
			"5uumee", // borrower liquidates itself, reduces borrow amount and collateral by 5
			"uumee",
		},
		nil,
	}

	repay := testTransaction{
		"repay",
		cli.GetCmdRepay(),
		[]string{
			"250uumee", // repays only the remaining borrowed balance, reduced automatically from 250
		},
		nil,
	}

	removeCollateral := testTransaction{
		"remove collateral",
		cli.GetCmdDecollateralize(),
		[]string{
			"895u/uumee", // 100 u/uumee will remain
		},
		nil,
	}

	withdraw := testTransaction{
		"withdraw",
		cli.GetCmdWithdraw(),
		[]string{
			"795u/uumee", // 200 u/uumee will remain
		},
		nil,
	}

	withdrawMax := testTransaction{
		"withdraw max",
		cli.GetCmdMaxWithdraw(),
		[]string{
			"uumee",
		},
		nil,
	}

	lt1 := sdk.MustNewDecFromStr("0.0085610525")

	nonzeroQueries := []testQuery{
		{
			"query account balances",
			cli.GetCmdQueryAccountBalances(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryAccountBalancesResponse{},
			&types.QueryAccountBalancesResponse{
				Supplied: sdk.NewCoins(
					sdk.NewInt64Coin(appparams.BondDenom, 1001),
				),
				Collateral: sdk.NewCoins(
					sdk.NewInt64Coin(types.ToUTokenDenom(appparams.BondDenom), 1000),
				),
				Borrowed: sdk.NewCoins(
					sdk.NewInt64Coin(appparams.BondDenom, 251),
				),
			},
		},
		{
			"query account summary",
			cli.GetCmdQueryAccountSummary(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryAccountSummaryResponse{},
			&types.QueryAccountSummaryResponse{
				// This result is umee's oracle exchange rate from
				// app/test_helpers.go/IntegrationTestNetworkConfig
				// times the amount of umee, and then times params
				// (1001 / 1000000) * 34.21 = 0.03424421
				SuppliedValue: sdk.MustNewDecFromStr("0.03424421"),
				// (1001 / 1000000) * 34.21 = 0.03424421
				CollateralValue: sdk.MustNewDecFromStr("0.03424421"),
				// (251 / 1000000) * 34.21 = 0.00858671
				BorrowedValue: sdk.MustNewDecFromStr("0.00858671"),
				// (1001 / 1000000) * 34.21 * 0.25 = 0.0085610525
				BorrowLimit: sdk.MustNewDecFromStr("0.0085610525"),
				// (1001 / 1000000) * 0.25 * 34.21 = 0.0085610525
				LiquidationThreshold: &lt1,
			},
		},
		{
			"query max withdraw (borrow limit reached)",
			cli.GetCmdQueryMaxWithdraw(),
			[]string{
				val.Address.String(),
				"uumee",
			},
			false,
			&types.QueryMaxWithdrawResponse{},
			&types.QueryMaxWithdrawResponse{
				Tokens:  sdk.NewCoins(),
				UTokens: sdk.NewCoins(),
			},
		},
		{
			"query max borrow (borrow limit reached)",
			cli.GetCmdQueryMaxBorrow(),
			[]string{
				val.Address.String(),
				"uumee",
			},
			false,
			&types.QueryMaxBorrowResponse{},
			&types.QueryMaxBorrowResponse{
				Tokens: sdk.NewCoins(),
			},
		},
	}

	postQueries := []testQuery{
		{
			"query account balances",
			cli.GetCmdQueryAccountBalances(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryAccountBalancesResponse{},
			&types.QueryAccountBalancesResponse{
				Supplied: sdk.NewCoins(
					sdk.NewInt64Coin(appparams.BondDenom, 201), // slightly increased uToken exchange rate
				),
				Collateral: sdk.NewCoins(
					sdk.NewInt64Coin(types.ToUTokenDenom(appparams.BondDenom), 100),
				),
				Borrowed: sdk.NewCoins(),
			},
		},
		{
			"query max withdraw (after repay)",
			cli.GetCmdQueryMaxWithdraw(),
			[]string{
				val.Address.String(),
				"uumee",
			},
			false,
			&types.QueryMaxWithdrawResponse{},
			&types.QueryMaxWithdrawResponse{
				Tokens: sdk.NewCoins(
					sdk.NewCoin("uumee", sdk.NewInt(201)),
				),
				UTokens: sdk.NewCoins(
					sdk.NewCoin("u/uumee", sdk.NewInt(200)),
				),
			},
		},
		{
			"query max borrow (after repay)",
			cli.GetCmdQueryMaxBorrow(),
			[]string{
				val.Address.String(),
				"uumee",
			},
			false,
			&types.QueryMaxBorrowResponse{},
			&types.QueryMaxBorrowResponse{
				Tokens: sdk.NewCoins(
					sdk.NewCoin("uumee", sdk.NewInt(25)),
				),
			},
		},
		{
			"query all max withdraw (after repay)",
			cli.GetCmdQueryMaxWithdraw(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryMaxWithdrawResponse{},
			&types.QueryMaxWithdrawResponse{
				Tokens: sdk.NewCoins(
					sdk.NewCoin("uumee", sdk.NewInt(201)),
				),
				UTokens: sdk.NewCoins(
					sdk.NewCoin("u/uumee", sdk.NewInt(200)),
				),
			},
		},
		{
			"query all max borrow (after repay)",
			cli.GetCmdQueryMaxBorrow(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryMaxBorrowResponse{},
			&types.QueryMaxBorrowResponse{
				Tokens: sdk.NewCoins(
					sdk.NewCoin("uumee", sdk.NewInt(25)),
				),
			},
		},
	}

	lastQueries := []testQuery{
		{
			"query account balances (empty after withdraw max)",
			cli.GetCmdQueryAccountBalances(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryAccountBalancesResponse{},
			&types.QueryAccountBalancesResponse{
				Supplied:   sdk.NewCoins(),
				Collateral: sdk.NewCoins(),
				Borrowed:   sdk.NewCoins(),
			},
		},

		{
			"query max withdraw (after withdraw max)",
			cli.GetCmdQueryMaxWithdraw(),
			[]string{
				val.Address.String(),
				"uumee",
			},
			false,
			&types.QueryMaxWithdrawResponse{},
			&types.QueryMaxWithdrawResponse{
				Tokens:  sdk.NewCoins(),
				UTokens: sdk.NewCoins(),
			},
		},
	}

	// These queries do not require any borrower setup
	s.runTestQueries(initialQueries...)

	// These transactions will set up nonzero leverage positions and allow nonzero query results
	s.runTestTransactions(
		supply,
		addCollateral,
		supplyCollateral,
		borrow,
		maxborrow,
	)

	// These queries run while the supplying and borrowing is active to produce nonzero output
	s.runTestQueries(nonzeroQueries...)

	// These transactions run after nonzero queries are finished
	s.runTestTransactions(
		liquidate,
		repay,
		removeCollateral,
		withdraw,
	)

	// Confirm additional transaction effects
	s.runTestQueries(postQueries...)

	// This transaction will run last
	s.runTestTransactions(
		withdrawMax,
	)

	// Confirm withdraw max transaction effects
	s.runTestQueries(lastQueries...)
}
