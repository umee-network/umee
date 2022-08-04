package tests

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/leverage/client/cli"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

func (s *IntegrationTestSuite) TestInvalidQueries() {
	invalidQueries := []TestCase{
		testQuery{
			"query market summary - invalid denom",
			cli.GetCmdQueryMarketSummary(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query account balances - invalid address",
			cli.GetCmdQueryAccountBalances(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
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
	s.runTestCases(invalidQueries...)
}

func (s *IntegrationTestSuite) TestLeverageScenario() {
	val := s.network.Validators[0]

	oraclePrice := sdk.MustNewDecFromStr("0.00003421")

	initialQueries := []TestCase{
		testQuery{
			"query params",
			cli.GetCmdQueryParams(),
			[]string{},
			false,
			&types.QueryParamsResponse{},
			&types.QueryParamsResponse{
				Params: types.DefaultParams(),
			},
		},
		testQuery{
			"query registered tokens",
			cli.GetCmdQueryRegisteredTokens(),
			[]string{},
			false,
			&types.QueryRegisteredTokensResponse{},
			&types.QueryRegisteredTokensResponse{
				Registry: []types.Token{
					{
						// must match app/test_helpers.go/IntegrationTestNetworkConfig
						BaseDenom:              umeeapp.BondDenom,
						SymbolDenom:            umeeapp.DisplayDenom,
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
					},
				},
			},
		},
		testQuery{
			"query market summary - zero supply",
			cli.GetCmdQueryMarketSummary(),
			[]string{
				umeeapp.BondDenom,
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
			val.Address.String(),
			"1000uumee",
		},
		nil,
	}

	addCollateral := testTransaction{
		"add collateral",
		cli.GetCmdCollateralize(),
		[]string{
			val.Address.String(),
			"1000u/uumee",
		},
		nil,
	}

	borrow := testTransaction{
		"borrow",
		cli.GetCmdBorrow(),
		[]string{
			val.Address.String(),
			"50uumee",
		},
		nil,
	}

	liquidate := testTransaction{
		"liquidate",
		cli.GetCmdLiquidate(),
		[]string{
			val.Address.String(),
			val.Address.String(),
			"5uumee",
			"uumee",
		},
		nil,
	}

	repay := testTransaction{
		"repay",
		cli.GetCmdRepay(),
		[]string{
			val.Address.String(),
			"50uumee",
		},
		nil,
	}

	removeCollateral := testTransaction{
		"remove collateral",
		cli.GetCmdDecollateralize(),
		[]string{
			val.Address.String(),
			"950u/uumee",
		},
		nil,
	}

	withdraw := testTransaction{
		"withdraw",
		cli.GetCmdWithdraw(),
		[]string{
			val.Address.String(),
			"950u/uumee",
		},
		nil,
	}

	nonzeroQueries := []TestCase{
		testQuery{
			"query account balances",
			cli.GetCmdQueryAccountBalances(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryAccountBalancesResponse{},
			&types.QueryAccountBalancesResponse{
				Supplied: sdk.NewCoins(
					sdk.NewInt64Coin(umeeapp.BondDenom, 1000),
				),
				Collateral: sdk.NewCoins(
					sdk.NewInt64Coin(types.UTokenFromTokenDenom(umeeapp.BondDenom), 1000),
				),
				Borrowed: sdk.NewCoins(
					sdk.NewInt64Coin(umeeapp.BondDenom, 51),
				),
			},
		},
		testQuery{
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
				// (47 / 1000000) * 34.21 = 0.00160787
				BorrowedValue: sdk.MustNewDecFromStr("0.00160787"),
				// (1001 / 1000000) * 34.21 * 0.05 = 0.0017122105
				BorrowLimit: sdk.MustNewDecFromStr("0.0017122105"),
				// (1001 / 1000000) * 0.05 * 34.21 = 0.0017122105
				LiquidationThreshold: sdk.MustNewDecFromStr("0.0017122105"),
			},
		},
	}

	// These queries do not require any borrower setup
	s.runTestCases(initialQueries...)

	// These transactions will set up nonzero leverage positions and allow nonzero query results
	s.runTestCases(
		supply,
		addCollateral,
		borrow,
	)

	// These queries run while the supplying and borrowing is active to produce nonzero output
	s.runTestCases(nonzeroQueries...)

	// These transactions run after nonzero queries are finished
	s.runTestCases(
		liquidate,
		repay,
		removeCollateral,
		withdraw,
	)
}
