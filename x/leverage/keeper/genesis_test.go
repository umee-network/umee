package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/leverage/types"
	"gotest.tools/v3/assert"
)

const (
	testAddr = "umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm"
	denom    = "umee"
	uDenom   = "u/umee"
)

func (s *IntegrationTestSuite) TestKeeper_InitGenesis() {
	tcs := []struct {
		name      string
		g         types.GenesisState
		expectErr bool
		errMsg    string
	}{
		{
			"invalid token",
			types.GenesisState{
				Params: types.DefaultParams(),
				Registry: []types.Token{
					{},
				},
			},
			true,
			"base_denom: invalid denom: ",
		},
		{
			"invalid address for borrow",
			types.GenesisState{
				Params: types.DefaultParams(),
				AdjustedBorrows: []types.AdjustedBorrow{
					{
						Address: "",
					},
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"invalid coin for borrow",
			types.GenesisState{
				Params: types.DefaultParams(),
				AdjustedBorrows: []types.AdjustedBorrow{
					{
						Address: testAddr,
						Amount:  sdk.DecCoin{},
					},
				},
			},
			true,
			"invalid denom: ",
		},
		{
			"invalid address for collateral",
			types.GenesisState{
				Params: types.DefaultParams(),
				Collateral: []types.Collateral{
					{
						Address: "",
					},
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"invalid coin for collateral",
			types.GenesisState{
				Params: types.DefaultParams(),
				Collateral: []types.Collateral{
					{
						Address: testAddr,
						Amount:  sdk.Coin{},
					},
				},
			},
			true,
			"invalid denom: ",
		},
		{
			"invalid coin for reserves",
			types.GenesisState{
				Params: types.DefaultParams(),
				Reserves: sdk.Coins{
					sdk.Coin{
						Denom: "",
					},
				},
			},
			true,
			"invalid denom: ",
		},
		{
			"invalid address for badDebt",
			types.GenesisState{
				Params: types.DefaultParams(),
				BadDebts: []types.BadDebt{
					{
						Address: "",
					},
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"invalid coin for badDebt",
			types.GenesisState{
				Params: types.DefaultParams(),
				BadDebts: []types.BadDebt{
					{
						Address: testAddr,
						Denom:   "",
					},
				},
			},
			true,
			"invalid denom: ",
		},
		{
			"invalid coin for interestScalars",
			types.GenesisState{
				Params: types.DefaultParams(),
				InterestScalars: []types.InterestScalar{
					{
						Denom: "",
					},
				},
			},
			true,
			"invalid denom: ",
		},
		{
			"valid",
			types.GenesisState{
				Params: types.DefaultParams(),
			},
			false,
			"",
		},
	}

	for _, tc := range tcs {
		s.Run(
			tc.name, func() {
				if tc.expectErr {
					s.PanicsWithError(tc.errMsg, func() { s.app.LeverageKeeper.InitGenesis(s.ctx, tc.g) })
				} else {
					s.NotPanics(func() { s.app.LeverageKeeper.InitGenesis(s.ctx, tc.g) })
				}
			},
		)
	}
}

func (s *IntegrationTestSuite) TestKeeper_ExportGenesis() {
	borrows := []types.AdjustedBorrow{
		{
			Address: testAddr,
			Amount:  sdk.NewDecCoin(denom, sdk.NewInt(100)),
		},
	}
	collateral := []types.Collateral{
		{
			Address: testAddr,
			Amount:  sdk.NewCoin(uDenom, sdk.NewInt(1000)),
		},
	}
	reserves := sdk.Coins{
		sdk.NewCoin(denom, sdkmath.NewInt(10)),
	}
	badDebts := []types.BadDebt{
		{
			Address: testAddr,
			Denom:   denom,
		},
	}
	interestScalars := []types.InterestScalar{
		{
			Denom:  denom,
			Scalar: sdk.NewDec(10),
		},
	}
	genesis := types.DefaultGenesis()
	genesis.AdjustedBorrows = borrows
	genesis.Collateral = collateral
	genesis.Reserves = reserves
	genesis.BadDebts = badDebts
	genesis.InterestScalars = interestScalars
	s.app.LeverageKeeper.InitGenesis(s.ctx, *genesis)

	export := s.app.LeverageKeeper.ExportGenesis(s.ctx)
	assert.DeepEqual(s.T(), borrows, export.AdjustedBorrows)
	assert.DeepEqual(s.T(), collateral, export.Collateral)
	assert.DeepEqual(s.T(), reserves, export.Reserves)
	assert.DeepEqual(s.T(), badDebts, export.BadDebts)
	assert.DeepEqual(s.T(), interestScalars, export.InterestScalars)
}
