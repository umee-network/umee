package tests

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/x/leverage/client/cli"
	"github.com/umee-network/umee/v4/x/leverage/fixtures"
	"github.com/umee-network/umee/v4/x/leverage/types"

	itestsuite "github.com/umee-network/umee/v4/tests/integration_suite"
)

func (s *IntegrationTests) TestInvalidQueries() {
	invalidQueries := []itestsuite.TestQuery{
		{
			Msg:     "query market summary - denom not registered",
			Command: cli.GetCmdQueryMarketSummary(),
			Args: []string{
				"abcd",
			},
			ResponseType:     nil,
			ExpectedResponse: nil,
			ErrMsg:           "not a registered Token",
		},
		{
			Msg:     "query account balances - invalid address",
			Command: cli.GetCmdQueryAccountBalances(),
			Args: []string{
				"xyz",
			},
			ResponseType:     nil,
			ExpectedResponse: nil,
			ErrMsg:           "invalid bech32",
		},
		{
			Msg:     "query account summary - invalid address",
			Command: cli.GetCmdQueryAccountSummary(),
			Args: []string{
				"xyz",
			},
			ResponseType:     nil,
			ExpectedResponse: nil,
			ErrMsg:           "invalid bech32",
		},
		{
			Msg:     "query max withdraw - invalid address",
			Command: cli.GetCmdQueryMaxWithdraw(),
			Args: []string{
				"xyz",
				"uumee",
			},
			ResponseType:     nil,
			ExpectedResponse: nil,
			ErrMsg:           "invalid bech32",
		},
		{
			Msg:              "query registered token - denom not registered",
			Command:          cli.GetCmdQueryRegisteredTokens(),
			Args:             []string{"umm"},
			ResponseType:     nil,
			ExpectedResponse: nil,
			ErrMsg:           "not a registered Token",
		},
	}

	// These queries do not require any borrower setup because they contain invalid arguments
	s.RunTestQueries(invalidQueries...)
}

func (s *IntegrationTests) TestLeverageScenario() {
	val := s.Network.Validators[0]

	oracleSymbolPrice := sdk.MustNewDecFromStr("34.21")

	initialQueries := []itestsuite.TestQuery{
		{
			Msg:          "query params",
			Command:      cli.GetCmdQueryParams(),
			Args:         []string{},
			ResponseType: &types.QueryParamsResponse{},
			ExpectedResponse: &types.QueryParamsResponse{
				Params: types.DefaultParams(),
			},
			ErrMsg: "",
		},
		{
			Msg:          "query registered tokens",
			Command:      cli.GetCmdQueryRegisteredTokens(),
			Args:         []string{},
			ResponseType: &types.QueryRegisteredTokensResponse{},
			ExpectedResponse: &types.QueryRegisteredTokensResponse{
				Registry: []types.Token{
					fixtures.Token(appparams.BondDenom, appparams.DisplayDenom, 6),
				},
			},
			ErrMsg: "",
		},
		{
			Msg:          "query registered token info by base_denom",
			Command:      cli.GetCmdQueryRegisteredTokens(),
			Args:         []string{appparams.BondDenom},
			ResponseType: &types.QueryRegisteredTokensResponse{},
			ExpectedResponse: &types.QueryRegisteredTokensResponse{
				Registry: []types.Token{
					fixtures.Token(appparams.BondDenom, appparams.DisplayDenom, 6),
				},
			},
			ErrMsg: "",
		},
		{
			Msg:     "query market summary - zero supply",
			Command: cli.GetCmdQueryMarketSummary(),
			Args: []string{
				appparams.BondDenom,
			},
			ResponseType: &types.QueryMarketSummaryResponse{},
			ExpectedResponse: &types.QueryMarketSummaryResponse{
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
			Msg:          "query bad debts",
			Command:      cli.GetCmdQueryBadDebts(),
			Args:         []string{},
			ResponseType: &types.QueryBadDebtsResponse{},
			ExpectedResponse: &types.QueryBadDebtsResponse{
				Targets: []types.BadDebt{},
			},
			ErrMsg: "",
		},
		{
			Msg:     "query max withdraw (zero)",
			Command: cli.GetCmdQueryMaxWithdraw(),
			Args: []string{
				val.Address.String(),
				"uumee",
			},
			ResponseType: &types.QueryMaxWithdrawResponse{},
			ExpectedResponse: &types.QueryMaxWithdrawResponse{
				Tokens:  sdk.NewCoins(),
				UTokens: sdk.NewCoins(),
			},
			ErrMsg: "",
		},
		{
			Msg:     "query max borrow (zero)",
			Command: cli.GetCmdQueryMaxBorrow(),
			Args: []string{
				val.Address.String(),
				"uumee",
			},
			ResponseType: &types.QueryMaxBorrowResponse{},
			ExpectedResponse: &types.QueryMaxBorrowResponse{
				Tokens: sdk.NewCoins(),
			},
			ErrMsg: "",
		},
	}

	supply := itestsuite.TestTransaction{
		Msg:     "supply",
		Command: cli.GetCmdSupply(),
		Args: []string{
			"700uumee",
		},
		ExpectedErr: nil,
	}

	addCollateral := itestsuite.TestTransaction{
		Msg:     "add collateral",
		Command: cli.GetCmdCollateralize(),
		Args: []string{
			"700u/uumee",
		},
		ExpectedErr: nil,
	}

	supplyCollateral := itestsuite.TestTransaction{
		Msg:     "supply collateral",
		Command: cli.GetCmdSupplyCollateral(),
		Args: []string{
			"300uumee",
		},
		ExpectedErr: nil,
	}

	borrow := itestsuite.TestTransaction{
		Msg:     "borrow",
		Command: cli.GetCmdBorrow(),
		Args: []string{
			"150uumee",
		},
		ExpectedErr: nil,
	}

	maxborrow := itestsuite.TestTransaction{
		Msg:     "max-borrow",
		Command: cli.GetCmdMaxBorrow(),
		Args: []string{
			"uumee", // should borrow up to the max of 250 uumee, which will become 251 due to rounding
		},
		ExpectedErr: nil,
	}

	liquidate := itestsuite.TestTransaction{
		Msg:     "liquidate",
		Command: cli.GetCmdLiquidate(),
		Args: []string{
			val.Address.String(),
			"5uumee", // borrower attempts to liquidate itself, but is ineligible
			"uumee",
		},
		ExpectedErr: types.ErrLiquidationIneligible,
	}

	repay := itestsuite.TestTransaction{
		Msg:     "repay",
		Command: cli.GetCmdRepay(),
		Args: []string{
			"255uumee", // repays only the remaining borrowed balance, reduced automatically from 255
		},
		ExpectedErr: nil,
	}

	removeCollateral := itestsuite.TestTransaction{
		Msg:     "remove collateral",
		Command: cli.GetCmdDecollateralize(),
		Args: []string{
			"900u/uumee", // 100 u/uumee will remain
		},
		ExpectedErr: nil,
	}

	withdraw := itestsuite.TestTransaction{
		Msg:     "withdraw",
		Command: cli.GetCmdWithdraw(),
		Args: []string{
			"800u/uumee", // 200 u/uumee will remain
		},
		ExpectedErr: nil,
	}

	withdrawMax := itestsuite.TestTransaction{
		Msg:     "withdraw max",
		Command: cli.GetCmdMaxWithdraw(),
		Args: []string{
			"uumee",
		},
		ExpectedErr: nil,
	}

	lt1 := sdk.MustNewDecFromStr("0.0089034946")

	nonzeroQueries := []itestsuite.TestQuery{
		{
			Msg:     "query account balances",
			Command: cli.GetCmdQueryAccountBalances(),
			Args: []string{
				val.Address.String(),
			},
			ResponseType: &types.QueryAccountBalancesResponse{},
			ExpectedResponse: &types.QueryAccountBalancesResponse{
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
			ErrMsg: "",
		},
		{
			Msg:     "query account summary",
			Command: cli.GetCmdQueryAccountSummary(),
			Args: []string{
				val.Address.String(),
			},
			ResponseType: &types.QueryAccountSummaryResponse{},
			ExpectedResponse: &types.QueryAccountSummaryResponse{
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
				// (1001 / 1000000) * 0.26 * 34.21 = 0.008903494600000000
				LiquidationThreshold: &lt1,
			},
			ErrMsg: "",
		},
		{
			Msg:     "query max withdraw (borrow limit reached)",
			Command: cli.GetCmdQueryMaxWithdraw(),
			Args: []string{
				val.Address.String(),
				"uumee",
			},
			ResponseType: &types.QueryMaxWithdrawResponse{},
			ExpectedResponse: &types.QueryMaxWithdrawResponse{
				Tokens:  sdk.NewCoins(),
				UTokens: sdk.NewCoins(),
			},
			ErrMsg: "",
		},
		{
			Msg:     "query max borrow (borrow limit reached)",
			Command: cli.GetCmdQueryMaxBorrow(),
			Args: []string{
				val.Address.String(),
				"uumee",
			},
			ResponseType: &types.QueryMaxBorrowResponse{},
			ExpectedResponse: &types.QueryMaxBorrowResponse{
				Tokens: sdk.NewCoins(),
			},
			ErrMsg: "",
		},
	}

	postQueries := []itestsuite.TestQuery{
		{
			Msg:     "query account balances",
			Command: cli.GetCmdQueryAccountBalances(),
			Args: []string{
				val.Address.String(),
			},
			ResponseType: &types.QueryAccountBalancesResponse{},
			ExpectedResponse: &types.QueryAccountBalancesResponse{
				Supplied: sdk.NewCoins(
					sdk.NewInt64Coin(appparams.BondDenom, 201), // slightly increased uToken exchange rate
				),
				Collateral: sdk.NewCoins(
					sdk.NewInt64Coin(types.ToUTokenDenom(appparams.BondDenom), 100),
				),
				Borrowed: sdk.NewCoins(),
			},
			ErrMsg: "",
		},
		{
			Msg:     "query max withdraw (after repay)",
			Command: cli.GetCmdQueryMaxWithdraw(),
			Args: []string{
				val.Address.String(),
				"uumee",
			},
			ResponseType: &types.QueryMaxWithdrawResponse{},
			ExpectedResponse: &types.QueryMaxWithdrawResponse{
				Tokens: sdk.NewCoins(
					sdk.NewCoin("uumee", sdk.NewInt(201)),
				),
				UTokens: sdk.NewCoins(
					sdk.NewCoin("u/uumee", sdk.NewInt(200)),
				),
			},
			ErrMsg: "",
		},
		{
			Msg:     "query max borrow (after repay)",
			Command: cli.GetCmdQueryMaxBorrow(),
			Args: []string{
				val.Address.String(),
				"uumee",
			},
			ResponseType: &types.QueryMaxBorrowResponse{},
			ExpectedResponse: &types.QueryMaxBorrowResponse{
				Tokens: sdk.NewCoins(
					sdk.NewCoin("uumee", sdk.NewInt(25)),
				),
			},
			ErrMsg: "",
		},
		{
			Msg:     "query all max withdraw (after repay)",
			Command: cli.GetCmdQueryMaxWithdraw(),
			Args: []string{
				val.Address.String(),
			},
			ResponseType: &types.QueryMaxWithdrawResponse{},
			ExpectedResponse: &types.QueryMaxWithdrawResponse{
				Tokens: sdk.NewCoins(
					sdk.NewCoin("uumee", sdk.NewInt(201)),
				),
				UTokens: sdk.NewCoins(
					sdk.NewCoin("u/uumee", sdk.NewInt(200)),
				),
			},
			ErrMsg: "",
		},
		{
			Msg:     "query all max borrow (after repay)",
			Command: cli.GetCmdQueryMaxBorrow(),
			Args: []string{
				val.Address.String(),
			},
			ResponseType: &types.QueryMaxBorrowResponse{},
			ExpectedResponse: &types.QueryMaxBorrowResponse{
				Tokens: sdk.NewCoins(
					sdk.NewCoin("uumee", sdk.NewInt(25)),
				),
			},
			ErrMsg: "",
		},
	}

	lastQueries := []itestsuite.TestQuery{
		{
			Msg:     "query account balances (empty after withdraw max)",
			Command: cli.GetCmdQueryAccountBalances(),
			Args: []string{
				val.Address.String(),
			},
			ResponseType: &types.QueryAccountBalancesResponse{},
			ExpectedResponse: &types.QueryAccountBalancesResponse{
				Supplied:   sdk.NewCoins(),
				Collateral: sdk.NewCoins(),
				Borrowed:   sdk.NewCoins(),
			},
			ErrMsg: "",
		},

		{
			Msg:     "query max withdraw (after withdraw max)",
			Command: cli.GetCmdQueryMaxWithdraw(),
			Args: []string{
				val.Address.String(),
				"uumee",
			},
			ResponseType: &types.QueryMaxWithdrawResponse{},
			ExpectedResponse: &types.QueryMaxWithdrawResponse{
				Tokens:  sdk.NewCoins(),
				UTokens: sdk.NewCoins(),
			},
			ErrMsg: "",
		},
	}

	// These queries do not require any borrower setup
	s.RunTestQueries(initialQueries...)

	// These transactions will set up nonzero leverage positions and allow nonzero query results
	s.RunTestTransactions(
		supply,
		addCollateral,
		supplyCollateral,
		borrow,
		maxborrow,
	)

	// These queries run while the supplying and borrowing is active to produce nonzero output
	s.RunTestQueries(nonzeroQueries...)

	// These transactions run after nonzero queries are finished
	s.RunTestTransactions(
		liquidate,
		repay,
		removeCollateral,
		withdraw,
	)

	// Confirm additional transaction effects
	s.RunTestQueries(postQueries...)

	// This transaction will run last
	s.RunTestTransactions(
		withdrawMax,
	)

	// Confirm withdraw max transaction effects
	s.RunTestQueries(lastQueries...)
}
