package inflation_test

import (
	"testing"
	"time"

	math "cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/golang/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/app/inflation"
	mocks "github.com/umee-network/umee/v6/app/inflation/mocks"
	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/tests/tsdk"
	"github.com/umee-network/umee/v6/util/bpmath"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/ugov"
	ugovmocks "github.com/umee-network/umee/v6/x/ugov/mocks"
)

func TestAdjustInflation(t *testing.T) {
	mintParams := minttypes.DefaultParams()

	tests := []struct {
		name           string
		totalSupply    math.Int
		maxSupply      math.Int
		minter         minttypes.Minter
		params         func(params minttypes.Params) minttypes.Params
		expectedResult sdk.Dec
	}{
		{
			name:        "No inflation change => Newly Minted Coins + Total Supply is less than Max supply",
			totalSupply: math.NewInt(1000000),
			maxSupply:   math.NewInt(2000000),
			minter:      minttypes.Minter{Inflation: sdk.NewDecWithPrec(15, 2)},
			params: func(params minttypes.Params) minttypes.Params {
				/***
				Newly Minted Coins + Total Supply <= Max Supply
				=> No Inflation Rate change
				**/
				params.BlocksPerYear = 1
				return params
			},
			expectedResult: sdk.NewDecWithPrec(15, 2),
		},
		{
			name:        "Inflation Rate Adjust => Newly Minted Coins + Total Supply is greater than Max supply",
			totalSupply: math.NewInt(1900000),
			maxSupply:   math.NewInt(2000000),
			minter:      minttypes.Minter{Inflation: sdk.MustNewDecFromStr("7.1231")},
			params: func(params minttypes.Params) minttypes.Params {
				/***
				Newly Minted Coins + Total Supply >= Max Supply
				=> Inflation will be adjusted to reach Max Supply
				**/
				params.BlocksPerYear = 1
				return params
			},
			expectedResult: sdk.MustNewDecFromStr("0.052631578947368421"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calc := inflation.Calculator{}
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
	mockMinter := minttypes.NewMinter(sdk.MustNewDecFromStr("0.15"), sdk.NewDec(0))
	mockInflationParams := ugov.InflationParams{
		MaxSupply:              coin.New(appparams.BondDenom, 100000000),
		InflationCycle:         time.Hour * 1,
		InflationReductionRate: bpmath.FixedBP(2500),
	}

	sdkContext, _ := tsdk.NewCtx(t, []storetypes.StoreKey{}, []storetypes.StoreKey{})

	tests := []struct {
		name            string
		totalSupply     math.Int
		minter          minttypes.Minter
		inflationParams func(ip ugov.InflationParams) ugov.InflationParams
		bondedRatio     sdk.Dec
		mintParams      func(params minttypes.Params) minttypes.Params
		cycleEndTime    func() time.Time
		ctx             func() sdk.Context
		expectedResult  func(expectedResult, bondedRatio sdk.Dec, mintParams minttypes.Params) sdk.Dec
	}{
		{
			name:        "inflation rate change for min and max: new inflation cyle is started from this block time",
			totalSupply: math.NewInt(900000),
			minter:      mockMinter,
			mintParams: func(params minttypes.Params) minttypes.Params {
				// AnnualProvisions = 900000 * 0.15 = 135000
				mintParams.BlocksPerYear = 1
				mintParams.InflationRateChange = sdk.OneDec()
				return mintParams
			},
			inflationParams: func(ip ugov.InflationParams) ugov.InflationParams {
				return ip
			},
			bondedRatio: sdk.NewDecWithPrec(20, 2),
			cycleEndTime: func() time.Time {
				// returns 2 hours back
				n := time.Now().Add(-time.Hour * 2)
				return n
			},
			ctx: func() sdk.Context {
				return sdkContext.WithBlockTime(time.Now())
			},
			expectedResult: func(minterInflation, bondedRatio sdk.Dec, mintParams minttypes.Params) sdk.Dec {
				factor := bpmath.One - bpmath.FixedBP(2500)
				return factor.MulDec(mintParams.InflationMax)
			},
		},
		{
			name:        "zero inflation : total supply equals max supply",
			totalSupply: math.NewInt(100000000),
			minter:      mockMinter,
			mintParams: func(params minttypes.Params) minttypes.Params {
				return params
			},
			inflationParams: func(ip ugov.InflationParams) ugov.InflationParams {
				ip.InflationCycle = time.Hour * 24 * 365
				return ip
			},
			bondedRatio: sdk.NewDecWithPrec(30, 2),
			cycleEndTime: func() time.Time {
				return time.Now()
			},
			ctx: func() sdk.Context {
				sdkContext = sdkContext.WithBlockTime(time.Now())
				return sdkContext
			},
			expectedResult: func(minterInflation, bondedRatio sdk.Dec, mintParams minttypes.Params) sdk.Dec {
				return sdk.ZeroDec()
			},
		},
		{
			name:        "no inflation rate change for min and max: inflation cycle is already started",
			totalSupply: math.NewInt(900000),
			minter:      mockMinter,
			mintParams: func(params minttypes.Params) minttypes.Params {
				mintParams.BlocksPerYear = 1
				mintParams.InflationMax = sdk.NewDec(7)
				return mintParams
			},
			inflationParams: func(ip ugov.InflationParams) ugov.InflationParams {
				return ip
			},
			bondedRatio: sdk.NewDecWithPrec(20, 2),
			cycleEndTime: func() time.Time {
				return time.Now().Add(2 * time.Hour)
			},
			ctx: func() sdk.Context {
				return sdkContext.WithBlockTime(time.Now())
			},
			expectedResult: func(minterInflation, bondedRatio sdk.Dec, mintParams minttypes.Params) sdk.Dec {
				inflationRateChangePerYear := sdk.OneDec().Sub(bondedRatio.Quo(mintParams.GoalBonded)).Mul(mintParams.InflationRateChange)
				inflationRateChanges := inflationRateChangePerYear.Quo(sdk.NewDec(int64(mintParams.BlocksPerYear)))
				// adjust the new annual inflation for this next cycle
				inflation := minterInflation.Add(inflationRateChanges) // note inflationRateChange may be negative
				// inflationRateChange := bpmath.FixedBP(25).ToDec().Quo(bpmath.FixedBP(100).ToDec())
				return inflation
			},
		},
		{
			name:        "adjust inflation: when new mint + total supply more than max supply",
			totalSupply: math.NewInt(999900),
			minter:      mockMinter,
			mintParams: func(params minttypes.Params) minttypes.Params {
				mintParams.BlocksPerYear = 1
				mintParams.InflationMax = sdk.NewDec(7)
				return mintParams
			},
			inflationParams: func(ip ugov.InflationParams) ugov.InflationParams {
				ip.MaxSupply = coin.New(appparams.BondDenom, 1149885)
				return ip
			},
			bondedRatio: sdk.NewDecWithPrec(20, 2),
			cycleEndTime: func() time.Time {
				return time.Now()
			},
			ctx: func() sdk.Context {
				return sdkContext.WithBlockTime(time.Now())
			},
			expectedResult: func(minterInflation, bondedRatio sdk.Dec, mintParams minttypes.Params) sdk.Dec {
				return sdk.MustNewDecFromStr("0.15")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockMintKeeper := mocks.NewMockMintKeeper(ctrl)
			mockUGovKeeper := ugovmocks.NewMockParamsKeeper(ctrl)

			mockMintKeeper.EXPECT().StakingTokenSupply(gomock.Any()).Return(test.totalSupply)
			mockMintKeeper.EXPECT().SetParams(gomock.Any(), gomock.Any()).AnyTimes()

			mockUGovKeeper.EXPECT().InflationParams().Return(test.inflationParams(mockInflationParams))
			mockUGovKeeper.EXPECT().InflationCycleEnd().Return(test.cycleEndTime()).AnyTimes()
			mockUGovKeeper.EXPECT().SetInflationCycleEnd(gomock.Any()).Return(nil).AnyTimes()

			calc := inflation.Calculator{
				MintKeeper:  mockMintKeeper,
				UgovKeeperB: ugovmocks.NewParamsBuilder(mockUGovKeeper),
			}
			result := calc.InflationRate(test.ctx(), test.minter, test.mintParams(mintParams), test.bondedRatio)

			assert.DeepEqual(t,
				test.expectedResult(test.minter.Inflation, test.bondedRatio, test.mintParams(mintParams)), result)
			ctrl.Finish()
		})
	}
}

func TestNextInflationRate(t *testing.T) {
	minter := minttypes.Minter{
		Inflation: sdk.NewDecWithPrec(0, 2),
	}

	mintParams := minttypes.DefaultParams()
	mintParams.InflationMax = sdk.NewDecWithPrec(40, 2)
	mintParams.InflationMin = sdk.NewDecWithPrec(1, 2)
	mintParams.InflationRateChange = sdk.NewDec(1)
	mintParams.BlocksPerYear = 100
	mintParams.GoalBonded = sdk.NewDec(33)

	bondedRatio := sdk.NewDec(20)

	// default inflation rate (1 year inflation rate change speed )
	ir := minter.NextInflationRate(mintParams, bondedRatio)
	assert.DeepEqual(t, mintParams.InflationMin, ir)

	// changing inflation rate speed from 1 year to 6 months
	mintParams.BlocksPerYear = mintParams.BlocksPerYear * 2
	nir := minter.NextInflationRate(mintParams, bondedRatio)

	assert.DeepEqual(t, mintParams.InflationMin, nir)
}

func TestInflationRateChange(t *testing.T) {
	minter := minttypes.Minter{
		Inflation: sdk.NewDecWithPrec(0, 2),
	}

	mintParams := minttypes.DefaultParams()
	mintParams.InflationMax = sdk.NewDecWithPrec(5, 1) // 0.5
	mintParams.InflationMin = sdk.NewDecWithPrec(2, 2) // 0.02
	mintParams.InflationRateChange = sdk.NewDec(1)     // will be overwritten in the `NextInflationRate`
	mintParams.BlocksPerYear = 100

	bondedRatio := sdk.NewDecWithPrec(1, 2)
	// after 50 blocks (half the year) inflation will be updated
	// every block, inflation = prevInflation + currentInflationRateChange
	var ir sdk.Dec
	ir = minter.NextInflationRate(mintParams, bondedRatio)
	// at initial based on bondedRatio and GoalBonded , the inflation will be at mintParams.InflationMin
	assert.Equal(t, ir, mintParams.InflationMin)
	for i := 0; i < 50; i++ {
		ir = minter.NextInflationRate(mintParams, bondedRatio)
		minter.Inflation = ir
	}
	// current inflation after the 50 blocks will be increased to MaxInflationRate
	nir := minter.NextInflationRate(mintParams, bondedRatio)
	assert.Equal(t, nir, mintParams.InflationMax)

	// current bonded ratio =1 then inflation rate change per year will be negative
	// so after the 50 blocks inflation will be minimum
	minter.Inflation = sdk.NewDecWithPrec(2, 2)
	bondedRatio = sdk.NewDec(1)
	for i := 0; i < 50; i++ {
		ir = minter.NextInflationRate(mintParams, bondedRatio)
		minter.Inflation = ir
	}
	// it should be minimum , because inflationRateChangePerYear will be negative
	assert.Equal(t, ir, mintParams.InflationMin)
}
