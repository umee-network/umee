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
			"query total supplied - invalid denom",
			cli.GetCmdQueryTotalSupplied(),
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
			"query total supplied value - invalid denom",
			cli.GetCmdQueryTotalSuppliedValue(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query supply APY - invalid denom",
			cli.GetCmdQuerySupplyAPY(),
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
			"query supplied - invalid address",
			cli.GetCmdQuerySupplied(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query supplied - invalid denom",
			cli.GetCmdQuerySupplied(),
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
			"query supplied value - invalid address",
			cli.GetCmdQuerySuppliedValue(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		testQuery{
			"query supplied value - invalid denom",
			cli.GetCmdQuerySuppliedValue(),
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
			"query supply APY",
			cli.GetCmdQuerySupplyAPY(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QuerySupplyAPYResponse{},
			// Supply utilization is zero, so supply APY is 0%
			&types.QuerySupplyAPYResponse{APY: sdk.MustNewDecFromStr("0")},
		},
		testQuery{
			"query borrow APY",
			cli.GetCmdQueryBorrowAPY(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryBorrowAPYResponse{},
			// Supply utilization is 0% so base borrow rate (2% APY) is used.
			&types.QueryBorrowAPYResponse{APY: sdk.MustNewDecFromStr("0.02")},
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
			"4uumee",
		},
		nil,
	}

	fixCollateral := testTransaction{
		"add back collateral received from liquidation",
		cli.GetCmdCollateralize(),
		[]string{
			val.Address.String(),
			"4u/uumee",
		},
		nil,
	}

	repay := testTransaction{
		"repay",
		cli.GetCmdRepay(),
		[]string{
			val.Address.String(),
			"51uumee",
		},
		nil,
	}

	removeCollateral := testTransaction{
		"remove collateral",
		cli.GetCmdDecollateralize(),
		[]string{
			val.Address.String(),
			"1000u/uumee",
		},
		nil,
	}

	withdraw := testTransaction{
		"withdraw",
		cli.GetCmdWithdraw(),
		[]string{
			val.Address.String(),
			"1000u/uumee",
		},
		nil,
	}

	nonzeroQueries := []TestCase{
		testQuery{
			"query token total supplied",
			cli.GetCmdQueryTotalSupplied(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryTotalSuppliedResponse{},
			&types.QueryTotalSuppliedResponse{TotalSupplied: sdk.NewInt(1001)},
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
			"query total supplied value",
			cli.GetCmdQueryTotalSuppliedValue(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryTotalSuppliedValueResponse{},
			&types.QueryTotalSuppliedValueResponse{TotalSuppliedValue: sdk.MustNewDecFromStr("0.03424421")},
		},
		testQuery{
			"query supplied - all",
			cli.GetCmdQuerySupplied(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QuerySuppliedResponse{},
			&types.QuerySuppliedResponse{
				Supplied: sdk.NewCoins(
					sdk.NewInt64Coin(umeeapp.BondDenom, 1001),
				),
			},
		},
		testQuery{
			"query supplied - denom",
			cli.GetCmdQuerySupplied(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QuerySuppliedResponse{},
			&types.QuerySuppliedResponse{
				Supplied: sdk.NewCoins(
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
			"query total borrowed - denom",
			cli.GetCmdQueryTotalBorrowed(),
			[]string{
				"uumee",
			},
			false,
			&types.QueryTotalBorrowedResponse{},
			&types.QueryTotalBorrowedResponse{
				Amount: sdk.NewInt(47),
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
			"query total collateral - denom",
			cli.GetCmdQueryTotalCollateral(),
			[]string{
				"u/uumee",
			},
			false,
			&types.QueryTotalCollateralResponse{},
			&types.QueryTotalCollateralResponse{
				Amount: sdk.NewInt(1000),
			},
		},
		testQuery{
			"query supplied value - all",
			cli.GetCmdQuerySuppliedValue(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QuerySuppliedValueResponse{},
			&types.QuerySuppliedValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				// This result is umee's oracle exchange rate times the
				// amount supplied.
				SuppliedValue: sdk.MustNewDecFromStr("0.03424421"),
				// (1001 / 1000000) umee * 34.21 = 0.03424421
			},
		},
		testQuery{
			"query supplied value - denom",
			cli.GetCmdQuerySuppliedValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QuerySuppliedValueResponse{},
			&types.QuerySuppliedValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				SuppliedValue: sdk.MustNewDecFromStr("0.03424421"),
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
		supply,
		addCollateral,
		borrow,
		liquidate,
		fixCollateral,
	)

	// These transactions are deferred to run after nonzero queries are finished
	defer s.runTestCases(
		repay,
		removeCollateral,
		withdraw,
	)

	// These queries run while the supplying and borrowing is active to produce nonzero output
	s.runTestCases(nonzeroQueries...)
}
