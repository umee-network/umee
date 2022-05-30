package tests

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/leverage/client/cli"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

func (s *IntegrationTestSuite) TestInvalidQueries() {
	val := s.network.Validators[0]

	invalidQueries := []TestCase{
		testQuery{
			"query reserved - invalid denom",
			cli.GetCmdQueryReserveAmount(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query token market size - invalid denom",
			cli.GetCmdQueryTokenMarketSize(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query available to borrow - invalid denom",
			cli.GetCmdQueryAvailableBorrow(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query exchange rate - invalid denom",
			cli.GetCmdQueryExchangeRate(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query market size - invalid denom",
			cli.GetCmdQueryMarketSize(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query lend APY - invalid denom",
			cli.GetCmdQueryLendAPY(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query borrow APY - invalid denom",
			cli.GetCmdQueryBorrowAPY(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query loaned - invalid address",
			cli.GetCmdQueryLoaned(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query loaned - invalid denom",
			cli.GetCmdQueryLoaned(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=abcd", cli.FlagDenom),
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query borrowed - invalid address",
			cli.GetCmdQueryBorrowed(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query borrowed - invalid denom",
			cli.GetCmdQueryBorrowed(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=abcd", cli.FlagDenom),
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query collateral - invalid address",
			cli.GetCmdQueryCollateral(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query collateral - invalid denom",
			cli.GetCmdQueryCollateral(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=abcd", cli.FlagDenom),
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query collateral setting - invalid address",
			cli.GetCmdQueryCollateralSetting(),
			[]string{
				"xyz",
				"u/uumee",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query collateral setting - invalid denom",
			cli.GetCmdQueryCollateralSetting(),
			[]string{
				val.Address.String(),
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query loaned value - invalid address",
			cli.GetCmdQueryLoanedValue(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query loaned value - invalid denom",
			cli.GetCmdQueryLoanedValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=abcd", cli.FlagDenom),
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query collateral value - invalid address",
			cli.GetCmdQueryCollateralValue(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query collateral value - invalid denom",
			cli.GetCmdQueryCollateralValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=u/abcd", cli.FlagDenom),
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query borrowed value - invalid address",
			cli.GetCmdQueryBorrowedValue(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query borrowed value - invalid denom",
			cli.GetCmdQueryBorrowedValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=abcd", cli.FlagDenom),
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query borrow limit - invalid address",
			cli.GetCmdQueryBorrowLimit(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query liquidation threshold - invalid address",
			cli.GetCmdQueryLiquidationThreshold(),
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
			cli.GetCmdQueryAllRegisteredTokens(),
			[]string{},
			false,
			&types.QueryRegisteredTokensResponse{},
			&types.QueryRegisteredTokensResponse{
				Registry: []types.Token{
					{
						// must match app/test_helpers.go/IntegrationTestNetworkConfig
						BaseDenom:            umeeapp.BondDenom,
						SymbolDenom:          umeeapp.DisplayDenom,
						Exponent:             6,
						ReserveFactor:        sdk.MustNewDecFromStr("0.1"),
						CollateralWeight:     sdk.MustNewDecFromStr("0.05"),
						LiquidationThreshold: sdk.MustNewDecFromStr("0.05"),
						BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
						KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
						MaxBorrowRate:        sdk.MustNewDecFromStr("1.5"),
						KinkUtilizationRate:  sdk.MustNewDecFromStr("0.2"),
						LiquidationIncentive: sdk.MustNewDecFromStr("0.18"),
						EnableMsgLend:        true,
						EnableMsgBorrow:      true,
						Blacklist:            false,
					},
				},
			},
		},
		testQuery{
			"query reserve amount",
			cli.GetCmdQueryReserveAmount(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryReserveAmountResponse{},
			&types.QueryReserveAmountResponse{
				Amount: sdk.ZeroInt(),
			},
		},
		testQuery{
			"query exchange rate",
			cli.GetCmdQueryExchangeRate(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryExchangeRateResponse{},
			&types.QueryExchangeRateResponse{
				ExchangeRate: sdk.OneDec(),
			},
		},
		testQuery{
			"query lend APY",
			cli.GetCmdQueryLendAPY(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryLendAPYResponse{},
			// Borrow rate * (1 - ReserveFactor - OracleRewardFactor)
			// 1.50 * (1 - 0.10 - 0.01) = 0.89 * 1.5 = 1.335
			&types.QueryLendAPYResponse{APY: sdk.MustNewDecFromStr("1.335")},
		},
		testQuery{
			"query borrow APY",
			cli.GetCmdQueryBorrowAPY(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryBorrowAPYResponse{},
			// This is an edge case technically - when effective supply, meaning
			// module balance + total borrows, is zero, utilization (0/0) is
			// interpreted as 100% so max borrow rate (150% APY) is used.
			&types.QueryBorrowAPYResponse{APY: sdk.MustNewDecFromStr("1.50")},
		},
	}

	lend := testTransaction{
		"lend",
		cli.GetCmdLendAsset(),
		[]string{
			val.Address.String(),
			"1000uumee",
		},
		nil,
	}

	setCollateral := testTransaction{
		"set collateral",
		cli.GetCmdSetCollateral(),
		[]string{
			val.Address.String(),
			"u/uumee",
			"true",
		},
		nil,
	}

	borrow := testTransaction{
		"borrow",
		cli.GetCmdBorrowAsset(),
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
			"1uumee",
		},
		nil,
	}

	repay := testTransaction{
		"repay",
		cli.GetCmdRepayAsset(),
		[]string{
			val.Address.String(),
			"51uumee",
		},
		nil,
	}

	withdraw := testTransaction{
		"withdraw",
		cli.GetCmdWithdrawAsset(),
		[]string{
			val.Address.String(),
			"1000uumee",
		},
		nil,
	}

	nonzeroQueries := []TestCase{
		testQuery{
			"query token market size",
			cli.GetCmdQueryTokenMarketSize(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryTokenMarketSizeResponse{},
			&types.QueryTokenMarketSizeResponse{MarketSize: sdk.NewInt(1001)},
		},
		testQuery{
			"query available to borrow",
			cli.GetCmdQueryAvailableBorrow(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryAvailableBorrowResponse{},
			&types.QueryAvailableBorrowResponse{Amount: sdk.NewInt(955)},
		},
		testQuery{
			"query market size",
			cli.GetCmdQueryMarketSize(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryMarketSizeResponse{},
			&types.QueryMarketSizeResponse{MarketSizeUsd: sdk.MustNewDecFromStr("0.03424421")},
		},
		testQuery{
			"query loaned - all",
			cli.GetCmdQueryLoaned(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryLoanedResponse{},
			&types.QueryLoanedResponse{
				Loaned: sdk.NewCoins(
					sdk.NewInt64Coin(umeeapp.BondDenom, 1001),
				),
			},
		},
		testQuery{
			"query loaned - denom",
			cli.GetCmdQueryLoaned(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QueryLoanedResponse{},
			&types.QueryLoanedResponse{
				Loaned: sdk.NewCoins(
					sdk.NewInt64Coin(umeeapp.BondDenom, 1001),
				),
			},
		},
		testQuery{
			"query borrowed - all",
			cli.GetCmdQueryBorrowed(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryBorrowedResponse{},
			&types.QueryBorrowedResponse{
				Borrowed: sdk.NewCoins(
					sdk.NewInt64Coin(umeeapp.BondDenom, 47),
				),
			},
		},
		testQuery{
			"query borrowed - denom",
			cli.GetCmdQueryBorrowed(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QueryBorrowedResponse{},
			&types.QueryBorrowedResponse{
				Borrowed: sdk.NewCoins(
					sdk.NewInt64Coin(umeeapp.BondDenom, 47),
				),
			},
		},
		testQuery{
			"query collateral setting",
			cli.GetCmdQueryCollateralSetting(),
			[]string{
				val.Address.String(),
				"u/uumee",
			},
			false,
			&types.QueryCollateralSettingResponse{},
			&types.QueryCollateralSettingResponse{
				Enabled: true,
			},
		},
		testQuery{
			"query collateral - all",
			cli.GetCmdQueryCollateral(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryCollateralResponse{},
			&types.QueryCollateralResponse{
				Collateral: sdk.NewCoins(
					sdk.NewInt64Coin("u/uumee", 1000),
				),
			},
		},
		testQuery{
			"query collateral - denom",
			cli.GetCmdQueryCollateral(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=u/uumee", cli.FlagDenom),
			},
			false,
			&types.QueryCollateralResponse{},
			&types.QueryCollateralResponse{
				Collateral: sdk.NewCoins(
					sdk.NewInt64Coin("u/uumee", 1000),
				),
			},
		},
		testQuery{
			"query loaned value - all",
			cli.GetCmdQueryLoanedValue(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryLoanedValueResponse{},
			&types.QueryLoanedValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				// This result is umee's oracle exchange rate times the
				// amount loaned.
				LoanedValue: sdk.MustNewDecFromStr("0.03424421"),
				// (1001 / 1000000) umee * 34.21 = 0.03424421
			},
		},
		testQuery{
			"query loaned value - denom",
			cli.GetCmdQueryLoanedValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QueryLoanedValueResponse{},
			&types.QueryLoanedValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				LoanedValue: sdk.MustNewDecFromStr("0.03424421"),
				// (1001 / 1000000) umee * 34.21 = 0.03424421
			},
		},
		testQuery{
			"query collateral value - all",
			cli.GetCmdQueryCollateralValue(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryCollateralValueResponse{},
			&types.QueryCollateralValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				CollateralValue: sdk.MustNewDecFromStr("0.03424421"),
				// (1001 / 1000000) umee * 34.21 = 0.03424421
			},
		},
		testQuery{
			"query collateral value - denom",
			cli.GetCmdQueryCollateralValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=u/uumee", cli.FlagDenom),
			},
			false,
			&types.QueryCollateralValueResponse{},
			&types.QueryCollateralValueResponse{
				CollateralValue: sdk.MustNewDecFromStr("0.03424421"),
				// (1001 / 1000000) umee * 34.21 = 0.03424421
			},
		},
		testQuery{
			"query borrowed value - all",
			cli.GetCmdQueryBorrowedValue(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryBorrowedValueResponse{},
			&types.QueryBorrowedValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				BorrowedValue: sdk.MustNewDecFromStr("0.00160787"),
				// (51 / 1000000) umee * 34.21 = 0.00160787
			},
		},
		testQuery{
			"query borrowed value - denom",
			cli.GetCmdQueryBorrowedValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QueryBorrowedValueResponse{},
			&types.QueryBorrowedValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				BorrowedValue: sdk.MustNewDecFromStr("0.00160787"),
				// (51 / 1000000) umee * 34.21 = 0.00160787
			},
		},
		testQuery{
			"query borrow limit",
			cli.GetCmdQueryBorrowLimit(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryBorrowLimitResponse{},
			&types.QueryBorrowLimitResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				BorrowLimit: sdk.MustNewDecFromStr("0.0017122105"),
				// (1001 / 1000000) * 0.05 * 34.21 = 0.0017122105
			},
		},
		testQuery{
			"query liquidation threshold",
			cli.GetCmdQueryLiquidationThreshold(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryLiquidationThresholdResponse{},
			&types.QueryLiquidationThresholdResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				LiquidationThreshold: sdk.MustNewDecFromStr("0.0017122105"),
				// (1001 / 1000000) * 0.05 * 34.21 = 0.0017122105
			},
		},
	}

	// These queries do not require any borrower setup
	s.runTestCases(initialQueries...)

	// These transactions will set up nonzero leverage positions and allow nonzero query results
	s.runTestCases(
		lend,
		setCollateral,
		borrow,
		liquidate,
	)

	// These transactions are deferred to run after nonzero queries are finished
	defer s.runTestCases(
		repay,
		withdraw,
	)

	// These queries run while the lending and borrowing is active to produce nonzero output
	s.runTestCases(nonzeroQueries...)
}
