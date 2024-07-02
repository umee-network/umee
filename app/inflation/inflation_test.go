package inflation

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/golang/mock/gomock"
	"gotest.tools/v3/assert"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/tests/tsdk"
	"github.com/umee-network/umee/v6/util/bpmath"
	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/ugov"
)

func TestAdjustInflation(t *testing.T) {
	mintParams := minttypes.DefaultParams()

	tests := []struct {
		name           string
		totalSupply    sdkmath.Int
		maxSupply      sdkmath.Int
		minter         minttypes.Minter
		params         func(params minttypes.Params) minttypes.Params
		expectedResult sdkmath.LegacyDec
	}{
		{
			name:        "No inflation change => Newly Minted Coins + Total Supply is less than Max supply",
			totalSupply: sdkmath.NewInt(1000000),
			maxSupply:   sdkmath.NewInt(2000000),
			minter:      minttypes.Minter{Inflation: sdkmath.LegacyNewDecWithPrec(15, 2)},
			params: func(params minttypes.Params) minttypes.Params {
				/***
				Newly Minted Coins + Total Supply <= Max Supply
				=> No Inflation Rate change
				**/
				params.BlocksPerYear = 1
				return params
			},
			expectedResult: sdkmath.LegacyNewDecWithPrec(15, 2),
		},
		{
			name:        "Inflation Rate Adjust => Newly Minted Coins + Total Supply is greater than Max supply",
			totalSupply: sdkmath.NewInt(1900000),
			maxSupply:   sdkmath.NewInt(2000000),
			minter:      minttypes.Minter{Inflation: sdkmath.LegacyMustNewDecFromStr("7.1231")},
			params: func(params minttypes.Params) minttypes.Params {
				/***
				Newly Minted Coins + Total Supply >= Max Supply
				=> Inflation will be adjusted to reach Max Supply
				**/
				params.BlocksPerYear = 1
				return params
			},
			expectedResult: sdkmath.LegacyMustNewDecFromStr("0.052631578947368421"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calc := Calculator{}
			mintParams := test.params(mintParams)
			result := calc.AdjustInflation(test.totalSupply, test.maxSupply, test.minter, mintParams)

			if !result.Equal(test.expectedResult) {
				t.Errorf("Expected %s, but got %s", test.expectedResult, result)
			}
		})
	}
}

func TestInflationRate(t *testing.T) {
	mintParams := minttypes.DefaultParams()
	mockMinter := minttypes.NewMinter(sdkmath.LegacyMustNewDecFromStr("0.15"), sdkmath.LegacyNewDec(0))
	// mockInflationParams := ugov.InflationParams{
	// 	MaxSupply:              coin.New(appparams.BondDenom, 100000000),
	// 	InflationCycle:         time.Hour * 1,
	// 	InflationReductionRate: bpmath.FixedBP(2500),
	// }

	sdkContext, _ := tsdk.NewCtx(t, []storetypes.StoreKey{}, []storetypes.StoreKey{})

	tests := []struct {
		name            string
		totalSupply     sdkmath.Int
		minter          minttypes.Minter
		inflationParams func(ip ugov.InflationParams) ugov.InflationParams
		bondedRatio     sdkmath.LegacyDec
		mintParams      func(params minttypes.Params) minttypes.Params
		cycleEndTime    func() time.Time
		ctx             func() sdk.Context
		expectedResult  func(expectedResult, bondedRatio sdkmath.LegacyDec, mintParams minttypes.Params) sdkmath.LegacyDec
	}{
		{
			name:        "inflation rate change for min and max: new inflation cyle is started from this block time",
			totalSupply: sdkmath.NewInt(900000),
			minter:      mockMinter,
			mintParams: func(params minttypes.Params) minttypes.Params {
				// AnnualProvisions = 900000 * 0.15 = 135000
				mintParams.BlocksPerYear = 1
				mintParams.InflationRateChange = sdkmath.LegacyOneDec()
				return mintParams
			},
			inflationParams: func(ip ugov.InflationParams) ugov.InflationParams {
				return ip
			},
			bondedRatio: sdkmath.LegacyNewDecWithPrec(20, 2),
			cycleEndTime: func() time.Time {
				// returns 2 hours back
				n := time.Now().Add(-time.Hour * 2)
				return n
			},
			ctx: func() sdk.Context {
				return sdkContext.WithBlockTime(time.Now())
			},
			expectedResult: func(minterInflation, bondedRatio sdkmath.LegacyDec, mintParams minttypes.Params) sdkmath.LegacyDec {
				factor := bpmath.One - bpmath.FixedBP(2500)
				return factor.MulDec(mintParams.InflationMax)
			},
		},
		{
			name:        "zero inflation : total supply equals max supply",
			totalSupply: sdkmath.NewInt(100000000),
			minter:      mockMinter,
			mintParams: func(params minttypes.Params) minttypes.Params {
				return params
			},
			inflationParams: func(ip ugov.InflationParams) ugov.InflationParams {
				ip.InflationCycle = time.Hour * 24 * 365
				return ip
			},
			bondedRatio: sdkmath.LegacyNewDecWithPrec(30, 2),
			cycleEndTime: func() time.Time {
				return time.Now()
			},
			ctx: func() sdk.Context {
				sdkContext = sdkContext.WithBlockTime(time.Now())
				return sdkContext
			},
			expectedResult: func(minterInflation, bondedRatio sdkmath.LegacyDec, mintParams minttypes.Params) sdkmath.LegacyDec {
				return sdkmath.LegacyZeroDec()
			},
		},
		{
			name:        "no inflation rate change for min and max: inflation cycle is already started",
			totalSupply: sdkmath.NewInt(900000),
			minter:      mockMinter,
			mintParams: func(params minttypes.Params) minttypes.Params {
				mintParams.BlocksPerYear = 1
				mintParams.InflationMax = sdkmath.LegacyNewDec(7)
				return mintParams
			},
			inflationParams: func(ip ugov.InflationParams) ugov.InflationParams {
				return ip
			},
			bondedRatio: sdkmath.LegacyNewDecWithPrec(20, 2),
			cycleEndTime: func() time.Time {
				return time.Now().Add(2 * time.Hour)
			},
			ctx: func() sdk.Context {
				return sdkContext.WithBlockTime(time.Now())
			},
			expectedResult: func(minterInflation, bondedRatio sdkmath.LegacyDec, mintParams minttypes.Params) sdkmath.LegacyDec {
				inflationRateChangePerYear := sdkmath.LegacyOneDec().Sub(bondedRatio.Quo(mintParams.GoalBonded)).Mul(mintParams.InflationRateChange)
				inflationRateChanges := inflationRateChangePerYear.Quo(sdkmath.LegacyNewDec(int64(mintParams.BlocksPerYear)))
				// adjust the new annual inflation for this next cycle
				inflation := minterInflation.Add(inflationRateChanges) // note inflationRateChange may be negative
				// inflationRateChange := bpmath.FixedBP(25).ToDec().Quo(bpmath.FixedBP(100).ToDec())
				return inflation
			},
		},
		{
			name:        "adjust inflation: when new mint + total supply more than max supply",
			totalSupply: sdkmath.NewInt(999900),
			minter:      mockMinter,
			mintParams: func(params minttypes.Params) minttypes.Params {
				mintParams.BlocksPerYear = 1
				mintParams.InflationMax = sdkmath.LegacyNewDec(7)
				return mintParams
			},
			inflationParams: func(ip ugov.InflationParams) ugov.InflationParams {
				ip.MaxSupply = coin.New(appparams.BondDenom, 1149885)
				return ip
			},
			bondedRatio: sdkmath.LegacyNewDecWithPrec(20, 2),
			cycleEndTime: func() time.Time {
				return time.Now()
			},
			ctx: func() sdk.Context {
				return sdkContext.WithBlockTime(time.Now())
			},
			expectedResult: func(minterInflation, bondedRatio sdkmath.LegacyDec, mintParams minttypes.Params) sdkmath.LegacyDec {
				return sdkmath.LegacyMustNewDecFromStr("0.15")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			// TODO: needs to re-test this
			// mockMintKeeper := mocks.NewMockMintKeeper(ctrl)
			// mockUGovKeeper := ugovmocks.NewMockParamsKeeper(ctrl)

			// mockMintKeeper.EXPECT().StakingTokenSupply(gomock.Any()).Return(test.totalSupply)
			// mockMintKeeper.EXPECT().SetParams(gomock.Any(), gomock.Any()).AnyTimes()

			// mockUGovKeeper.EXPECT().InflationParams().Return(test.inflationParams(mockInflationParams))
			// mockUGovKeeper.EXPECT().InflationCycleEnd().Return(test.cycleEndTime()).AnyTimes()
			// mockUGovKeeper.EXPECT().SetInflationCycleEnd(gomock.Any()).Return(nil).AnyTimes()

			// calc := Calculator{
			// 	MintKeeper:  mockMintKeeper,
			// 	UgovKeeperB: ugovmocks.NewParamsBuilder(mockUGovKeeper),
			// }
			// result := calc.InflationRate(test.ctx(), test.minter, test.mintParams(mintParams), test.bondedRatio)

			// assert.DeepEqual(t,
			// 	test.expectedResult(test.minter.Inflation, test.bondedRatio, test.mintParams(mintParams)), result)
			ctrl.Finish()
		})
	}
}

func TestInflationRateChange(t *testing.T) {
	bondedRatio := sdkmath.LegacyNewDecWithPrec(1, 1) // 10% -> below the goal
	mparamsStd := minttypes.Params{                   // minting params for a standard x/mint minting process
		MintDenom:     sdk.DefaultBondDenom,
		InflationMax:  sdkmath.LegacyNewDecWithPrec(5, 1), // 0.5
		InflationMin:  sdkmath.LegacyNewDecWithPrec(1, 1), // 0.1
		GoalBonded:    sdkmath.LegacyNewDecWithPrec(5, 1), // 0.5
		BlocksPerYear: 5 * 60 * 24 * 365,                  // 1 block per 6s => 5 blocks per min.
	}
	mparamsFast := mparamsStd // minting params for the umee inflation calculator
	mparamsFast.InflationRateChange = fastInflationRateChange(mparamsFast)
	mparamsStd.InflationRateChange = mparamsFast.InflationRateChange.Quo(two)
	minterFast := minttypes.Minter{
		Inflation: sdkmath.LegacyNewDecWithPrec(1, 2), // 0.01  -- less than InflationMin
	}
	minterStd := minterFast

	//
	// Test1: inflation rate should jump to InflationMin in the first round.
	//
	ir := minterFast.NextInflationRate(mparamsFast, bondedRatio)
	assert.Equal(t, ir, mparamsFast.InflationMin, "initial rate should immediately adjust to InflationMin")

	//
	// Test2
	// in half a year inflation should go from 0 towards max. Note: with the existing Cosmos SDK
	// algorithm, we won't reach max. So we compare our settings with the standard minter, and check
	// if it's almost the same.
	//

	// in 5 months, the fast minter should not reach the max.
	month := int(mparamsFast.BlocksPerYear / 12)
	for i := 0; i <= month*5; i++ {
		minterFast.Inflation = minterFast.NextInflationRate(mparamsFast, bondedRatio)
	}
	assert.Assert(t, minterFast.Inflation.LT(mparamsFast.InflationMax), "current: %v", minterFast.Inflation)

	// we should get similar result to the standard minter after 10 months
	for i := 0; i <= month*10; i++ {
		minterStd.Inflation = minterStd.NextInflationRate(mparamsStd, bondedRatio)
	}

	checkers.RequireDecMaxDiff(t, minterStd.Inflation, minterFast.Inflation, sdkmath.LegacyNewDecWithPrec(1, 5),
		"fast minter and standard minter should end up with similar inflation change after 5months and 10months repectively")

	// continue one more month
	for i := 0; i <= month; i++ {
		minterFast.Inflation = minterFast.NextInflationRate(mparamsFast, bondedRatio)
	}
	checkers.RequireDecMaxDiff(t, mparamsFast.InflationMax, minterFast.Inflation,
		mparamsFast.InflationRateChange.QuoInt64(10),
		"fast minter, afer 6 months should go close enough to max")

	//
	// test3, let's see with smaller min and max.
	//
	mparamsFast.InflationMin = sdkmath.LegacyNewDecWithPrec(3, 2) // 0.03
	mparamsFast.InflationMax = sdkmath.LegacyNewDecWithPrec(7, 2) // 0.07
	minterFast.Inflation = sdkmath.LegacyNewDecWithPrec(1, 2)     // 0.01
	for i := 0; i <= month*6; i++ {
		minterFast.Inflation = minterFast.NextInflationRate(mparamsFast, bondedRatio)
	}
	checkers.RequireDecMaxDiff(t, mparamsFast.InflationMax, minterFast.Inflation,
		mparamsFast.InflationRateChange.QuoInt64(10),
		"fast minter, afer 6 months should go close enough to max")

	//
	// test 4 check going from max towards min
	//
	bondedRatio = sdkmath.LegacyNewDecWithPrec(9, 1)          // 0.7
	minterFast.Inflation = sdkmath.LegacyNewDecWithPrec(9, 1) // 0.9
	mparamsFast.InflationRateChange = fastInflationRateChange(mparamsFast)

	ir = minterFast.NextInflationRate(mparamsFast, bondedRatio)
	assert.Equal(t, ir, mparamsFast.InflationMax, "initial rate should immediately adjust to InflationMin")

	// in 5 months, the fast minter should not reach the min.
	for i := 0; i <= month*5; i++ {
		minterFast.Inflation = minterFast.NextInflationRate(mparamsFast, bondedRatio)
	}
	assert.Assert(t, minterFast.Inflation.GT(mparamsFast.InflationMin), "current: %v", minterFast.Inflation)

	// continue one more month
	for i := 0; i <= month; i++ {
		minterFast.Inflation = minterFast.NextInflationRate(mparamsFast, bondedRatio)
	}
	checkers.RequireDecMaxDiff(t, mparamsFast.InflationMin, minterFast.Inflation,
		mparamsFast.InflationRateChange.QuoInt64(10),
		"fast minter, afer 6 months should go close enough to min")

	//
	// test 5, when bondedRatio is closer to the goal bonded we should still go fast.
	//
	bondedRatio = sdkmath.LegacyNewDecWithPrec(7, 1)          // 0.7
	minterFast.Inflation = sdkmath.LegacyNewDecWithPrec(9, 1) // 0.9
	// continue one more month
	for i := 0; i <= month*6; i++ {
		minterFast.Inflation = minterFast.NextInflationRate(mparamsFast, bondedRatio)
	}
	checkers.RequireDecMaxDiff(t, mparamsFast.InflationMin, minterFast.Inflation,
		mparamsFast.InflationRateChange.QuoInt64(3), // TODO: the diff should be smaller, but it will require to change the standard cosmos infation rate algorithm
		"fast minter, afer 6 months should go close enough to min")
}
