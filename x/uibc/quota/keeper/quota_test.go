package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	ctypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/umee-network/umee/v4/x/uibc"

	"github.com/umee-network/umee/v4/x/oracle/types"

	lfixtures "github.com/umee-network/umee/v4/x/leverage/fixtures"

	sdkmath "cosmossdk.io/math"

	ltypes "github.com/umee-network/umee/v4/x/leverage/types"

	"github.com/golang/mock/gomock"
	"github.com/umee-network/umee/v4/x/uibc/mocks"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestResetQuota(t *testing.T) {
	s := initKeeperTestSuite(t)
	ctx, k := s.ctx, s.app.UIbcQuotaKeeper

	umeeQuota := sdk.NewInt64DecCoin("uumee", 1000)
	k.SetTokenOutflow(ctx, umeeQuota)
	q, err := k.GetTokenOutflows(ctx, umeeQuota.Denom)
	assert.NilError(t, err)
	assert.DeepEqual(t, q, umeeQuota)

	// reset the quota
	k.ResetAllQuotas(ctx)

	// check the quota after reset
	q, err = k.GetTokenOutflows(ctx, umeeQuota.Denom)
	assert.NilError(t, err)
	assert.DeepEqual(t, q.Amount, sdk.NewDec(0))
}

func TestKeeper_CheckAndUpdateQuota(t *testing.T) {
	invalidToken := sdk.NewCoin("u/u/umee", sdkmath.NewInt(100))
	umeeUToken := sdk.NewCoin("u/umee", sdkmath.NewInt(100))
	atomToken := sdk.NewCoin("atom", sdkmath.NewInt(1000))
	daiToken := sdk.NewCoin("dai", sdkmath.NewInt(50))
	// gomock initializations
	leverageCtrl := gomock.NewController(t)
	defer leverageCtrl.Finish()
	leverageMock := mocks.NewMockLeverageKeeper(leverageCtrl)

	oracleCtrl := gomock.NewController(t)
	defer oracleCtrl.Finish()
	oracleMock := mocks.NewMockOracle(oracleCtrl)

	interfaceRegistry := ctypes.NewInterfaceRegistry()
	marshaller := codec.NewProtoCodec(interfaceRegistry)
	ctx, k := initFullKeeper(t, marshaller, nil, leverageMock, oracleMock)
	err := k.ResetAllQuotas(ctx)
	assert.NilError(t, err)

	// invalid token, returns error from mock leverage
	leverageMock.EXPECT().ExchangeUToken(ctx, invalidToken).Return(sdk.Coin{}, ltypes.ErrNotUToken).AnyTimes()

	err = k.CheckAndUpdateQuota(ctx, invalidToken.Denom, invalidToken.Amount)
	assert.ErrorIs(t, err, ltypes.ErrNotUToken)

	// UMEE uToken, exchanges correctly, but returns ErrNotRegisteredToken when trying to get Token's settings
	// from leverage mock keeper
	leverageMock.EXPECT().ExchangeUToken(ctx, umeeUToken).Return(
		sdk.NewCoin("umee", sdkmath.NewInt(100)),
		nil,
	).AnyTimes()
	leverageMock.EXPECT().GetTokenSettings(ctx, "umee").Return(ltypes.Token{}, ltypes.ErrNotRegisteredToken).AnyTimes()

	err = k.CheckAndUpdateQuota(ctx, umeeUToken.Denom, umeeUToken.Amount)
	// returns nil when the error is ErrNotRegisteredToken
	assert.NilError(t, err)

	// ATOM, returns token settings correctly from leverage mock keeper,
	// then returns an error when trying to get token prices from oracle mock keeper
	leverageMock.EXPECT().GetTokenSettings(ctx, "atom").Return(
		lfixtures.Token("atom", "ATOM", 6), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "ATOM").Return(sdk.Dec{}, types.ErrMalformedLatestAvgPrice)

	err = k.CheckAndUpdateQuota(ctx, atomToken.Denom, atomToken.Amount)
	assert.ErrorIs(t, err, types.ErrMalformedLatestAvgPrice)

	// DAI returns token settings and prices from mock leverage and oracle keepers, no errors expected
	leverageMock.EXPECT().GetTokenSettings(ctx, "dai").Return(
		lfixtures.Token("dai", "DAI", 6), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "DAI").Return(sdk.MustNewDecFromStr("0.37"), nil)

	err = k.SetParams(ctx, uibc.DefaultParams())
	assert.NilError(t, err)

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("dai", 10000)}
	k.SetTokenOutflows(ctx, setQuotas)

	err = k.CheckAndUpdateQuota(ctx, daiToken.Denom, daiToken.Amount)
	assert.NilError(t, err)
}

func TestKeeper_UndoUpdateQuota(t *testing.T) {
	umeeAmount := sdkmath.NewInt(100_000000)
	umeePrice := sdk.MustNewDecFromStr("0.37")
	umeeQuota := sdkmath.NewInt(10000)
	umeeToken := sdk.NewCoin("umee", umeeAmount)
	umeeExponent := 6
	// gomock initializations
	leverageCtrl := gomock.NewController(t)
	defer leverageCtrl.Finish()
	leverageMock := mocks.NewMockLeverageKeeper(leverageCtrl)

	oracleCtrl := gomock.NewController(t)
	defer oracleCtrl.Finish()
	oracleMock := mocks.NewMockOracle(oracleCtrl)

	interfaceRegistry := ctypes.NewInterfaceRegistry()
	marshaller := codec.NewProtoCodec(interfaceRegistry)
	ctx, k := initFullKeeper(t, marshaller, nil, leverageMock, oracleMock)
	err := k.ResetAllQuotas(ctx)
	assert.NilError(t, err)

	// UMEE, returns token settings and prices from mock leverage and oracle keepers, no errors expected
	leverageMock.EXPECT().GetTokenSettings(ctx, "umee").Return(
		lfixtures.Token("umee", "UMEE", uint32(umeeExponent)), nil,
	).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "UMEE").Return(umeePrice, nil).AnyTimes()

	err = k.UndoUpdateQuota(ctx, umeeToken.Denom, umeeToken.Amount)
	// the result is ignored due to quota reset
	assert.NilError(t, err)

	o, err := k.GetTokenOutflows(ctx, umeeToken.Denom)
	assert.NilError(t, err)
	assert.DeepEqual(t, o.Amount, sdk.ZeroDec())

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("umee", umeeQuota.Int64())}
	k.SetTokenOutflows(ctx, setQuotas)

	err = k.UndoUpdateQuota(ctx, umeeToken.Denom, umeeToken.Amount)
	assert.NilError(t, err)

	o, err = k.GetTokenOutflows(ctx, umeeToken.Denom)
	assert.NilError(t, err)

	// the expected quota is calculated as follows:
	// umee_value = umee_amount * umee_price
	// expected_quota = current_quota - umee_value
	powerReduction := sdk.MustNewDecFromStr("10").Power(uint64(umeeExponent))
	expectedQuota := sdk.NewDec(umeeQuota.Int64()).Sub(sdk.NewDecFromInt(umeeToken.Amount).Quo(powerReduction).Mul(umeePrice))
	assert.DeepEqual(t, o.Amount, expectedQuota)
}
