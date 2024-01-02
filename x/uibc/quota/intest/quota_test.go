package intest

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"gotest.tools/v3/assert"

	lfixtures "github.com/umee-network/umee/v6/x/leverage/fixtures"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/oracle/types"
	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/mocks"
)

func TestResetQuota(t *testing.T) {
	s := initTestSuite(t)
	k := s.app.UIbcQuotaKeeperB.Keeper(&s.ctx)

	umeeQuota := sdk.NewInt64DecCoin("uumee", 1000)
	k.SetTokenOutflow(umeeQuota)
	q := k.GetTokenOutflows(umeeQuota.Denom)
	assert.DeepEqual(t, q, umeeQuota)

	k.SetTokenInflow(umeeQuota)
	i := k.GetTokenInflow(umeeQuota.Denom)
	assert.DeepEqual(t, i, umeeQuota)

	// reset the quota
	k.ResetAllQuotas()

	// check the quota after reset
	q = k.GetTokenOutflows(umeeQuota.Denom)
	assert.DeepEqual(t, q.Amount, sdk.NewDec(0))
	i = k.GetTokenInflow(umeeQuota.Denom)
	assert.DeepEqual(t, i.Amount, sdk.NewDec(0))
}

func TestKeeper_CheckAndUpdateQuota(t *testing.T) {
	invalidToken := sdk.NewCoin("u/u/umee", sdkmath.NewInt(100))
	umeeUToken := sdk.NewCoin("u/umee", sdkmath.NewInt(100))
	atomToken := sdk.NewCoin("atom", sdkmath.NewInt(1000))
	daiToken := sdk.NewCoin("dai", sdkmath.NewInt(50))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leverageMock := mocks.NewMockLeverage(ctrl)
	oracleMock := mocks.NewMockOracle(ctrl)

	marshaller := codec.NewProtoCodec(nil)
	ctx, k := initKeeper(t, marshaller, leverageMock, oracleMock)
	err := k.ResetAllQuotas()
	assert.NilError(t, err)

	// invalid token, returns error from mock leverage
	leverageMock.EXPECT().ToToken(ctx, invalidToken).Return(sdk.Coin{}, ltypes.ErrNotUToken).AnyTimes()

	err = k.CheckAndUpdateQuota(invalidToken.Denom, invalidToken.Amount)
	assert.ErrorIs(t, err, ltypes.ErrNotUToken)

	// UMEE uToken, exchanges correctly, but returns ErrNotRegisteredToken when trying to get Token's settings
	// from leverage mock keeper
	leverageMock.EXPECT().ToToken(ctx, umeeUToken).Return(
		sdk.NewCoin("umee", sdkmath.NewInt(100)),
		nil,
	).AnyTimes()
	leverageMock.EXPECT().GetTokenSettings(ctx, "umee").Return(ltypes.Token{}, ltypes.ErrNotRegisteredToken).AnyTimes()

	err = k.CheckAndUpdateQuota(umeeUToken.Denom, umeeUToken.Amount)
	// returns nil when the error is ErrNotRegisteredToken
	assert.NilError(t, err)

	// ATOM, returns token settings correctly from leverage mock keeper,
	// then returns an error when trying to get token prices from oracle mock keeper
	leverageMock.EXPECT().GetTokenSettings(ctx, "atom").Return(
		lfixtures.Token("atom", "ATOM", 6), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "ATOM").Return(sdk.Dec{}, types.ErrMalformedLatestAvgPrice)

	err = k.CheckAndUpdateQuota(atomToken.Denom, atomToken.Amount)
	assert.ErrorIs(t, err, types.ErrMalformedLatestAvgPrice)

	// DAI returns token settings and prices from mock leverage and oracle keepers, no errors expected
	leverageMock.EXPECT().GetTokenSettings(ctx, "dai").Return(
		lfixtures.Token("dai", "DAI", 6), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "DAI").Return(sdk.MustNewDecFromStr("0.37"), nil)

	err = k.SetParams(uibc.DefaultParams())
	assert.NilError(t, err)

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("dai", 10000)}
	k.SetTokenOutflows(setQuotas)

	err = k.CheckAndUpdateQuota(daiToken.Denom, daiToken.Amount)
	assert.NilError(t, err)
}

func TestKeeper_UndoUpdateQuota(t *testing.T) {
	umeeAmount := sdkmath.NewInt(100_000000)
	umeePrice := sdk.MustNewDecFromStr("0.37")
	umeeQuota := sdkmath.NewInt(10000)
	atomToken := sdk.NewCoin("umee", umeeAmount)
	umeeExponent := 6

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leverageMock := mocks.NewMockLeverage(ctrl)
	oracleMock := mocks.NewMockOracle(ctrl)

	marshaller := codec.NewProtoCodec(nil)
	ctx, k := initKeeper(t, marshaller, leverageMock, oracleMock)
	err := k.ResetAllQuotas()
	assert.NilError(t, err)

	// UMEE, returns token settings and prices from mock leverage and oracle keepers, no errors expected
	leverageMock.EXPECT().GetTokenSettings(ctx, "umee").Return(
		lfixtures.Token("umee", "UMEE", uint32(umeeExponent)), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "UMEE").Return(umeePrice, nil).AnyTimes()

	err = k.UndoUpdateQuota(atomToken.Denom, atomToken.Amount)
	// the result is ignored due to quota reset
	assert.NilError(t, err)

	o := k.GetTokenOutflows(atomToken.Denom)
	assert.DeepEqual(t, o.Amount, sdk.ZeroDec())

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("umee", umeeQuota.Int64())}
	k.SetTokenOutflows(setQuotas)

	err = k.UndoUpdateQuota(atomToken.Denom, atomToken.Amount)
	assert.NilError(t, err)

	o = k.GetTokenOutflows(atomToken.Denom)

	// the expected quota is calculated as follows:
	// umee_value = umee_amount * umee_price
	// expected_quota = current_quota - umee_value
	powerReduction := sdk.MustNewDecFromStr("10").Power(uint64(umeeExponent))
	expectedQuota := sdk.NewDec(umeeQuota.Int64()).Sub(sdk.NewDecFromInt(atomToken.Amount).Quo(powerReduction).Mul(umeePrice))
	assert.DeepEqual(t, o.Amount, expectedQuota)
}

func TestKeeper_RecordIBCInflow(t *testing.T) {
	atomAmount := sdkmath.NewInt(100_000000)
	atomPrice := sdk.MustNewDecFromStr("0.37")
	// ibc incoming denom from packet is `port/path/base_denom`
	atomToken := sdk.NewCoin("transfer/channel-10/uatom", atomAmount)
	atomExponent := 6
	inflowBaseDenom := "ibc/D6372674F0E9A3A7ADC2FEFD8B2708C5008C7ED04DA6566E279DC1321BDDCB6F"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leverageMock := mocks.NewMockLeverage(ctrl)
	oracleMock := mocks.NewMockOracle(ctrl)

	marshaller := codec.NewProtoCodec(nil)
	ctx, k := initKeeper(t, marshaller, leverageMock, oracleMock)
	err := k.ResetAllQuotas()
	assert.NilError(t, err)

	// ATOM, returns token settings and prices from mock leverage and oracle keepers, no errors expected
	leverageMock.EXPECT().GetTokenSettings(ctx, inflowBaseDenom).Return(
		lfixtures.Token(inflowBaseDenom, "ATOM", uint32(atomExponent)), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "ATOM").Return(atomPrice, nil).AnyTimes()

	packet := channeltypes.Packet{
		Sequence:           10,
		SourcePort:         "transfer",
		DestinationPort:    "transfer",
		SourceChannel:      "channel-10",
		DestinationChannel: "channel-1",
		Data:               nil,
	}

	ackErr := k.RecordIBCInflow(packet, atomToken.Denom, atomToken.Amount.String())
	assert.Assert(t, nil, ackErr)

	o := k.GetTokenInflow(inflowBaseDenom)
	// expected inflow amount 37 =( atomPrice * atomAmount) / atomExponent
	inflowAmount := sdkmath.LegacyMustNewDecFromStr("37")
	assert.DeepEqual(t, o.Amount, inflowAmount)

	ackErr = k.RecordIBCInflow(packet, atomToken.Denom, atomToken.Amount.String())
	assert.Assert(t, nil, ackErr)

	o = k.GetTokenInflow(inflowBaseDenom)
	assert.DeepEqual(t, o.Amount, inflowAmount.Add(inflowAmount))

	allInflows, err := k.GetAllInflows()
	assert.NilError(t, err)
	assert.DeepEqual(t, allInflows, sdk.DecCoins{o})

	err = k.ResetAllQuotas()
	assert.NilError(t, err)

	o = k.GetTokenInflow(inflowBaseDenom)
	assert.DeepEqual(t, o.Amount, sdk.ZeroDec())
}
