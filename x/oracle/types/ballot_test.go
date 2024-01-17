package types

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
				ExchangeRate: sdkmath.LegacyNewDec(1600),
				Power:        100,
			},
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        UmeeDenom,
				ExchangeRate: sdkmath.LegacyZeroDec(),
				Power:        100,
			},
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        UmeeDenom,
				ExchangeRate: sdkmath.LegacyNewDec(1500),
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
			assert.Equal(t, true, ok)
			assert.Equal(t, exchangeRate, vote.ExchangeRate)
		} else {
			assert.Equal(t, false, ok)
		}
	}
}

func TestSqrt(t *testing.T) {
	num := sdkmath.LegacyNewDecWithPrec(144, 4)
	floatNum, err := strconv.ParseFloat(num.String(), 64)
	assert.NilError(t, err)

	floatNum = math.Sqrt(floatNum)
	num, err = sdkmath.LegacyNewDecFromStr(fmt.Sprintf("%f", floatNum))
	assert.NilError(t, err)

	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(12, 2), num)
}

func TestPBPower(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	valAccAddrs, sk := GenerateRandomTestCase()
	pb := ExchangeRateBallot{}
	ballotPower := int64(0)

	for i := 0; i < len(sk.Validators()); i++ {
		power := sk.Validator(ctx, valAccAddrs[i]).GetConsensusPower(sdk.DefaultPowerReduction)
		vote := NewVoteForTally(
			sdkmath.LegacyZeroDec(),
			UmeeDenom,
			valAccAddrs[i],
			power,
		)

		pb = append(pb, vote)
		ballotPower += vote.Power
	}

	assert.Equal(t, ballotPower, pb.Power())

	// Mix in a fake validator, the total power should not have changed.
	pubKey := secp256k1.GenPrivKey().PubKey()
	faceValAddr := sdk.ValAddress(pubKey.Address())
	fakeVote := NewVoteForTally(
		sdkmath.LegacyOneDec(),
		UmeeDenom,
		faceValAddr,
		0,
	)

	pb = append(pb, fakeVote)
	assert.Equal(t, ballotPower, pb.Power())
}

func TestPBWeightedMedian(t *testing.T) {
	tests := []struct {
		inputs      []int64
		weights     []int64
		isValidator []bool
		median      sdkmath.LegacyDec
		errMsg      string
	}{
		{
			// Supermajority one number
			[]int64{1, 2, 10, 100000},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			sdkmath.LegacyNewDec(10),
			"",
		},
		{
			// Adding fake validator doesn't change outcome
			[]int64{1, 2, 10, 100000, 10000000000},
			[]int64{1, 1, 100, 1, 10000},
			[]bool{true, true, true, true, false},
			sdkmath.LegacyNewDec(10),
			"",
		},
		{
			// Tie votes
			[]int64{1, 2, 3, 4},
			[]int64{1, 100, 100, 1},
			[]bool{true, true, true, true},
			sdkmath.LegacyNewDec(2),
			"",
		},
		{
			// No votes
			[]int64{},
			[]int64{},
			[]bool{true, true, true, true},
			sdkmath.LegacyNewDec(0),
			"",
		},
		{
			// Out of order
			[]int64{1, 2, 10, 3},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			sdkmath.LegacyNewDec(10),
			"ballot must be sorted before this operation",
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
				sdkmath.LegacyNewDec(int64(input)),
				UmeeDenom,
				valAddr,
				power,
			)

			pb = append(pb, vote)
		}

		median, err := pb.WeightedMedian()
		if tc.errMsg == "" {
			assert.NilError(t, err)
			assert.DeepEqual(t, tc.median, median)
		} else {
			assert.ErrorContains(t, err, tc.errMsg)
		}
	}
}

func TestPBStandardDeviation(t *testing.T) {
	tests := []struct {
		inputs            []sdkmath.LegacyDec
		weights           []int64
		isValidator       []bool
		standardDeviation sdkmath.LegacyDec
	}{
		{
			// Supermajority one number
			[]sdkmath.LegacyDec{
				sdkmath.LegacyMustNewDecFromStr("1.0"),
				sdkmath.LegacyMustNewDecFromStr("2.0"),
				sdkmath.LegacyMustNewDecFromStr("10.0"),
				sdkmath.LegacyMustNewDecFromStr("100000.00"),
			},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			sdkmath.LegacyMustNewDecFromStr("49995.000362536252310906"),
		},
		{
			// Adding fake validator doesn't change outcome
			[]sdkmath.LegacyDec{
				sdkmath.LegacyMustNewDecFromStr("1.0"),
				sdkmath.LegacyMustNewDecFromStr("2.0"),
				sdkmath.LegacyMustNewDecFromStr("10.0"),
				sdkmath.LegacyMustNewDecFromStr("100000.00"),
				sdkmath.LegacyMustNewDecFromStr("10000000000"),
			},
			[]int64{1, 1, 100, 1, 10000},
			[]bool{true, true, true, true, false},
			sdkmath.LegacyMustNewDecFromStr("4472135950.751005519905537611"),
		},
		{
			// Tie votes
			[]sdkmath.LegacyDec{
				sdkmath.LegacyMustNewDecFromStr("1.0"),
				sdkmath.LegacyMustNewDecFromStr("2.0"),
				sdkmath.LegacyMustNewDecFromStr("3.0"),
				sdkmath.LegacyMustNewDecFromStr("4.00"),
			},
			[]int64{1, 100, 100, 1},
			[]bool{true, true, true, true},
			sdkmath.LegacyMustNewDecFromStr("1.224744871391589049"),
		},
		{
			// No votes
			[]sdkmath.LegacyDec{},
			[]int64{},
			[]bool{true, true, true, true},
			sdkmath.LegacyNewDecWithPrec(0, 0),
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
				input,
				UmeeDenom,
				valAddr,
				power,
			)

			pb = append(pb, vote)
		}
		stdDev, _ := pb.StandardDeviation()
		assert.DeepEqual(t, tc.standardDeviation, stdDev)
	}
}

func TestPBStandardDeviation_Overflow(t *testing.T) {
	valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())
	overflowRate, err := sdkmath.LegacyNewDecFromStr("100000000000000000000000000000000000000000000000000000000.0")
	assert.NilError(t, err)
	pb := ExchangeRateBallot{
		NewVoteForTally(
			sdkmath.LegacyOneDec(),
			UmeeSymbol,
			valAddr,
			2,
		),
		NewVoteForTally(
			sdkmath.LegacyNewDec(1234),
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
	assert.NilError(t, err)
	expectedDevation := sdkmath.LegacyMustNewDecFromStr("871.862661203013097586")
	assert.DeepEqual(t, expectedDevation, deviation)
}

func TestBallotMapToSlice(t *testing.T) {
	valAddress := GenerateRandomValAddr(1)

	pb := ExchangeRateBallot{
		NewVoteForTally(
			sdkmath.LegacyNewDec(1234),
			UmeeSymbol,
			valAddress[0],
			2,
		),
		NewVoteForTally(
			sdkmath.LegacyNewDec(12345),
			UmeeSymbol,
			valAddress[0],
			1,
		),
	}

	ballotSlice := BallotMapToSlice(map[string]ExchangeRateBallot{
		UmeeDenom:    pb,
		IbcDenomAtom: pb,
	})
	assert.DeepEqual(t, []BallotDenom{{Ballot: pb, Denom: IbcDenomAtom}, {Ballot: pb, Denom: UmeeDenom}}, ballotSlice)
}

func TestExchangeRateBallotSwap(t *testing.T) {
	valAddress := GenerateRandomValAddr(2)

	voteTallies := []VoteForTally{
		NewVoteForTally(
			sdkmath.LegacyNewDec(1234),
			UmeeSymbol,
			valAddress[0],
			2,
		),
		NewVoteForTally(
			sdkmath.LegacyNewDec(12345),
			UmeeSymbol,
			valAddress[1],
			1,
		),
	}

	pb := ExchangeRateBallot{voteTallies[0], voteTallies[1]}

	assert.DeepEqual(t, pb[0], voteTallies[0])
	assert.DeepEqual(t, pb[1], voteTallies[1])
	pb.Swap(1, 0)
	assert.DeepEqual(t, pb[1], voteTallies[0])
	assert.DeepEqual(t, pb[0], voteTallies[1])
}

func TestStandardDeviationUnsorted(t *testing.T) {
	valAddress := GenerateRandomValAddr(1)
	pb := ExchangeRateBallot{
		NewVoteForTally(
			sdkmath.LegacyNewDec(1234),
			UmeeSymbol,
			valAddress[0],
			2,
		),
		NewVoteForTally(
			sdkmath.LegacyNewDec(12),
			UmeeSymbol,
			valAddress[0],
			1,
		),
	}

	deviation, err := pb.StandardDeviation()
	assert.ErrorIs(t, err, ErrBallotNotSorted)
	assert.Equal(t, "0.000000000000000000", deviation.String())
}

func TestClaimMapToSlice(t *testing.T) {
	valAddress := GenerateRandomValAddr(1)
	claim := NewClaim(10, 1, 4, valAddress[0])
	claimSlice := ClaimMapToSlice(map[string]Claim{
		"testClaim":    claim,
		"anotherClaim": claim,
	})
	assert.DeepEqual(t, []Claim{claim, claim}, claimSlice)
}

func TestExchangeRateBallotSort(t *testing.T) {
	v1 := VoteForTally{ExchangeRate: sdkmath.LegacyMustNewDecFromStr("0.2"), Voter: sdk.ValAddress{0, 1}}
	v1Cpy := VoteForTally{ExchangeRate: sdkmath.LegacyMustNewDecFromStr("0.2"), Voter: sdk.ValAddress{0, 1}}
	v2 := VoteForTally{ExchangeRate: sdkmath.LegacyMustNewDecFromStr("0.1"), Voter: sdk.ValAddress{0, 1, 1}}
	v3 := VoteForTally{ExchangeRate: sdkmath.LegacyMustNewDecFromStr("0.1"), Voter: sdk.ValAddress{0, 1}}
	v4 := VoteForTally{ExchangeRate: sdkmath.LegacyMustNewDecFromStr("0.5"), Voter: sdk.ValAddress{1}}

	tcs := []struct {
		got      ExchangeRateBallot
		expected ExchangeRateBallot
	}{
		{
			got:      ExchangeRateBallot{v1, v2, v3, v4},
			expected: ExchangeRateBallot{v3, v2, v1, v4},
		},
		{
			got:      ExchangeRateBallot{v1},
			expected: ExchangeRateBallot{v1},
		},
		{
			got:      ExchangeRateBallot{v1, v1Cpy},
			expected: ExchangeRateBallot{v1, v1Cpy},
		},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			sort.Sort(tc.got)
			assert.DeepEqual(t, tc.expected, tc.got)
		})
	}
}
