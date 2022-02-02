package tests

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/umee-network/umee/app"
	"github.com/umee-network/umee/x/leverage/client/cli"
	"github.com/umee-network/umee/x/leverage/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")

	s.network.Cleanup()
}

type testTransaction struct {
	name        string
	command     *cobra.Command
	args        []string
	expectedErr *sdkerrors.Error
}

// runTestTransactions implements repetitive test logic where transaction commands must be run
// with flags to bypass gas and signature checks, then checked for success or an expected error
func runTestTransactions(s *IntegrationTestSuite, tcs []testTransaction) {
	clientCtx := s.network.Validators[0].ClientCtx

	txFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	for _, tc := range tcs {
		tc := tc // for the linter
		tc.args = append(tc.args, txFlags...)

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.command, tc.args)
			s.Require().NoError(err)

			resp := &sdk.TxResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp), out.String())

			if tc.expectedErr == nil {
				s.Require().Equal(0, int(resp.Code))
			} else {
				s.Require().Equal(int(tc.expectedErr.ABCICode()), int(resp.Code))
			}
		})
	}
}

type testQuery struct {
	name             string
	command          *cobra.Command
	args             []string
	expectErr        bool
	responseType     proto.Message
	expectedResponse proto.Message
}

// runTestQueries implements repetitive test logic where query commands must be run
// then checked for an exact response (including matching fields)
func runTestQueries(s *IntegrationTestSuite, tcs []testQuery) {
	clientCtx := s.network.Validators[0].ClientCtx

	queryFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	for _, tc := range tcs {
		tc := tc // for the linter
		tc.args = append(tc.args, queryFlags...)

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.command, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.responseType), out.String())
				s.Require().Equal(tc.expectedResponse, tc.responseType)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryAllRegisteredTokens() {
	testCases := []testQuery{
		{
			"query registered tokens",
			cli.GetCmdQueryAllRegisteredTokens(),
			[]string{},
			false,
			&types.QueryRegisteredTokensResponse{},
			&types.QueryRegisteredTokensResponse{
				Registry: []types.Token{
					{
						// must match app/beta/test_helpers.go/IntegrationTestNetworkConfig
						BaseDenom:            app.BondDenom,
						SymbolDenom:          app.DisplayDenom,
						Exponent:             6,
						ReserveFactor:        sdk.MustNewDecFromStr("0.1"),
						CollateralWeight:     sdk.MustNewDecFromStr("0.05"),
						BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
						KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
						MaxBorrowRate:        sdk.MustNewDecFromStr("1.5"),
						KinkUtilizationRate:  sdk.MustNewDecFromStr("0.2"),
						LiquidationIncentive: sdk.MustNewDecFromStr("0.18"),
					},
				},
			},
		},
	}

	runTestQueries(s, testCases)
}

func (s *IntegrationTestSuite) TestQueryParams() {
	testCases := []testQuery{
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
	}

	runTestQueries(s, testCases)
}

func (s *IntegrationTestSuite) TestQueryBorrowed() {
	val := s.network.Validators[0]

	setupCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uumee",
			},
			nil,
		},
		{
			"set collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"true",
			},
			nil,
		},
		{
			"borrow",
			cli.GetCmdBorrowAsset(),
			[]string{
				val.Address.String(),
				"50uumee",
			},
			nil,
		},
	}

	testCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryBorrowed(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query all borrowed",
			cli.GetCmdQueryBorrowed(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryBorrowedResponse{},
			&types.QueryBorrowedResponse{
				Borrowed: sdk.NewCoins(
					sdk.NewInt64Coin(app.BondDenom, 50),
				),
			},
		},
		{
			"invalid denom",
			cli.GetCmdQueryBorrowed(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=abcd", cli.FlagDenom),
			},
			true,
			nil,
			nil,
		},
		{
			"query denom borrowed",
			cli.GetCmdQueryBorrowed(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QueryBorrowedResponse{},
			&types.QueryBorrowedResponse{
				Borrowed: sdk.NewCoins(
					sdk.NewInt64Coin(app.BondDenom, 50),
				),
			},
		},
	}

	cleanupCommands := []testTransaction{
		{
			"repay",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				"50uumee",
			},
			nil,
		},
		{
			"unset collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"false",
			},
			nil,
		},
		{
			"withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"1000u/uumee",
			},
			nil,
		},
	}

	runTestTransactions(s, setupCommands)
	runTestQueries(s, testCases)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestQueryReserveAmount() {
	testCases := []testQuery{
		{
			"query reserve amount",
			cli.GetCmdQueryReserveAmount(),
			[]string{
				app.BondDenom,
			},
			false,
			&types.QueryReserveAmountResponse{},
			&types.QueryReserveAmountResponse{
				Amount: sdk.ZeroInt(),
			},
		},
		{
			"invalid denom",
			cli.GetCmdQueryReserveAmount(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
	}

	runTestQueries(s, testCases)
}

func (s *IntegrationTestSuite) TestQueryCollateral() {
	val := s.network.Validators[0]

	setupCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uumee",
			},
			nil,
		},
		{
			"set collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"true",
			},
			nil,
		},
	}

	testCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryCollateral(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query all collateral",
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
		{
			"invalid denom",
			cli.GetCmdQueryCollateral(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=abcd", cli.FlagDenom),
			},
			true,
			nil,
			nil,
		},
		{
			"query denom collateral",
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
	}

	cleanupCommands := []testTransaction{
		{
			"unset collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"false",
			},
			nil,
		},
		{
			"withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"1000u/uumee",
			},
			nil,
		},
	}

	runTestTransactions(s, setupCommands)
	runTestQueries(s, testCases)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestQueryCollateralSetting() {
	val := s.network.Validators[0]

	setupCommands := []testTransaction{
		{
			"set collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"true",
			},
			nil,
		},
	}

	testCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryCollateralSetting(),
			[]string{
				"xyz",
				"u/uumee",
			},
			true,
			nil,
			nil,
		},
		{
			"invalid denom",
			cli.GetCmdQueryCollateralSetting(),
			[]string{
				val.Address.String(),
				"abcd",
			},
			true,
			nil,
			nil,
		},
		{
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
	}

	cleanupCommands := []testTransaction{
		{
			"unset collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"false",
			},
			nil,
		},
	}

	runTestTransactions(s, setupCommands)
	runTestQueries(s, testCases)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestQueryExchangeRate() {
	testCases := []testQuery{
		{
			"invalid denom",
			cli.GetCmdQueryExchangeRate(),
			[]string{
				"abcd",
			},
			true,
			nil,
			nil,
		},
		{
			"query exchange rate",
			cli.GetCmdQueryExchangeRate(),
			[]string{
				app.BondDenom,
			},
			false,
			&types.QueryExchangeRateResponse{},
			&types.QueryExchangeRateResponse{
				ExchangeRate: sdk.OneDec(),
			},
		},
	}

	runTestQueries(s, testCases)
}

func (s *IntegrationTestSuite) TestQueryBorrowLimit() {
	val := s.network.Validators[0]

	simpleCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryBorrowLimit(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query zero borrow limit",
			cli.GetCmdQueryBorrowLimit(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryBorrowLimitResponse{},
			&types.QueryBorrowLimitResponse{
				BorrowLimit: sdk.ZeroDec(),
			},
		},
	}

	setupCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"20000000uumee", // 20 umee
			},
			nil,
		},
		{
			"set collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"true",
			},
			nil,
		},
	}

	nonzeroCase := []testQuery{
		{
			"query nonzero borrow limit",
			cli.GetCmdQueryBorrowLimit(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryBorrowLimitResponse{},
			&types.QueryBorrowLimitResponse{
				// From app/beta/test_helpers.go/IntegrationTestNetworkConfig
				// This result is umee's collateral weight times the collateral
				// amount lent, times its initial oracle exchange rate.
				BorrowLimit: sdk.MustNewDecFromStr("34.21"),
				// 0.05 * 20 * 34.21 = 34.21
			},
		},
	}

	cleanupCommands := []testTransaction{
		{
			"unset collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"false",
			},
			nil,
		},
		{
			"withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"20000000u/uumee",
			},
			nil,
		},
	}

	runTestQueries(s, simpleCases)
	runTestTransactions(s, setupCommands)
	runTestQueries(s, nonzeroCase)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestQueryLendAPY() {
	testCasesLendAPY := []testQuery{
		{
			"not accepted Token denom",
			cli.GetCmdQueryLendAPY(),
			[]string{
				"invalidToken",
			},
			true,
			nil,
			nil,
		},
		{
			"invalid denom",
			cli.GetCmdQueryLendAPY(),
			[]string{
				"",
			},
			true,
			nil,
			nil,
		},
		{
			"valid asset",
			cli.GetCmdQueryLendAPY(),
			[]string{
				app.BondDenom,
			},
			false,
			&types.QueryLendAPYResponse{},
			// Borrow rate * (1 - LiquidationIncentive - OracleRewardFactor)
			// 1.50 * (1 - 0.18 - 0.01) = 1.215
			&types.QueryLendAPYResponse{APY: sdk.MustNewDecFromStr("1.215")},
		},
	}
	runTestQueries(s, testCasesLendAPY)
}

func (s *IntegrationTestSuite) TestQueryBorrowAPY() {
	testCasesBorrowAPY := []testQuery{
		{
			"not accepted Token denom",
			cli.GetCmdQueryBorrowAPY(),
			[]string{
				"invalidToken",
			},
			true,
			nil,
			nil,
		},
		{
			"invalid denom",
			cli.GetCmdQueryBorrowAPY(),
			[]string{
				"",
			},
			true,
			nil,
			nil,
		},
		{
			"valid asset",
			cli.GetCmdQueryBorrowAPY(),
			[]string{
				app.BondDenom,
			},
			false,
			&types.QueryBorrowAPYResponse{},
			// This is an edge case technically - when effective supply, meaning
			// module balance + total borrows, is zero, utilization (0/0) is
			// interpreted as 100% so max borrow rate (150% APY) is used.
			&types.QueryBorrowAPYResponse{APY: sdk.MustNewDecFromStr("1.50")},
		},
	}
	runTestQueries(s, testCasesBorrowAPY)
}

func (s *IntegrationTestSuite) TestQueryMarketSize() {
	testCasesMarketSizeBeforeLend := []testQuery{
		{
			"not accepted Token denom",
			cli.GetCmdQueryMarketSize(),
			[]string{
				"invalidToken",
			},
			true,
			nil,
			nil,
		},
		{
			"invalid denom",
			cli.GetCmdQueryMarketSize(),
			[]string{
				"",
			},
			true,
			nil,
			nil,
		},
		{
			"valid asset",
			cli.GetCmdQueryMarketSize(),
			[]string{
				app.BondDenom,
			},
			false,
			&types.QueryMarketSizeResponse{},
			&types.QueryMarketSizeResponse{MarketSizeUsd: sdk.ZeroDec()},
		},
	}

	val := s.network.Validators[0]
	lendCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000000uumee",
			},
			nil,
		},
	}

	testCasesMarketSizeAfterLend := []testQuery{
		{
			"valid asset",
			cli.GetCmdQueryMarketSize(),
			[]string{
				app.BondDenom,
			},
			false,
			&types.QueryMarketSizeResponse{},
			&types.QueryMarketSizeResponse{MarketSizeUsd: sdk.MustNewDecFromStr("34.21")},
		},
	}

	cleanupCommands := []testTransaction{
		{
			"withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"1000000u/uumee",
			},
			nil,
		},
	}

	runTestQueries(s, testCasesMarketSizeBeforeLend)
	runTestTransactions(s, lendCommands)
	runTestQueries(s, testCasesMarketSizeAfterLend)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestCmdLend() {
	val := s.network.Validators[0]

	testCases := []testTransaction{
		{
			"invalid asset",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uabcd",
			},
			types.ErrInvalidAsset,
		},
		{
			"insufficient funds",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"10000000000000000uumee",
			},
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"valid lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uumee",
			},
			nil,
		},
	}

	cleanupCommands := []testTransaction{
		{
			"withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"1000u/uumee",
			},
			nil,
		},
	}

	runTestTransactions(s, testCases)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestCmdWithdraw() {
	val := s.network.Validators[0]

	setupCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uumee",
			},
			nil,
		},
	}

	testCases := []testTransaction{
		{
			"invalid asset",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"1000uabcd",
			},
			types.ErrInvalidAsset,
		},
		{
			"valid withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"1000u/uumee",
			},
			nil,
		},
		{
			"lending pool insufficient",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"10000000u/uumee",
			},
			types.ErrLendingPoolInsufficient,
		},
	}

	runTestTransactions(s, setupCommands)
	runTestTransactions(s, testCases)
}

func (s *IntegrationTestSuite) TestCmdBorrow() {
	val := s.network.Validators[0]

	setupCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uumee",
			},
			nil,
		},
		{
			"set collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"true",
			},
			nil,
		},
	}

	testCases := []testTransaction{
		{
			"invalid asset",
			cli.GetCmdBorrowAsset(),
			[]string{
				val.Address.String(),
				"10xyz",
			},
			types.ErrInvalidAsset,
		},
		{
			"invalid asset (uToken)",
			cli.GetCmdBorrowAsset(),
			[]string{
				val.Address.String(),
				"10u/umee",
			},
			types.ErrInvalidAsset,
		},
		{
			"lending pool insufficient",
			cli.GetCmdBorrowAsset(),
			[]string{
				val.Address.String(),
				"7000uumee",
			},
			types.ErrLendingPoolInsufficient,
		},
		{
			"borrow limit low",
			cli.GetCmdBorrowAsset(),
			[]string{
				val.Address.String(),
				"70uumee",
			},
			types.ErrBorrowLimitLow,
		},
		{
			"borrow",
			cli.GetCmdBorrowAsset(),
			[]string{
				val.Address.String(),
				"50uumee",
			},
			nil,
		},
	}

	cleanupCommands := []testTransaction{
		{
			"repay",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				"50uumee",
			},
			nil,
		},
		{
			"unset collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"false",
			},
			nil,
		},
		{
			"withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"1000u/uumee",
			},
			nil,
		},
	}

	runTestTransactions(s, setupCommands)
	runTestTransactions(s, testCases)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestCmdAvailableBorrow() {
	testCasesAvailableBorrowBeforeLend := []testQuery{
		{
			"not accepted Token denom",
			cli.GetCmdQueryAvailableBorrow(),
			[]string{
				"invalidToken",
			},
			true,
			nil,
			nil,
		},
		{
			"invalid denom",
			cli.GetCmdQueryAvailableBorrow(),
			[]string{
				"",
			},
			true,
			nil,
			nil,
		},
		{
			"valid asset",
			cli.GetCmdQueryAvailableBorrow(),
			[]string{
				app.BondDenom,
			},
			false,
			&types.QueryAvailableBorrowResponse{},
			&types.QueryAvailableBorrowResponse{Amount: sdk.ZeroInt()},
		},
	}

	val := s.network.Validators[0]
	amountLend := int64(1000000)
	lendCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("%duumee", amountLend),
			},
			nil,
		},
	}

	testCasesAvailableBorrowAfterLend := []testQuery{
		{
			"valid asset",
			cli.GetCmdQueryAvailableBorrow(),
			[]string{
				app.BondDenom,
			},
			false,
			&types.QueryAvailableBorrowResponse{},
			&types.QueryAvailableBorrowResponse{Amount: sdk.NewInt(amountLend)},
		},
	}

	cleanupCommands := []testTransaction{
		{
			"withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("%du/uumee", amountLend),
			},
			nil,
		},
	}

	runTestQueries(s, testCasesAvailableBorrowBeforeLend)
	runTestTransactions(s, lendCommands)
	runTestQueries(s, testCasesAvailableBorrowAfterLend)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestCmdRepay() {
	val := s.network.Validators[0]

	setupCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uumee",
			},
			nil,
		},
		{
			"set collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"true",
			},
			nil,
		},
		{
			"borrow",
			cli.GetCmdBorrowAsset(),
			[]string{
				val.Address.String(),
				"50uumee",
			},
			nil,
		},
	}

	testCases := []testTransaction{
		{
			"invalid asset",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				"10xyz",
			},
			types.ErrInvalidAsset,
		},
		{
			"invalid asset (uToken)",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				"10u/umee",
			},
			types.ErrInvalidAsset,
		},
		{
			"repay",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				"50uumee",
			},
			nil,
		},
	}

	cleanupCommands := []testTransaction{
		{
			"unset collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"false",
			},
			nil,
		},
		{
			"withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"1000u/uumee",
			},
			nil,
		},
	}

	runTestTransactions(s, setupCommands)
	runTestTransactions(s, testCases)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestCmdLiquidate() {
	val := s.network.Validators[0]

	noTargetsQuery := []testQuery{
		{
			"no targets",
			cli.GetCmdQueryLiquidationTargets(),
			[]string{},
			false,
			&types.QueryLiquidationTargetsResponse{},
			&types.QueryLiquidationTargetsResponse{
				Targets: []string{},
			},
		},
	}

	setupCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uumee",
			},
			nil,
		},
		{
			"set collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"true",
			},
			nil,
		},
		{
			"borrow",
			cli.GetCmdBorrowAsset(),
			[]string{
				val.Address.String(),
				"50uumee",
			},
			nil,
		},
	}

	oneTargetQuery := []testQuery{
		{
			"one liquidation target",
			cli.GetCmdQueryLiquidationTargets(),
			[]string{},
			false,
			&types.QueryLiquidationTargetsResponse{},
			&types.QueryLiquidationTargetsResponse{
				Targets: []string{
					val.Address.String(),
				},
			},
		},
	}

	testCases := []testTransaction{
		{
			"valid liquidate",
			cli.GetCmdLiquidate(),
			[]string{
				val.Address.String(),
				val.Address.String(),
				// note: partial liquidation, so cleanup still requires a CmdRepay
				"5uumee",
				"u/uumee",
			},
			nil,
		},
	}

	cleanupCommands := []testTransaction{
		{
			"repay",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				// note: amount will be reduced from 50 to remaining amount owed
				"50uumee",
			},
			nil,
		},
		{
			"unset collateral",
			cli.GetCmdSetCollateral(),
			[]string{
				val.Address.String(),
				"u/uumee",
				"false",
			},
			nil,
		},
		{
			"withdraw",
			cli.GetCmdWithdrawAsset(),
			[]string{
				val.Address.String(),
				"1000u/uumee",
			},
			nil,
		},
	}

	runTestQueries(s, noTargetsQuery)
	runTestTransactions(s, setupCommands)
	updateCollateralWeight(s, "uumee", sdk.MustNewDecFromStr("0.01")) // lower to allow liquidation
	runTestQueries(s, oneTargetQuery)
	runTestTransactions(s, testCases)
	updateCollateralWeight(s, "uumee", sdk.MustNewDecFromStr("0.05")) // reset to original
	runTestTransactions(s, cleanupCommands)
}
