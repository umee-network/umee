package tests

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	appparams "github.com/umee-network/umee/v6/app/params"
	itestsuite "github.com/umee-network/umee/v6/tests/cli"
	"github.com/umee-network/umee/v6/x/incentive"
	"github.com/umee-network/umee/v6/x/incentive/client/cli"
	leveragecli "github.com/umee-network/umee/v6/x/leverage/client/cli"
)

func (s *IntegrationTests) TestInvalidQueries() {
	invalidQueries := []itestsuite.TestQuery{
		{
			Name:    "query pending rewards (invalid address)",
			Command: cli.QueryPendingRewards(),
			Args: []string{
				"xyz",
			},
			Response:         nil,
			ExpectedResponse: nil,
			ErrMsg:           "invalid bech32",
		},
		{
			Name:    "query current rates (not uToken)",
			Command: cli.QueryCurrentRates(),
			Args: []string{
				"uumee",
			},
			Response:         nil,
			ExpectedResponse: nil,
			ErrMsg:           "denom should be a uToken",
		},
		{
			Name:    "query actual rates (not uToken)",
			Command: cli.QueryActualRates(),
			Args: []string{
				"uumee",
			},
			Response:         nil,
			ExpectedResponse: nil,
			ErrMsg:           "denom should be a uToken",
		},
	}

	s.RunTestQueries(invalidQueries...)
}

func (s *IntegrationTests) TestIncentiveScenario() {
	val := s.Network.Validators[0]

	zeroQueries := []itestsuite.TestQuery{
		{
			Name:     "query params",
			Command:  cli.QueryParams(),
			Args:     []string{},
			Response: &incentive.QueryParamsResponse{},
			ExpectedResponse: &incentive.QueryParamsResponse{
				Params: incentive.DefaultParams(),
			},
			ErrMsg: "",
		},
		{
			Name:     "query upcoming incentive programs",
			Command:  cli.QueryUpcomingIncentivePrograms(),
			Args:     []string{},
			Response: &incentive.QueryUpcomingIncentiveProgramsResponse{},
			ExpectedResponse: &incentive.QueryUpcomingIncentiveProgramsResponse{
				Programs: []incentive.IncentiveProgram{},
			},
			ErrMsg: "",
		},
		{
			Name:     "query ongoing incentive programs",
			Command:  cli.QueryUpcomingIncentivePrograms(),
			Args:     []string{},
			Response: &incentive.QueryOngoingIncentiveProgramsResponse{},
			ExpectedResponse: &incentive.QueryOngoingIncentiveProgramsResponse{
				Programs: []incentive.IncentiveProgram{},
			},
			ErrMsg: "",
		},
		{
			Name:     "query completed incentive programs",
			Command:  cli.QueryCompletedIncentivePrograms(),
			Args:     []string{},
			Response: &incentive.QueryCompletedIncentiveProgramsResponse{},
			ExpectedResponse: &incentive.QueryCompletedIncentiveProgramsResponse{
				Programs: []incentive.IncentiveProgram{},
			},
			ErrMsg: "",
		},
		{
			Name:     "query total bonded - no denom",
			Command:  cli.QueryTotalBonded(),
			Args:     []string{},
			Response: &incentive.QueryTotalBondedResponse{},
			ExpectedResponse: &incentive.QueryTotalBondedResponse{
				Bonded: sdk.NewCoins(),
			},
		},
		{
			Name:    "query total bonded - specific denom",
			Command: cli.QueryTotalBonded(),
			Args: []string{
				"u/" + appparams.BondDenom,
			},
			Response: &incentive.QueryTotalBondedResponse{},
			ExpectedResponse: &incentive.QueryTotalBondedResponse{
				Bonded: sdk.NewCoins(),
			},
		},
		{
			Name:     "query total unbonding - no denom",
			Command:  cli.QueryTotalUnbonding(),
			Args:     []string{},
			Response: &incentive.QueryTotalUnbondingResponse{},
			ExpectedResponse: &incentive.QueryTotalUnbondingResponse{
				Unbonding: sdk.NewCoins(),
			},
		},
		{
			Name:    "query total unbonding - specific denom",
			Command: cli.QueryTotalUnbonding(),
			Args: []string{
				"u/" + appparams.BondDenom,
			},
			Response: &incentive.QueryTotalUnbondingResponse{},
			ExpectedResponse: &incentive.QueryTotalUnbondingResponse{
				Unbonding: sdk.NewCoins(),
			},
		},
		{
			Name:    "query current rates (zero)",
			Command: cli.QueryCurrentRates(),
			Args: []string{
				"u/" + appparams.BondDenom,
			},
			Response: &incentive.QueryCurrentRatesResponse{},
			ExpectedResponse: &incentive.QueryCurrentRatesResponse{
				ReferenceBond: sdk.NewInt64Coin("u/"+appparams.BondDenom, 1),
				Rewards:       sdk.NewCoins(),
			},
			ErrMsg: "",
		},
		{
			Name:    "query actual rates (zero)",
			Command: cli.QueryActualRates(),
			Args: []string{
				"u/" + appparams.BondDenom,
			},
			Response: &incentive.QueryActualRatesResponse{},
			ExpectedResponse: &incentive.QueryActualRatesResponse{
				APY: sdkmath.LegacyZeroDec(),
			},
			ErrMsg: "",
		},
		{
			Name:    "query pending rewards (zero)",
			Command: cli.QueryPendingRewards(),
			Args: []string{
				val.Address.String(),
			},
			Response: &incentive.QueryPendingRewardsResponse{},
			ExpectedResponse: &incentive.QueryPendingRewardsResponse{
				Rewards: sdk.NewCoins(),
			},
			ErrMsg: "",
		},
	}

	supplyCollateral := itestsuite.TestTransaction{
		Name:    "(setup) supply collateral",
		Command: leveragecli.SupplyCollateral(),
		Args: []string{
			"300uumee",
		},
		ExpectedErr: nil,
	}

	bond := itestsuite.TestTransaction{
		Name:    "bond",
		Command: cli.Bond(),
		Args: []string{
			"300u/uumee",
		},
		ExpectedErr: nil,
	}

	beingUnbonding := itestsuite.TestTransaction{
		Name:    "begin unbonding",
		Command: cli.BeginUnbonding(),
		Args: []string{
			"100u/uumee",
		},
		ExpectedErr: nil,
	}

	emergencyUnbond := itestsuite.TestTransaction{
		Name:    "emergency unbond",
		Command: cli.EmergencyUnbond(),
		Args: []string{
			"100u/uumee",
		},
		ExpectedErr: nil,
	}

	claim := itestsuite.TestTransaction{
		Name:        "claim",
		Command:     cli.Claim(),
		Args:        []string{},
		ExpectedErr: nil,
	}

	sponsor := itestsuite.TestTransaction{
		Name:        "sponsor (program does not exist)",
		Command:     cli.Sponsor(),
		Args:        []string{"1"},
		ExpectedErr: sdkerrors.ErrNotFound,
	}

	s.RunTestQueries(zeroQueries...)

	s.RunTestTransactions(
		supplyCollateral,
		bond,
		beingUnbonding,
		emergencyUnbond,
		claim,
		sponsor,
	)
}
