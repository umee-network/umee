package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestParamKeyTable(t *testing.T) {
	require.NotNil(t, ParamKeyTable())
}

func TestValidateVotePeriod(t *testing.T) {
	err := validateVotePeriod("invalidUint64")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateVotePeriod(uint64(0))
	require.ErrorContains(t, err, "vote period must be positive: 0")

	err = validateVotePeriod(uint64(10))
	require.Nil(t, err)
}

func TestValidateVoteThreshold(t *testing.T) {
	err := validateVoteThreshold("invalidSdkType")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateVoteThreshold(sdk.MustNewDecFromStr("0.31"))
	require.ErrorContains(t, err, "vote threshold must be bigger than 33%: 0.310000000000000000")

	err = validateVoteThreshold(sdk.MustNewDecFromStr("40.0"))
	require.ErrorContains(t, err, "vote threshold too large: 40.000000000000000000")

	err = validateVoteThreshold(sdk.MustNewDecFromStr("0.35"))
	require.Nil(t, err)
}

func TestValidateRewardBand(t *testing.T) {
	err := validateRewardBand("invalidSdkType")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateRewardBand(sdk.MustNewDecFromStr("-0.31"))
	require.ErrorContains(t, err, "reward band must be positive: -0.310000000000000000")

	err = validateRewardBand(sdk.MustNewDecFromStr("40.0"))
	require.ErrorContains(t, err, "reward band is too large: 40.000000000000000000")

	err = validateRewardBand(sdk.OneDec())
	require.Nil(t, err)
}

func TestValidateRewardDistributionWindow(t *testing.T) {
	err := validateRewardDistributionWindow("invalidUint64")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateRewardDistributionWindow(uint64(0))
	require.ErrorContains(t, err, "reward distribution window must be positive: 0")

	err = validateRewardDistributionWindow(uint64(10))
	require.Nil(t, err)
}

func TestValidateAcceptList(t *testing.T) {
	err := validateAcceptList("invalidUint64")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateAcceptList(DenomList{
		{BaseDenom: ""},
	})
	require.ErrorContains(t, err, "oracle parameter AcceptList Denom must have BaseDenom")

	err = validateAcceptList(DenomList{
		{BaseDenom: DenomUmee.BaseDenom, SymbolDenom: ""},
	})
	require.ErrorContains(t, err, "oracle parameter AcceptList Denom must have SymbolDenom")

	err = validateAcceptList(DenomList{
		{BaseDenom: DenomUmee.BaseDenom, SymbolDenom: DenomUmee.SymbolDenom},
	})
	require.Nil(t, err)
}

func TestValidateSlashFraction(t *testing.T) {
	err := validateSlashFraction("invalidSdkType")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateSlashFraction(sdk.MustNewDecFromStr("-0.31"))
	require.ErrorContains(t, err, "slash fraction must be positive: -0.310000000000000000")

	err = validateSlashFraction(sdk.MustNewDecFromStr("40.0"))
	require.ErrorContains(t, err, "slash fraction is too large: 40.000000000000000000")

	err = validateSlashFraction(sdk.OneDec())
	require.Nil(t, err)
}

func TestValidateSlashWindow(t *testing.T) {
	err := validateSlashWindow("invalidUint64")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateSlashWindow(uint64(0))
	require.ErrorContains(t, err, "slash window must be positive: 0")

	err = validateSlashWindow(uint64(10))
	require.Nil(t, err)
}

func TestValidateMinValidPerWindow(t *testing.T) {
	err := validateMinValidPerWindow("invalidSdkType")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateMinValidPerWindow(sdk.MustNewDecFromStr("-0.31"))
	require.ErrorContains(t, err, "min valid per window must be positive: -0.310000000000000000")

	err = validateMinValidPerWindow(sdk.MustNewDecFromStr("40.0"))
	require.ErrorContains(t, err, "min valid per window is too large: 40.000000000000000000")

	err = validateMinValidPerWindow(sdk.OneDec())
	require.Nil(t, err)
}

func TestValidateHistoricStampPeriod(t *testing.T) {
	err := validateHistoricStampPeriod("invalidUint64")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateHistoricStampPeriod(uint64(0))
	require.ErrorContains(t, err, "historic stamp period must be positive: 0")

	err = validateHistoricStampPeriod(uint64(10))
	require.Nil(t, err)
}

func TestValidateMedianStampPeriod(t *testing.T) {
	err := validateMedianStampPeriod("invalidUint64")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateMedianStampPeriod(uint64(0))
	require.ErrorContains(t, err, "median stamp period must be positive: 0")

	err = validateMedianStampPeriod(uint64(10))
	require.Nil(t, err)
}

func TestValidateMaximumPriceStamps(t *testing.T) {
	err := validateMaximumPriceStamps("invalidUint64")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateMaximumPriceStamps(uint64(0))
	require.ErrorContains(t, err, "maximum price stamps must be positive: 0")

	err = validateMaximumPriceStamps(uint64(10))
	require.Nil(t, err)
}

func TestValidateMaximumMedianStamps(t *testing.T) {
	err := validateMaximumMedianStamps("invalidUint64")
	require.ErrorContains(t, err, "invalid parameter type: string")

	err = validateMaximumMedianStamps(uint64(0))
	require.ErrorContains(t, err, "maximum median stamps must be positive: 0")

	err = validateMaximumMedianStamps(uint64(10))
	require.Nil(t, err)
}

func TestParamsEqual(t *testing.T) {
	p1 := DefaultParams()
	err := p1.Validate()
	require.NoError(t, err)

	// minus vote period
	p1.VotePeriod = 0
	err = p1.Validate()
	require.Error(t, err)

	// small vote threshold
	p2 := DefaultParams()
	p2.VoteThreshold = sdk.ZeroDec()
	err = p2.Validate()
	require.Error(t, err)

	// negative reward band
	p3 := DefaultParams()
	p3.RewardBand = sdk.NewDecWithPrec(-1, 2)
	err = p3.Validate()
	require.Error(t, err)

	// negative slash fraction
	p4 := DefaultParams()
	p4.SlashFraction = sdk.NewDec(-1)
	err = p4.Validate()
	require.Error(t, err)

	// negative min valid per window
	p5 := DefaultParams()
	p5.MinValidPerWindow = sdk.NewDec(-1)
	err = p5.Validate()
	require.Error(t, err)

	// small slash window
	p6 := DefaultParams()
	p6.SlashWindow = 0
	err = p6.Validate()
	require.Error(t, err)

	// small distribution window
	p7 := DefaultParams()
	p7.RewardDistributionWindow = 0
	err = p7.Validate()
	require.Error(t, err)

	// HistoricStampPeriod < MedianStampPeriod
	p8 := DefaultParams()
	p8.HistoricStampPeriod = 10
	p8.MedianStampPeriod = 1
	err = p8.Validate()
	require.Error(t, err)

	// HistoricStampPeriod and MedianStampPeriod are multiples of VotePeriod
	p9 := DefaultParams()
	p9.HistoricStampPeriod = 10
	p9.VotePeriod = 3
	err = p9.Validate()
	require.Error(t, err)
	p9.MedianStampPeriod = 10
	err = p9.Validate()
	require.Error(t, err)

	// empty name
	p10 := DefaultParams()
	p10.AcceptList[0].BaseDenom = ""
	p10.AcceptList[0].SymbolDenom = "ATOM"
	err = p10.Validate()
	require.Error(t, err)

	// empty
	p11 := DefaultParams()
	p11.AcceptList[0].BaseDenom = "uatom"
	p11.AcceptList[0].SymbolDenom = ""
	err = p11.Validate()
	require.Error(t, err)

	p13 := DefaultParams()
	require.NotNil(t, p13.ParamSetPairs())
	require.NotNil(t, p13.String())
}
