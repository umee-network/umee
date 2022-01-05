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
	//todo: discover why collat weight isn't updating. the following line should break this test but doesn't
	originalCollateralWeight := updateCollateralWeight(s, "uumee", sdk.MustNewDecFromStr("0.10"))
	runTestQueries(s, nonzeroCase)
	_ = updateCollateralWeight(s, "uumee", originalCollateralWeight)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestQueryLiquidationTargets() {
	testCase := []testQuery{
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

	runTestQueries(s, testCase)
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
			&types.QueryBorrowAPYResponse{APY: sdk.ZeroDec()},
		},
	}
	runTestQueries(s, testCasesBorrowAPY)
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

/*
func (s *IntegrationTestSuite) TestCmdLiquidate() {
	val := s.network.Validators[0]

	setupCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"100000uumee",
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
				"5000uumee",
			},
			nil,
		},
	}

	testCases := []testTransaction{
		{
			"valid liquidate",
			cli.GetCmdLiquidate(),
			[]string{
				val.Address.String(),
				val.Address.String(),
				// note: liquidation amount will be automatically reduced to maximum eligible amount
				"5000uumee",
				"u/uumee",
			},
			nil,
		},
		{
			"liquidation ineligible",
			cli.GetCmdLiquidate(),
			[]string{
				val.Address.String(),
				val.Address.String(),
				"500uumee",
				"u/uumee",
			},
			types.ErrLiquidationIneligible,
		},
	}

	cleanupCommands := []testTransaction{
		{
			"repay",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				// note: repay amount will be automatically reduced post-liquidation
				"5000uumee",
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
				"100000u/uumee",
			},
			nil,
		},
	}

	runTestTransactions(s, setupCommands)
	originalCollateralWeight := updateCollateralWeight(s, "uumee", sdk.MustNewDecFromStr("0.00"))
	runTestTransactions(s, testCases)
	_ = updateCollateralWeight(s, "uumee", originalCollateralWeight)
	runTestTransactions(s, cleanupCommands)

}
*/
