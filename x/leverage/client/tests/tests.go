package tests

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v3/app/params"
	"github.com/umee-network/umee/v3/x/leverage/client/cli"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

func (s *IntegrationTestSuite) TestInvalidQueries() {
	invalidQueries := []testQuery{
		{
			"query market summary - invalid denom",
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
	}

	// These queries do not require any borrower setup because they contain invalid arguments
	s.runTestQueries(invalidQueries...)
}

func (s *IntegrationTestSuite) TestLeverageScenario() {
	val := s.network.Validators[0]

	oraclePrice := sdk.MustNewDecFromStr("0.00003421")

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
					{
						// must match app/test_helpers.go/IntegrationTestNetworkConfig
						BaseDenom:              appparams.BondDenom,
						SymbolDenom:            appparams.DisplayDenom,
						Exponent:               6,
						ReserveFactor:          sdk.MustNewDecFromStr("0.1"),
						CollateralWeight:       sdk.MustNewDecFromStr("0.05"),
						LiquidationThreshold:   sdk.MustNewDecFromStr("0.05"),
						BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
						KinkBorrowRate:         sdk.MustNewDecFromStr("0.2"),
						MaxBorrowRate:          sdk.MustNewDecFromStr("1.5"),
						KinkUtilization:        sdk.MustNewDecFromStr("0.2"),
						LiquidationIncentive:   sdk.MustNewDecFromStr("0.18"),
						EnableMsgSupply:        true,
						EnableMsgBorrow:        true,
						Blacklist:              false,
						MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
						MaxSupplyUtilization:   sdk.MustNewDecFromStr("1"),
						MinCollateralLiquidity: sdk.MustNewDecFromStr("0"),
						MaxSupply:              sdk.NewInt(100000000000),
					},
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
				SymbolDenom:        "UMEE",
				Exponent:           6,
				OraclePrice:        &oraclePrice,
				UTokenExchangeRate: sdk.OneDec(),
				// Borrow rate * (1 - ReserveFactor - OracleRewardFactor)
				// 1.50 * (1 - 0.10 - 0.01) = 0.89 * 1.5 = 1.335
				Supply_APY: sdk.MustNewDecFromStr("1.335"),
				// This is an edge case technically - when effective supply, meaning
				// module balance + total borrows, is zero, utilization (0/0) is
				// interpreted as 100% so max borrow rate (150% APY) is used.
				Borrow_APY:             sdk.MustNewDecFromStr("1.50"),
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
	}

	supply := testTransaction{
		"supply",
		cli.GetCmdSupply(),
		[]string{
			"1000uumee",
		},
		nil,
	}

	addCollateral := testTransaction{
		"add collateral",
		cli.GetCmdCollateralize(),
		[]string{
			"1000u/uumee",
		},
		nil,
	}

	borrow := testTransaction{
		"borrow",
		cli.GetCmdBorrow(),
		[]string{
			"50uumee",
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
			"50uumee", // repays only the remaining borrowed balance, reduced automatically from 50
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
					sdk.NewInt64Coin(appparams.BondDenom, 1000),
				),
				Collateral: sdk.NewCoins(
					sdk.NewInt64Coin(types.ToUTokenDenom(appparams.BondDenom), 1000),
				),
				Borrowed: sdk.NewCoins(
					sdk.NewInt64Coin(appparams.BondDenom, 51),
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
				// (1000 / 1000000) * 34.21 = 0.03421
				SuppliedValue: sdk.MustNewDecFromStr("0.03421"),
				// (1000 / 1000000) * 34.21 = 0.03421
				CollateralValue: sdk.MustNewDecFromStr("0.03421"),
				// (51 / 1000000) * 34.21 = 0.00174471
				BorrowedValue: sdk.MustNewDecFromStr("0.00174471"),
				// (1000 / 1000000) * 34.21 * 0.05 = 0.0017105
				BorrowLimit: sdk.MustNewDecFromStr("0.0017105"),
				// (1000 / 1000000) * 0.05 * 34.21 = 0.0017105
				LiquidationThreshold: sdk.MustNewDecFromStr("0.0017105"),
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
	}

	// These queries do not require any borrower setup
	s.runTestQueries(initialQueries...)

	// These transactions will set up nonzero leverage positions and allow nonzero query results
	s.runTestTransactions(
		supply,
		addCollateral,
		borrow,
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

	// Confirm cleanup transaction effects
	s.runTestQueries(postQueries...)
}
