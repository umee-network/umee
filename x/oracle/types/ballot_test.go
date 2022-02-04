package types

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestToMap(t *testing.T) {
	tests := struct {
		votes   []VoteForTally
		isValid []bool
	}{

		[]VoteForTally{
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        UmeeDenom,
				ExchangeRate: sdk.NewDec(1600),
				Power:        100,
			},
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        UmeeDenom,
				ExchangeRate: sdk.ZeroDec(),
				Power:        100,
			},
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        UmeeDenom,
				ExchangeRate: sdk.NewDec(1500),
				Power:        100,
			},
		},
		[]bool{true, false, true},
	}

	pb := ExchangeRateBallot(tests.votes)
	mapData := pb.ToMap()

	for i, vote := range tests.votes {
		exchangeRate, ok := mapData[vote.Voter.String()]
		if tests.isValid[i] {
			require.True(t, ok)
			require.Equal(t, exchangeRate, vote.ExchangeRate)
		} else {
			require.False(t, ok)
		}
	}
}

func TestSqrt(t *testing.T) {
	num := sdk.NewDecWithPrec(144, 4)
	floatNum, err := strconv.ParseFloat(num.String(), 64)
	require.NoError(t, err)

	floatNum = math.Sqrt(floatNum)
	num, err = sdk.NewDecFromStr(fmt.Sprintf("%f", floatNum))
	require.NoError(t, err)

	require.Equal(t, sdk.NewDecWithPrec(12, 2), num)
}

func TestPBPower(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	_, valAccAddrs, sk := GenerateRandomTestCase()
	pb := ExchangeRateBallot{}
	ballotPower := int64(0)

	for i := 0; i < len(sk.Validators()); i++ {
		power := sk.Validator(ctx, valAccAddrs[i]).GetConsensusPower(sdk.DefaultPowerReduction)
		vote := NewVoteForTally(
			sdk.ZeroDec(),
			UmeeDenom,
			valAccAddrs[i],
			power,
		)

		pb = append(pb, vote)
		require.NotEqual(t, int64(0), vote.Power)

		ballotPower += vote.Power
	}

	require.Equal(t, ballotPower, pb.Power())

	// Mix in a fake validator, the total power should not have changed.
	pubKey := secp256k1.GenPrivKey().PubKey()
	faceValAddr := sdk.ValAddress(pubKey.Address())
	fakeVote := NewVoteForTally(
		sdk.OneDec(),
		UmeeDenom,
		faceValAddr,
		0,
	)

	pb = append(pb, fakeVote)
	require.Equal(t, ballotPower, pb.Power())
}

func TestPBWeightedMedian(t *testing.T) {
	tests := []struct {
		inputs      []int64
		weights     []int64
		isValidator []bool
		median      sdk.Dec
		success     bool
	}{
		{
			// Supermajority one number
			[]int64{1, 2, 10, 100000},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			sdk.NewDec(10),
			true,
		},
		{
			// Adding fake validator doesn't change outcome
			[]int64{1, 2, 10, 100000, 10000000000},
			[]int64{1, 1, 100, 1, 10000},
			[]bool{true, true, true, true, false},
			sdk.NewDec(10),
			true,
		},
		{
			// Tie votes
			[]int64{1, 2, 3, 4},
			[]int64{1, 100, 100, 1},
			[]bool{true, true, true, true},
			sdk.NewDec(2),
			true,
		},
		{
			// No votes
			[]int64{},
			[]int64{},
			[]bool{true, true, true, true},
			sdk.NewDec(0),
			true,
		},
		{
			// Out of order
			[]int64{1, 2, 10, 3},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			sdk.NewDec(10),
			false,
		},
	}

	for _, tc := range tests {
		pb := ExchangeRateBallot{}
		for i, input := range tc.inputs {
			valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())

			power := tc.weights[i]
			if !tc.isValidator[i] {
				power = 0
			}

			vote := NewVoteForTally(
				sdk.NewDec(int64(input)),
				UmeeDenom,
				valAddr,
				power,
			)

			pb = append(pb, vote)
		}

		median, err := pb.WeightedMedian()
		if tc.success {
			require.NoError(t, err)
			require.Equal(t, tc.median, median)
		} else {
			require.Error(t, err)
		}

	}
}

func TestPBStandardDeviation(t *testing.T) {
	tests := []struct {
		inputs            []float64
		weights           []int64
		isValidator       []bool
		standardDeviation sdk.Dec
	}{
		{
			// Supermajority one number
			[]float64{1.0, 2.0, 10.0, 100000.0},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			sdk.MustNewDecFromStr("49995.000362536252310906"),
		},
		{
			// Adding fake validator doesn't change outcome
			[]float64{1.0, 2.0, 10.0, 100000.0, 10000000000},
			[]int64{1, 1, 100, 1, 10000},
			[]bool{true, true, true, true, false},
			sdk.MustNewDecFromStr("4472135950.751005519905537611"),
		},
		{
			// Tie votes
			[]float64{1.0, 2.0, 3.0, 4.0},
			[]int64{1, 100, 100, 1},
			[]bool{true, true, true, true},
			sdk.MustNewDecFromStr("1.224744871391589049"),
		},
		{
			// No votes
			[]float64{},
			[]int64{},
			[]bool{true, true, true, true},
			sdk.NewDecWithPrec(0, 0),
		},
	}

	base := math.Pow10(OracleDecPrecision)
	for _, tc := range tests {
		pb := ExchangeRateBallot{}
		for i, input := range tc.inputs {
			valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())

			power := tc.weights[i]
			if !tc.isValidator[i] {
				power = 0
			}

			vote := NewVoteForTally(
				sdk.NewDecWithPrec(int64(input*base), int64(OracleDecPrecision)),
				UmeeDenom,
				valAddr,
				power,
			)

			pb = append(pb, vote)
		}
		stdDev, _ := pb.StandardDeviation()

		require.Equal(t, tc.standardDeviation, stdDev)
	}
}

func TestPBStandardDeviation_Overflow(t *testing.T) {
	valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())
	overflowRate, err := sdk.NewDecFromStr("100000000000000000000000000000000000000000000000000000000.0")
	require.NoError(t, err)
	pb := ExchangeRateBallot{
		NewVoteForTally(
			sdk.OneDec(),
			UmeeSymbol,
			valAddr,
			2,
		),
		NewVoteForTally(
			sdk.NewDec(1234),
			UmeeSymbol,
			valAddr,
			2,
		),
		NewVoteForTally(
			overflowRate,
			UmeeSymbol,
			valAddr,
			1,
		),
	}

	deviation, err := pb.StandardDeviation()
	require.NoError(t, err)
	expectedDevation, _ := sdk.NewDecFromStr("871.862661203013097586")
	require.Equal(t, expectedDevation, deviation)
}
