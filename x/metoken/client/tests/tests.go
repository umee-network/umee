package tests

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	itestsuite "github.com/umee-network/umee/v6/tests/cli"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/client/cli"
	mfixtures "github.com/umee-network/umee/v6/x/metoken/mocks"
)

func (s *IntegrationTests) TestInvalidQueries() {
	invalidQueries := []itestsuite.TestQuery{
		{

			Name:    "query swap fee - invalid asset for swap",
			Command: cli.GetCmdSwapFee(),
			Args: []string{
				"{abcd}{100000000}",
				"xyz",
			},
			Response:         nil,
			ExpectedResponse: nil,
			ErrMsg:           "invalid decimal coin expression",
		},
		{
			Name:    "query swap fee - index not found",
			Command: cli.GetCmdSwapFee(),
			Args: []string{
				"1000abcd",
				"xyz",
			},
			Response:         nil,
			ExpectedResponse: nil,
			ErrMsg:           "index xyz not found",
		},
		{
			Name:    "query redeem fee - invalid meToken for redemption",
			Command: cli.GetCmdRedeemFee(),
			Args: []string{
				"{abcd}{100000000}",
				"xyz",
			},
			Response:         nil,
			ExpectedResponse: nil,
			ErrMsg:           "invalid decimal coin expression",
		},
		{
			Name:    "query redeem fee - index not found",
			Command: cli.GetCmdRedeemFee(),
			Args: []string{
				"1000xyz",
				"abcd",
			},
			Response:         nil,
			ExpectedResponse: nil,
			ErrMsg:           "index xyz not found",
		},
	}

	// These queries do not require any setup because they contain invalid arguments
	s.RunTestQueries(invalidQueries...)
}

func (s *IntegrationTests) TestValidQueries() {
	queries := []itestsuite.TestQuery{
		{
			Name:     "query params",
			Command:  cli.GetCmdQueryParams(),
			Args:     []string{},
			Response: &metoken.QueryParamsResponse{},
			ExpectedResponse: &metoken.QueryParamsResponse{
				Params: metoken.DefaultParams(),
			},
			ErrMsg: "",
		},
		{
			Name:     "query indexes",
			Command:  cli.GetCmdIndexes(),
			Args:     []string{},
			Response: &metoken.QueryIndexesResponse{},
			ExpectedResponse: &metoken.QueryIndexesResponse{
				Registry: []metoken.Index{mfixtures.BondIndex()},
			},
			ErrMsg: "",
		},
		{
			Name:     "query balances",
			Command:  cli.GetCmdIndexBalances(),
			Args:     []string{},
			Response: &metoken.QueryIndexBalancesResponse{},
			ExpectedResponse: &metoken.QueryIndexBalancesResponse{
				IndexBalances: []metoken.IndexBalances{mfixtures.BondBalance()},
				Prices: []metoken.IndexPrices{
					{
						Denom:    mfixtures.MeBondDenom,
						Price:    sdk.MustNewDecFromStr("34.21"),
						Exponent: 6,
						Assets: []metoken.AssetPrice{
							{
								BaseDenom:   mfixtures.BondDenom,
								SymbolDenom: "UMEE",
								Price:       sdk.MustNewDecFromStr("34.21"),
								Exponent:    6,
								SwapRate:    sdk.OneDec(),
								RedeemRate:  sdk.OneDec(),
							},
						},
					},
				},
			},
			ErrMsg: "",
		},
		{
			Name:    "query swap fee for 1876 uumee",
			Command: cli.GetCmdSwapFee(),
			Args: []string{
				"1876000000uumee",
				mfixtures.MeBondDenom,
			},
			Response: &metoken.QuerySwapFeeResponse{},
			ExpectedResponse: &metoken.QuerySwapFeeResponse{
				// swap_fee = 0.01 * 1876_000000 = 18760000
				Asset: sdk.NewCoin(
					"uumee",
					sdkmath.NewInt(18_760000),
				),
			},
			ErrMsg: "",
		},
		{
			Name:    "query redeem fee for 100 meUSD to uumee",
			Command: cli.GetCmdRedeemFee(),
			Args: []string{
				"100000000me/uumee",
				"uumee",
			},
			Response: &metoken.QueryRedeemFeeResponse{},
			ExpectedResponse: &metoken.QueryRedeemFeeResponse{
				// with all balances in 0
				// current_allocation = 0
				// redeem_delta_allocation = target_allocation - current_allocation
				// redeem_delta_allocation = 1.0 - 0 = 1.0
				// fee = redeem_delta_allocation * balanced_fee + balanced_fee
				// fee = 1.0 * 0.2 + 0.2 = 0.4
				// exchange_rate = 1
				// asset_to_redeem = exchange_rate * metoken_amount
				// asset_to_redeem = 1 * 100_000000 = 100_000000
				// total_fee = asset_to_redeem * fee
				// total_fee = 100_000000 * 0.4 = 40_000000

				Asset: sdk.NewCoin(
					"uumee",
					sdkmath.NewInt(40_000000),
				),
			},
			ErrMsg: "",
		},
	}

	// These queries do not require any setup
	s.RunTestQueries(queries...)
}
