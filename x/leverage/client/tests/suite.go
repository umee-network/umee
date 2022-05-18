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

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/leverage/client/cli"
	"github.com/umee-network/umee/v2/x/leverage/types"
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
				s.Require().Equal(0, int(resp.Code), "events %v", resp.Events)
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
						EnableLend:           true,
						EnableBorrow:         true,
						Blacklist:            false,
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

	// 51 borrowed will be returned from query due to adjusted borrow rounding up
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
					sdk.NewInt64Coin(umeeapp.BondDenom, 51),
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
					sdk.NewInt64Coin(umeeapp.BondDenom, 51),
				),
			},
		},
	}

	// 51 will need to be repaid due to adjusted borrow rounding up
	cleanupCommands := []testTransaction{
		{
			"repay",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				"51uumee",
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

func (s *IntegrationTestSuite) TestQueryBorrowedValue() {
	val := s.network.Validators[0]

	simpleCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryBorrowedValue(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query zero borrowed value",
			cli.GetCmdQueryBorrowedValue(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryBorrowedValueResponse{},
			&types.QueryBorrowedValueResponse{
				BorrowedValue: sdk.ZeroDec(),
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
				"49uumee", // will round up to 50 due to adjusted borrow value
			},
			nil,
		},
	}

	nonzeroCase := []testQuery{
		{
			"query nonzero borrowed value",
			cli.GetCmdQueryBorrowedValue(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryBorrowedValueResponse{},
			&types.QueryBorrowedValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				BorrowedValue: sdk.MustNewDecFromStr("0.0017105"),
				// (50 / 1000000) umee * 34.21 = 0.0017105
			},
		},
		{
			"query nonzero borrowed value of denom",
			cli.GetCmdQueryBorrowedValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QueryBorrowedValueResponse{},
			&types.QueryBorrowedValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				BorrowedValue: sdk.MustNewDecFromStr("0.0017105"),
				// (50 / 1000000) umee * 34.21 = 0.0017105
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

	runTestQueries(s, simpleCases)
	runTestTransactions(s, setupCommands)
	runTestQueries(s, nonzeroCase)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestQueryLoaned() {
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

	testCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryLoaned(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query all loaned",
			cli.GetCmdQueryLoaned(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryLoanedResponse{},
			&types.QueryLoanedResponse{
				Loaned: sdk.NewCoins(
					sdk.NewInt64Coin(umeeapp.BondDenom, 1000),
				),
			},
		},
		{
			"invalid denom",
			cli.GetCmdQueryLoaned(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=abcd", cli.FlagDenom),
			},
			true,
			nil,
			nil,
		},
		{
			"query denom loaned",
			cli.GetCmdQueryLoaned(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QueryLoanedResponse{},
			&types.QueryLoanedResponse{
				Loaned: sdk.NewCoins(
					sdk.NewInt64Coin(umeeapp.BondDenom, 1000),
				),
			},
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

	runTestTransactions(s, setupCommands)
	runTestQueries(s, testCases)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestQueryLoanedValue() {
	val := s.network.Validators[0]

	simpleCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryLoanedValue(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query zero loaned value",
			cli.GetCmdQueryLoanedValue(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryLoanedValueResponse{},
			&types.QueryLoanedValueResponse{
				LoanedValue: sdk.ZeroDec(),
			},
		},
	}

	setupCommands := []testTransaction{
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

	testCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryLoanedValue(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query all loaned value",
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
				LoanedValue: sdk.MustNewDecFromStr("34.21"),
				// 1 umee * 34.21 = 34.21
			},
		},
		{
			"query denom loaned",
			cli.GetCmdQueryLoanedValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=uumee", cli.FlagDenom),
			},
			false,
			&types.QueryLoanedValueResponse{},
			&types.QueryLoanedValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				LoanedValue: sdk.MustNewDecFromStr("34.21"),
				// 1 umee * 34.21 = 34.21
			},
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

	runTestQueries(s, simpleCases)
	runTestTransactions(s, setupCommands)
	runTestQueries(s, testCases)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestQueryCollateralValue() {
	val := s.network.Validators[0]

	simpleCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryCollateralValue(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query zero collateral value",
			cli.GetCmdQueryCollateralValue(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryCollateralValueResponse{},
			&types.QueryCollateralValueResponse{
				CollateralValue: sdk.ZeroDec(),
			},
		},
	}

	setupCommands := []testTransaction{
		{
			"lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000000uumee",
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
			cli.GetCmdQueryCollateralValue(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query all collateral value",
			cli.GetCmdQueryCollateralValue(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryCollateralValueResponse{},
			&types.QueryCollateralValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				// This result is umee's oracle exchange rate times the
				// amount set as collateral.
				CollateralValue: sdk.MustNewDecFromStr("34.21"),
				// 1 umee * 34.21 = 34.21
			},
		},
		{
			"query denom collateral value",
			cli.GetCmdQueryCollateralValue(),
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=u/uumee", cli.FlagDenom),
			},
			false,
			&types.QueryCollateralValueResponse{},
			&types.QueryCollateralValueResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				CollateralValue: sdk.MustNewDecFromStr("34.21"),
				// 1 umee * 34.21 = 34.21
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
				"1000000u/uumee",
			},
			nil,
		},
	}

	runTestQueries(s, simpleCases)
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
				umeeapp.BondDenom,
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
				umeeapp.BondDenom,
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
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				// This result is umee's collateral weight times the collateral
				// amount loaned, times its initial oracle exchange rate.
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

func (s *IntegrationTestSuite) TestQueryLiquidationThreshold() {
	val := s.network.Validators[0]

	simpleCases := []testQuery{
		{
			"invalid address",
			cli.GetCmdQueryLiquidationThreshold(),
			[]string{
				"xyz",
			},
			true,
			nil,
			nil,
		},
		{
			"query zero liquidation threshold",
			cli.GetCmdQueryLiquidationThreshold(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryLiquidationThresholdResponse{},
			&types.QueryLiquidationThresholdResponse{
				LiquidationThreshold: sdk.ZeroDec(),
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
			"query nonzero liquidation threshold",
			cli.GetCmdQueryLiquidationThreshold(),
			[]string{
				val.Address.String(),
			},
			false,
			&types.QueryLiquidationThresholdResponse{},
			&types.QueryLiquidationThresholdResponse{
				// From app/test_helpers.go/IntegrationTestNetworkConfig
				// This result is umee's liquidation threshold times the collateral
				// amount loaned, times its initial oracle exchange rate.
				LiquidationThreshold: sdk.MustNewDecFromStr("34.21"),
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
				umeeapp.BondDenom,
			},
			false,
			&types.QueryLendAPYResponse{},
			// Borrow rate * (1 - ReserveFactor - OracleRewardFactor)
			// 1.50 * (1 - 0.10 - 0.01) = 0.89 * 1.5 = 1.335
			&types.QueryLendAPYResponse{APY: sdk.MustNewDecFromStr("1.335")},
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
	runTestQueries(s, testCasesBorrowAPY)
}

func (s *IntegrationTestSuite) TestQueryMarketSize() {
	simpleCases := []testQuery{
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
				umeeapp.BondDenom,
			},
			false,
			&types.QueryMarketSizeResponse{},
			&types.QueryMarketSizeResponse{MarketSizeUsd: sdk.ZeroDec()},
		},
	}

	val := s.network.Validators[0]
	setupCommands := []testTransaction{
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

	testCases := []testQuery{
		{
			"valid asset",
			cli.GetCmdQueryMarketSize(),
			[]string{
				umeeapp.BondDenom,
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

	runTestQueries(s, simpleCases)
	runTestTransactions(s, setupCommands)
	runTestQueries(s, testCases)
	runTestTransactions(s, cleanupCommands)
}

func (s *IntegrationTestSuite) TestQueryTokenMarketSize() {
	simpleCases := []testQuery{
		{
			"not accepted Token denom",
			cli.GetCmdQueryTokenMarketSize(),
			[]string{
				"invalidToken",
			},
			true,
			nil,
			nil,
		},
		{
			"invalid denom",
			cli.GetCmdQueryTokenMarketSize(),
			[]string{
				"",
			},
			true,
			nil,
			nil,
		},
		{
			"valid asset",
			cli.GetCmdQueryTokenMarketSize(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryTokenMarketSizeResponse{},
			&types.QueryTokenMarketSizeResponse{MarketSize: sdk.ZeroInt()},
		},
	}

	val := s.network.Validators[0]
	setupCommands := []testTransaction{
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

	testCases := []testQuery{
		{
			"valid asset",
			cli.GetCmdQueryTokenMarketSize(),
			[]string{
				umeeapp.BondDenom,
			},
			false,
			&types.QueryTokenMarketSizeResponse{},
			&types.QueryTokenMarketSizeResponse{MarketSize: sdk.NewInt(1000000)},
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

	runTestQueries(s, simpleCases)
	runTestTransactions(s, setupCommands)
	runTestQueries(s, testCases)
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
			types.ErrUndercollaterized,
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

	// 51 will need to be repaid due to adjusted borrow rounding up
	cleanupCommands := []testTransaction{
		{
			"repay",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				"51uumee",
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
				umeeapp.BondDenom,
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
				umeeapp.BondDenom,
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
			"nothing to repay",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				"10xyz",
			},
			types.ErrInvalidRepayment,
		},
		{
			"nothing to repay (uToken)",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				"10u/umee",
			},
			types.ErrInvalidRepayment,
		},
		{
			"repay",
			cli.GetCmdRepayAsset(),
			[]string{
				val.Address.String(),
				// 51 will need to be repaid due to adjusted borrow rounding up
				"51uumee",
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
				"0uumee",
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
