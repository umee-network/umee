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
	require.ErrorContains(t, validateVoteThreshold("invalidSdkType"), "invalid parameter type: string")
	require.ErrorContains(t, validateVoteThreshold(sdk.NewDecWithPrec(31, 2)), "vote threshold must be bigger than 33%: 0.310000000000000000")
	require.ErrorContains(t, validateVoteThreshold(sdk.NewDecWithPrec(4000, 2)), "vote threshold too large: 40.000000000000000000")
	require.Nil(t, validateVoteThreshold(sdk.NewDecWithPrec(35, 2)))
}

func TestValidateRewardBand(t *testing.T) {
	require.ErrorContains(t, validateRewardBand("invalidSdkType"), "invalid parameter type: string")
	require.ErrorContains(t, validateRewardBand(sdk.NewDecWithPrec(-31, 2)), "reward band must be positive: -0.310000000000000000")
	require.ErrorContains(t, validateRewardBand(sdk.MustNewDecFromStr("40.0")), "reward band is too large: 40.000000000000000000")
	require.Nil(t, validateRewardBand(sdk.OneDec()))
}

func TestValidateRewardDistributionWindow(t *testing.T) {
	require.ErrorContains(t, validateRewardDistributionWindow("invalidUint64"), "invalid parameter type: string")
	require.ErrorContains(t, validateRewardDistributionWindow(uint64(0)), "reward distribution window must be positive: 0")
	require.Nil(t, validateRewardDistributionWindow(uint64(10)))
}

func TestValidateAcceptList(t *testing.T) {
	require.ErrorContains(t, validateAcceptList("invalidUint64"), "invalid parameter type: string")
	require.ErrorContains(t, validateAcceptList(DenomList{
		{BaseDenom: ""},
	}), "oracle parameter AcceptList Denom must have BaseDenom")
	require.ErrorContains(t, validateAcceptList(DenomList{
		{BaseDenom: DenomUmee.BaseDenom, SymbolDenom: ""},
	}), "oracle parameter AcceptList Denom must have SymbolDenom")
	require.Nil(t, validateAcceptList(DenomList{
		{BaseDenom: DenomUmee.BaseDenom, SymbolDenom: DenomUmee.SymbolDenom},
	}))
}

func TestValidateSlashFraction(t *testing.T) {
	require.ErrorContains(t, validateSlashFraction("invalidSdkType"), "invalid parameter type: string")
	require.ErrorContains(t, validateSlashFraction(sdk.NewDecWithPrec(-31, 2)), "slash fraction must be positive: -0.310000000000000000")
	require.ErrorContains(t, validateSlashFraction(sdk.NewDecWithPrec(4000, 2)), "slash fraction is too large: 40.000000000000000000")
	require.Nil(t, validateSlashFraction(sdk.OneDec()))
}

func TestValidateSlashWindow(t *testing.T) {
	require.ErrorContains(t, validateSlashWindow("invalidUint64"), "invalid parameter type: string")
	require.ErrorContains(t, validateSlashWindow(uint64(0)), "slash window must be positive: 0")
	require.Nil(t, validateSlashWindow(uint64(10)))
}

func TestValidateMinValidPerWindow(t *testing.T) {
	require.ErrorContains(t, validateMinValidPerWindow("invalidSdkType"), "invalid parameter type: string")
	require.ErrorContains(t, validateMinValidPerWindow(sdk.NewDecWithPrec(-31, 2)), "min valid per window must be positive: -0.310000000000000000")
	require.ErrorContains(t, validateMinValidPerWindow(sdk.NewDecWithPrec(4000, 2)), "min valid per window is too large: 40.000000000000000000")
	require.Nil(t, validateMinValidPerWindow(sdk.OneDec()))
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

	// empty name
	p9 := DefaultParams()
	p9.AcceptList[0].BaseDenom = ""
	p9.AcceptList[0].SymbolDenom = "ATOM"
	err = p9.Validate()
	require.Error(t, err)

	// empty
	p10 := DefaultParams()
	p10.AcceptList[0].BaseDenom = "uatom"
	p10.AcceptList[0].SymbolDenom = ""
	err = p10.Validate()
	require.Error(t, err)

	p11 := DefaultParams()
	require.NotNil(t, p11.ParamSetPairs())
	require.NotNil(t, p11.String())
}
