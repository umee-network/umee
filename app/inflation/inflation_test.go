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

	"github.com/umee-network/umee/v5/app/inflation"
	mocks "github.com/umee-network/umee/v5/app/inflation/mocks"
	appparams "github.com/umee-network/umee/v5/app/params"
	"github.com/umee-network/umee/v5/tests/tsdk"
	"github.com/umee-network/umee/v5/util/bpmath"
	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/ugov"
	ugovmocks "github.com/umee-network/umee/v5/x/ugov/mocks"
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
			// Create the mock MintKeeper and UgovKeeper
			mockMintKeeper := mocks.NewMockMintKeeper(ctrl)

			calc := inflation.Calculator{
				MintKeeper:  mockMintKeeper,
				UgovKeeperB: ugovmocks.NewUgovParamsBuilder(ctrl),
			}

			// mockUGovBuilder returns the mockUGovKeeper
			mockUGovBuilder.EXPECT().Keeper(gomock.Any()).Return(mockUGovKeeper)

			// Set up the mock behavior for MintKeeper and UgovKeeper
			mockUGovKeeper.EXPECT().InflationParams().Return(test.inflationParams(mockInflationParams))
			mockMintKeeper.EXPECT().StakingTokenSupply(gomock.Any()).Return(test.totalSupply)
			mockMintKeeper.EXPECT().SetParams(gomock.Any(), gomock.Any()).AnyTimes()
			mockUGovKeeper.EXPECT().InflationCycleEnd().Return(test.cycleEndTime()).AnyTimes()
			mockUGovKeeper.EXPECT().SetInflationCycleEnd(gomock.Any()).Return(nil).AnyTimes()

			result := calc.InflationRate(test.ctx(), test.minter, test.mintParams(mintParams), test.bondedRatio)

			assert.DeepEqual(t, test.expectedResult(test.minter.Inflation, test.bondedRatio, test.mintParams(mintParams)), result)
			ctrl.Finish()
		})
	}
}
