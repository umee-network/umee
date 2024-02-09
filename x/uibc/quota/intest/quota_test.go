package intest

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"gotest.tools/v3/assert"

	appparams "github.com/umee-network/umee/v6/app/params"
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
	assert.DeepEqual(t, q.Amount, sdkmath.LegacyNewDec(0))
	i = k.GetTokenInflow(umeeQuota.Denom)
	assert.DeepEqual(t, i.Amount, sdkmath.LegacyNewDec(0))
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
	oracleMock.EXPECT().Price(ctx, "ATOM").Return(sdkmath.LegacyDec{}, types.ErrMalformedLatestAvgPrice)

	err = k.CheckAndUpdateQuota(atomToken.Denom, atomToken.Amount)
	assert.ErrorIs(t, err, types.ErrMalformedLatestAvgPrice)

	// DAI returns token settings and prices from mock leverage and oracle keepers, no errors expected
	leverageMock.EXPECT().GetTokenSettings(ctx, "dai").Return(
		lfixtures.Token("dai", "DAI", 6), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "DAI").Return(sdkmath.LegacyMustNewDecFromStr("0.37"), nil)

	err = k.SetParams(uibc.DefaultParams())
	assert.NilError(t, err)

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("dai", 10000)}
	k.SetTokenOutflows(setQuotas)

	err = k.CheckAndUpdateQuota(daiToken.Denom, daiToken.Amount)
	assert.NilError(t, err)
}

func TestKeeper_UndoUpdateQuota(t *testing.T) {
	umeeAmount := sdkmath.NewInt(100_000000)
	umeePrice := sdkmath.LegacyMustNewDecFromStr("0.37")
	umeeQuota := sdkmath.NewInt(10000)
	umeeToken := sdk.NewCoin("umee", umeeAmount)
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

	err = k.UndoUpdateQuota(umeeToken.Denom, umeeToken.Amount)
	// the result is ignored due to quota reset
	assert.NilError(t, err)

	o := k.GetTokenOutflows(umeeToken.Denom)
	assert.DeepEqual(t, o.Amount, sdkmath.LegacyZeroDec())

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("umee", umeeQuota.Int64())}
	k.SetTokenOutflows(setQuotas)

	err = k.UndoUpdateQuota(umeeToken.Denom, umeeToken.Amount)
	assert.NilError(t, err)

	o = k.GetTokenOutflows(umeeToken.Denom)

	// the expected quota is calculated as follows:
	// umee_value = umee_amount * umee_price
	// expected_quota = current_quota - umee_value
	powerReduction := sdkmath.LegacyMustNewDecFromStr("10").Power(uint64(umeeExponent))
	expectedQuota := sdkmath.LegacyNewDec(umeeQuota.Int64()).Sub(sdkmath.LegacyNewDecFromInt(umeeToken.Amount).Quo(powerReduction).Mul(umeePrice))
	assert.DeepEqual(t, o.Amount, expectedQuota)
}

func TestKeeper_RecordIBCInflow(t *testing.T) {
	tokenAmount := sdkmath.NewInt(100_000000)
	tokenPrice := sdkmath.LegacyMustNewDecFromStr("0.37")
	tokenExponent := 6
	// ibc denom = base_denom when sender chain is source chain
	atomToken := sdk.NewCoin("uatom", tokenAmount)
	// ibc denom = (port/channel/base_denom) when we receive token back which is send from UMEE
	umeeNativeToken := sdk.NewCoin("transfer/channel-10/uumee", tokenAmount)
	atomIBCDenom := "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leverageMock := mocks.NewMockLeverage(ctrl)
	oracleMock := mocks.NewMockOracle(ctrl)

	marshaller := codec.NewProtoCodec(nil)
	ctx, k := initKeeper(t, marshaller, leverageMock, oracleMock)
	err := k.ResetAllQuotas()
	assert.NilError(t, err)

	// ATOM returns token settings and prices from mock leverage and oracle keepers, no errors expected
	leverageMock.EXPECT().GetTokenSettings(ctx, atomIBCDenom).Return(
		lfixtures.Token(atomIBCDenom, "ATOM", uint32(tokenExponent)), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "ATOM").Return(tokenPrice, nil).AnyTimes()
	// UMEE
	leverageMock.EXPECT().GetTokenSettings(ctx, appparams.BondDenom).Return(
		lfixtures.Token(appparams.BondDenom, appparams.DisplayDenom, uint32(tokenExponent)), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, appparams.DisplayDenom).Return(tokenPrice, nil).AnyTimes()

	packet := channeltypes.Packet{
		Sequence:           10,
		SourcePort:         "transfer",
		DestinationPort:    "transfer",
		SourceChannel:      "channel-10",
		DestinationChannel: "channel-1",
		Data:               nil,
	}

	// A -> B (record if B is receiver chain)
	ackErr := k.RecordIBCInflow(packet, atomToken.Denom, atomToken.Amount.String())
	assert.Assert(t, nil, ackErr)
	// A -> B (record if B is source chain, we are receiving token back which is send from A chain)
	ackErr = k.RecordIBCInflow(packet, umeeNativeToken.Denom, umeeNativeToken.Amount.String())
	assert.Assert(t, nil, ackErr)

	o := k.GetTokenInflow(atomIBCDenom)
	// expected inflow amount 37 =( atomPrice * atomAmount) / atomExponent
	inflowAmount := sdkmath.LegacyMustNewDecFromStr("37")
	assert.DeepEqual(t, o.Amount, inflowAmount)

	ackErr = k.RecordIBCInflow(packet, atomToken.Denom, atomToken.Amount.String())
	assert.Assert(t, nil, ackErr)

	o = k.GetTokenInflow(atomIBCDenom)
	assert.DeepEqual(t, o.Amount, inflowAmount.Add(inflowAmount))

	p := k.GetTokenInflow(appparams.BondDenom)
	assert.DeepEqual(t, p.Amount, inflowAmount)

	allInflows, err := k.GetAllInflows()
	assert.NilError(t, err)
	assert.DeepEqual(t, allInflows, sdk.DecCoins{o, p})

	err = k.ResetAllQuotas()
	assert.NilError(t, err)

	o = k.GetTokenInflow(atomIBCDenom)
	assert.DeepEqual(t, o.Amount, sdkmath.LegacyZeroDec())
}
