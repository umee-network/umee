package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgFeederDelegation(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
		sdk.AccAddress([]byte("addr2_______________")),
	}

	tests := []struct {
		delegator  sdk.ValAddress
		delegate   sdk.AccAddress
		expectPass bool
	}{
		{sdk.ValAddress(addrs[0]), addrs[1], true},
		{sdk.ValAddress{}, addrs[1], false},
		{sdk.ValAddress(addrs[0]), sdk.AccAddress{}, false},
		{nil, nil, false},
	}

	for i, tc := range tests {
		msg := NewMsgDelegateFeedConsent(tc.delegator, tc.delegate)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgAggregateExchangeRatePrevote(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
	}

	exchangeRates := sdk.DecCoins{sdk.NewDecCoinFromDec(UmeeDenom, sdk.OneDec()), sdk.NewDecCoinFromDec(UmeeDenom, sdk.NewDecWithPrec(32121, 1))}
	bz := GetAggregateVoteHash("1", exchangeRates.String(), sdk.ValAddress(addrs[0]))

	tests := []struct {
		hash          AggregateVoteHash
		exchangeRates sdk.DecCoins
		feeder        sdk.AccAddress
		validator     sdk.AccAddress
		expectPass    bool
	}{
		{bz, exchangeRates, addrs[0], addrs[0], true},
		{[]byte("0\x01"), exchangeRates, addrs[0], addrs[0], false},
		{bz[1:], exchangeRates, addrs[0], addrs[0], false},
		{bz, exchangeRates, sdk.AccAddress{}, addrs[0], false},
		{AggregateVoteHash{}, exchangeRates, addrs[0], addrs[0], false},
		{bz, exchangeRates, addrs[0], sdk.AccAddress{}, false},
	}

	for i, tc := range tests {
		msg := NewMsgAggregateExchangeRatePrevote(tc.hash, tc.feeder, sdk.ValAddress(tc.validator))
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgAggregateExchangeRateVote(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
	}

	invalidExchangeRates := "a,b"
	exchangeRates := "foo:1.0,bar:1232.123"
	zeroExchangeRates := "foo:0.0,bar:1232.132"
	negativeExchangeRates := "foo:-1234.5,bar:1232.132"
	overFlowMsgExchangeRates := StringWithCharset(4097, "56432")
	overFlowExchangeRates := "foo:100000000000000000000000000000000000000000000000000000000000000000000000000000.01,bar:1232.132"
	validSalt := "0cf33fb528b388660c3a42c3f3250e983395290b75fef255050fb5bc48a6025f"
	saltWithColon := "0cf33fb528b388660c3a42c3f3250e983395290b75fef255050fb5bc48a6025:"
	tests := []struct {
		feeder        sdk.AccAddress
		validator     sdk.AccAddress
		salt          string
		exchangeRates string
		expectPass    bool
	}{
		{addrs[0], addrs[0], validSalt, exchangeRates, true},
		{addrs[0], addrs[0], validSalt, invalidExchangeRates, false},
		{addrs[0], addrs[0], validSalt, zeroExchangeRates, false},
		{addrs[0], addrs[0], validSalt, negativeExchangeRates, false},
		{addrs[0], addrs[0], validSalt, overFlowMsgExchangeRates, false},
		{addrs[0], addrs[0], validSalt, overFlowExchangeRates, false},
		{sdk.AccAddress{}, sdk.AccAddress{}, validSalt, exchangeRates, false},
		{addrs[0], sdk.AccAddress{}, validSalt, exchangeRates, false},
		{addrs[0], addrs[0], "", exchangeRates, false},
		{addrs[0], addrs[0], validSalt, "", false},
		{addrs[0], addrs[0], saltWithColon, exchangeRates, false},
	}

	for i, tc := range tests {
		msg := NewMsgAggregateExchangeRateVote(tc.salt, tc.exchangeRates, tc.feeder, sdk.ValAddress(tc.validator))
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestNewMsgAggregateExchangeRatePrevote(t *testing.T) {
	vals := GenerateRandomValAddr(2)
	feederAddr := sdk.AccAddress(vals[1])

	exchangeRates := sdk.DecCoins{sdk.NewDecCoinFromDec(UmeeDenom, sdk.OneDec()), sdk.NewDecCoinFromDec(UmeeDenom, sdk.NewDecWithPrec(32121, 1))}
	bz := GetAggregateVoteHash("1", exchangeRates.String(), sdk.ValAddress(vals[0]))

	aggregateExchangeRatePreVote := NewMsgAggregateExchangeRatePrevote(
		bz,
		feederAddr,
		vals[0],
	)

	require.Equal(t, aggregateExchangeRatePreVote.Route(), RouterKey)
	require.Equal(t, aggregateExchangeRatePreVote.Type(), TypeMsgAggregateExchangeRatePrevote)
	require.NotNil(t, aggregateExchangeRatePreVote.GetSignBytes())
	require.Equal(t, aggregateExchangeRatePreVote.GetSigners(), []sdk.AccAddress{feederAddr})
}

func TestNewMsgAggregateExchangeRateVote(t *testing.T) {
	vals := GenerateRandomValAddr(2)
	feederAddr := sdk.AccAddress(vals[1])

	aggregateExchangeRateVote := NewMsgAggregateExchangeRateVote(
		"salt",
		"0.1",
		feederAddr,
		vals[0],
	)

	require.Equal(t, aggregateExchangeRateVote.Route(), RouterKey)
	require.Equal(t, aggregateExchangeRateVote.Type(), TypeMsgAggregateExchangeRateVote)
	require.NotNil(t, aggregateExchangeRateVote.GetSignBytes())
	require.Equal(t, aggregateExchangeRateVote.GetSigners(), []sdk.AccAddress{feederAddr})
}

func TestMsgDelegateFeedConsent(t *testing.T) {
	vals := GenerateRandomValAddr(2)
	msgFeedConsent := NewMsgDelegateFeedConsent(vals[0], sdk.AccAddress(vals[1]))

	require.Equal(t, msgFeedConsent.Route(), RouterKey)
	require.Equal(t, msgFeedConsent.Type(), TypeMsgDelegateFeedConsent)
	require.NotNil(t, msgFeedConsent.GetSignBytes())
	require.Equal(t, msgFeedConsent.GetSigners(), []sdk.AccAddress{sdk.AccAddress(vals[0])})
}
