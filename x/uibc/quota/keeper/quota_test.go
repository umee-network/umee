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

	ibcutil "github.com/umee-network/umee/v4/util/ibc"
)

func TestGetQuotas(t *testing.T) {
	ctx, k := initSimpleKeeper(t)

	quotas, err := k.GetAllOutflows(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(quotas), 0)

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("test_uumee", 10000)}

	k.SetOutflows(ctx, setQuotas)
	quotas, err = k.GetAllOutflows(ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, setQuotas, quotas)

	// get the quota of denom
	quota, err := k.GetOutflows(ctx, setQuotas[0].Denom)
	assert.NilError(t, err)
	assert.Equal(t, quota.Denom, setQuotas[0].Denom)
}

func TestGetLocalDenom(t *testing.T) {
	out := ibcutil.GetLocalDenom("umee")
	assert.Equal(t, "umee", out)
}

func TestResetQuota(t *testing.T) {
	s := initKeeperTestSuite(t)
	ctx, k := s.ctx, s.app.UIbcQuotaKeeper

	umeeQuota := sdk.NewInt64DecCoin("uumee", 1000)
	k.SetDenomOutflow(ctx, umeeQuota)
	q, err := k.GetOutflows(ctx, umeeQuota.Denom)
	assert.NilError(t, err)
	assert.DeepEqual(t, q, umeeQuota)

	// reset the quota
	k.ResetAllQuotas(ctx)

	// check the quota after reset
	q, err = k.GetOutflows(ctx, umeeQuota.Denom)
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
	leverageMlk := mocks.NewMockLeverageKeeper(leverageCtrl)

	oracleCtrl := gomock.NewController(t)
	defer oracleCtrl.Finish()
	oracleMlk := mocks.NewMockOracle(oracleCtrl)

	interfaceRegistry := ctypes.NewInterfaceRegistry()
	marshaller := codec.NewProtoCodec(interfaceRegistry)
	ctx, k := initFullKeeper(t, marshaller, nil, leverageMlk, oracleMlk)
	err := k.ResetAllQuotas(ctx)
	assert.NilError(t, err)

	// invalid token
	leverageMlk.EXPECT().ExchangeUToken(ctx, invalidToken).Return(sdk.Coin{}, ltypes.ErrNotUToken).AnyTimes()

	err = k.CheckAndUpdateQuota(ctx, invalidToken.Denom, invalidToken.Amount)
	assert.ErrorIs(t, err, ltypes.ErrNotUToken)

	// uumee
	leverageMlk.EXPECT().ExchangeUToken(ctx, umeeUToken).Return(
		sdk.NewCoin("umee", sdkmath.NewInt(100)),
		nil,
	).AnyTimes()
	leverageMlk.EXPECT().GetTokenSettings(ctx, "umee").Return(ltypes.Token{}, ltypes.ErrNotRegisteredToken).AnyTimes()

	err = k.CheckAndUpdateQuota(ctx, umeeUToken.Denom, umeeUToken.Amount)
	// returns nil when the error is ErrNotRegisteredToken
	assert.NilError(t, err)

	// atom
	leverageMlk.EXPECT().GetTokenSettings(ctx, "atom").Return(
		lfixtures.Token("atom", "ATOM", 6), nil,
	).AnyTimes()
	oracleMlk.EXPECT().Price(ctx, "ATOM").Return(sdk.Dec{}, types.ErrMalformedLatestAvgPrice)

	err = k.CheckAndUpdateQuota(ctx, atomToken.Denom, atomToken.Amount)
	assert.ErrorIs(t, err, types.ErrMalformedLatestAvgPrice)

	// dai
	leverageMlk.EXPECT().GetTokenSettings(ctx, "dai").Return(
		lfixtures.Token("dai", "DAI", 6), nil,
	).AnyTimes()
	oracleMlk.EXPECT().Price(ctx, "DAI").Return(sdk.MustNewDecFromStr("0.37"), nil)

	err = k.SetParams(ctx, uibc.DefaultParams())
	assert.NilError(t, err)

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("dai", 10000)}
	k.SetOutflows(ctx, setQuotas)

	err = k.CheckAndUpdateQuota(ctx, daiToken.Denom, daiToken.Amount)
	assert.NilError(t, err)
}
