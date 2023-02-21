package types

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidateVotePeriod(t *testing.T) {
	err := validateVotePeriod("invalidUint64")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateVotePeriod(uint64(0))
	assert.ErrorContains(t, err, "vote period must be positive: 0")

	err = validateVotePeriod(uint64(10))
	assert.NilError(t, err)
}

func TestValidateVoteThreshold(t *testing.T) {
	err := validateVoteThreshold("invalidSdkType")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateVoteThreshold(sdk.MustNewDecFromStr("0.31"))
	assert.ErrorContains(t, err, "threshold must be bigger than 0.330000000000000000 and <= 1: invalid request")

	err = validateVoteThreshold(sdk.MustNewDecFromStr("40.0"))
	assert.ErrorContains(t, err, "threshold must be bigger than 0.330000000000000000 and <= 1: invalid request")

	err = validateVoteThreshold(sdk.MustNewDecFromStr("0.35"))
	assert.NilError(t, err)
}

func TestValidateRewardBand(t *testing.T) {
	err := validateRewardBand("invalidSdkType")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateRewardBand(sdk.MustNewDecFromStr("-0.31"))
	assert.ErrorContains(t, err, "reward band must be positive: -0.310000000000000000")

	err = validateRewardBand(sdk.MustNewDecFromStr("40.0"))
	assert.ErrorContains(t, err, "reward band is too large: 40.000000000000000000")

	err = validateRewardBand(sdk.OneDec())
	assert.NilError(t, err)
}

func TestValidateRewardDistributionWindow(t *testing.T) {
	err := validateRewardDistributionWindow("invalidUint64")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateRewardDistributionWindow(uint64(0))
	assert.ErrorContains(t, err, "reward distribution window must be positive: 0")

	err = validateRewardDistributionWindow(uint64(10))
	assert.NilError(t, err)
}

func TestValidateAcceptList(t *testing.T) {
	err := validateAcceptList("invalidUint64")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateAcceptList(DenomList{
		{BaseDenom: ""},
	})
	assert.ErrorContains(t, err, "oracle parameter AcceptList Denom must have BaseDenom")

	err = validateAcceptList(DenomList{
		{BaseDenom: DenomUmee.BaseDenom, SymbolDenom: ""},
	})
	assert.ErrorContains(t, err, "oracle parameter AcceptList Denom must have SymbolDenom")

	err = validateAcceptList(DenomList{
		{BaseDenom: DenomUmee.BaseDenom, SymbolDenom: DenomUmee.SymbolDenom},
	})
	assert.NilError(t, err)
}

func TestValidateSlashFraction(t *testing.T) {
	err := validateSlashFraction("invalidSdkType")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateSlashFraction(sdk.MustNewDecFromStr("-0.31"))
	assert.ErrorContains(t, err, "slash fraction must be positive: -0.310000000000000000")

	err = validateSlashFraction(sdk.MustNewDecFromStr("40.0"))
	assert.ErrorContains(t, err, "slash fraction is too large: 40.000000000000000000")

	err = validateSlashFraction(sdk.OneDec())
	assert.NilError(t, err)
}

func TestValidateSlashWindow(t *testing.T) {
	err := validateSlashWindow("invalidUint64")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateSlashWindow(uint64(0))
	assert.ErrorContains(t, err, "slash window must be positive: 0")

	err = validateSlashWindow(uint64(10))
	assert.NilError(t, err)
}

func TestValidateMinValidPerWindow(t *testing.T) {
	err := validateMinValidPerWindow("invalidSdkType")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateMinValidPerWindow(sdk.MustNewDecFromStr("-0.31"))
	assert.ErrorContains(t, err, "min valid per window must be positive: -0.310000000000000000")

	err = validateMinValidPerWindow(sdk.MustNewDecFromStr("40.0"))
	assert.ErrorContains(t, err, "min valid per window is too large: 40.000000000000000000")

	err = validateMinValidPerWindow(sdk.OneDec())
	assert.NilError(t, err)
}

func TestValidateHistoricStampPeriod(t *testing.T) {
	err := validateHistoricStampPeriod("invalidUint64")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateHistoricStampPeriod(uint64(0))
	assert.ErrorContains(t, err, "historic stamp period must be positive: 0")

	err = validateHistoricStampPeriod(uint64(10))
	assert.NilError(t, err)
}

func TestValidateMedianStampPeriod(t *testing.T) {
	err := validateMedianStampPeriod("invalidUint64")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateMedianStampPeriod(uint64(0))
	assert.ErrorContains(t, err, "median stamp period must be positive: 0")

	err = validateMedianStampPeriod(uint64(10))
	assert.NilError(t, err)
}

func TestValidateMaximumPriceStamps(t *testing.T) {
	err := validateMaximumPriceStamps("invalidUint64")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateMaximumPriceStamps(uint64(0))
	assert.ErrorContains(t, err, "maximum price stamps must be positive: 0")

	err = validateMaximumPriceStamps(uint64(10))
	assert.NilError(t, err)
}

func TestValidateMaximumMedianStamps(t *testing.T) {
	err := validateMaximumMedianStamps("invalidUint64")
	assert.ErrorContains(t, err, "invalid parameter type: string")

	err = validateMaximumMedianStamps(uint64(0))
	assert.ErrorContains(t, err, "maximum median stamps must be positive: 0")

	err = validateMaximumMedianStamps(uint64(10))
	assert.NilError(t, err)
}

func TestParamsEqual(t *testing.T) {
	p1 := DefaultParams()
	err := p1.Validate()
	assert.NilError(t, err)

	// minus vote period
	p1.VotePeriod = 0
	err = p1.Validate()
	assert.ErrorContains(t, err, "oracle parameter VotePeriod must be > 0")

	// small vote threshold
	p2 := DefaultParams()
	p2.VoteThreshold = sdk.ZeroDec()
	err = p2.Validate()
	assert.ErrorContains(t, err, "oracle parameter VoteThreshold must be greater than 33 percent")

	// negative reward band
	p3 := DefaultParams()
	p3.RewardBand = sdk.NewDecWithPrec(-1, 2)
	err = p3.Validate()
	assert.ErrorContains(t, err, "oracle parameter RewardBand must be between [0, 1]")

	// negative slash fraction
	p4 := DefaultParams()
	p4.SlashFraction = sdk.NewDec(-1)
	err = p4.Validate()
	assert.ErrorContains(t, err, "oracle parameter SlashFraction must be between [0, 1]")

	// negative min valid per window
	p5 := DefaultParams()
	p5.MinValidPerWindow = sdk.NewDec(-1)
	err = p5.Validate()
	assert.ErrorContains(t, err, "oracle parameter MinValidPerWindow must be between [0, 1")

	// small slash window
	p6 := DefaultParams()
	p6.SlashWindow = 0
	err = p6.Validate()
	assert.ErrorContains(t, err, "oracle parameter SlashWindow must be greater than or equal with VotePeriod")

	// small distribution window
	p7 := DefaultParams()
	p7.RewardDistributionWindow = 0
	err = p7.Validate()
	assert.ErrorContains(t, err, "oracle parameter RewardDistributionWindow must be greater than or equal with VotePeriod")

	// HistoricStampPeriod < MedianStampPeriod
	p8 := DefaultParams()
	p8.HistoricStampPeriod = 10
	p8.MedianStampPeriod = 1
	err = p8.Validate()
	assert.ErrorContains(t, err, "oracle parameter MedianStampPeriod must be greater than or equal with HistoricStampPeriod")

	// HistoricStampPeriod and MedianStampPeriod are multiples of VotePeriod
	p9 := DefaultParams()
	p9.HistoricStampPeriod = 10
	p9.VotePeriod = 3
	err = p9.Validate()
	assert.ErrorContains(t, err, "oracle parameters HistoricStampPeriod and MedianStampPeriod must be exact multiples of VotePeriod")
	p9.MedianStampPeriod = 10
	err = p9.Validate()
	assert.ErrorContains(t, err, "oracle parameters HistoricStampPeriod and MedianStampPeriod must be exact multiples of VotePeriod")

	// empty name
	p10 := DefaultParams()
	p10.AcceptList[0].BaseDenom = ""
	p10.AcceptList[0].SymbolDenom = "ATOM"
	err = p10.Validate()
	assert.ErrorContains(t, err, "oracle parameter AcceptList Denom must have BaseDenom")

	// empty
	p11 := DefaultParams()
	p11.AcceptList[0].BaseDenom = "uatom"
	p11.AcceptList[0].SymbolDenom = ""
	err = p11.Validate()
	assert.ErrorContains(t, err, "oracle parameter AcceptList Denom must have SymbolDenom")

	p13 := DefaultParams()
	assert.Equal(t, len(p13.AcceptList), 2)
}

func TestValidateVotingThreshold(t *testing.T) {
	tcs := []struct {
		name   string
		t      sdk.Dec
		errMsg string
	}{
		{"fail: negative", sdk.MustNewDecFromStr("-1"), "threshold must be"},
		{"fail: zero", sdk.ZeroDec(), "threshold must be"},
		{"fail: less than 0.33", sdk.MustNewDecFromStr("0.3"), "threshold must be"},
		{"fail: equal 0.33", sdk.MustNewDecFromStr("0.33"), "threshold must be"},
		{"fail: more than 1", sdk.MustNewDecFromStr("1.1"), "threshold must be"},
		{"fail: more than 1", sdk.MustNewDecFromStr("10"), "threshold must be"},
		{"fail: max precision 2", sdk.MustNewDecFromStr("0.333"), "maximum 2 decimals"},
		{"fail: max precision 2", sdk.MustNewDecFromStr("0.401"), "maximum 2 decimals"},
		{"fail: max precision 2", sdk.MustNewDecFromStr("0.409"), "maximum 2 decimals"},
		{"fail: max precision 2", sdk.MustNewDecFromStr("0.4009"), "maximum 2 decimals"},
		{"fail: max precision 2", sdk.MustNewDecFromStr("0.999"), "maximum 2 decimals"},

		{"ok: 1", sdk.MustNewDecFromStr("1"), ""},
		{"ok: 0.34", sdk.MustNewDecFromStr("0.34"), ""},
		{"ok: 0.99", sdk.MustNewDecFromStr("0.99"), ""},
	}

	for _, tc := range tcs {
		err := ValidateVoteThreshold(tc.t)
		if tc.errMsg == "" {
			assert.NilError(t, err, "test_case", tc.name)
		} else {
			assert.ErrorContains(t, err, tc.errMsg, tc.name)
		}
	}
}
