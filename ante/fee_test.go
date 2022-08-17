package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v2/ante"
	oracletypes "github.com/umee-network/umee/v2/x/oracle/types"
)

func (suite *IntegrationTestSuite) TestFeeAndPriority() {
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	require := suite.Require()

	msgs := testdata.NewTestMsg(addr1)
	require.NoError(suite.txBuilder.SetMsgs(msgs))
	fee := testdata.NewTestFeeAmount()
	suite.txBuilder.SetFeeAmount(fee)
	suite.txBuilder.SetGasLimit(testdata.NewTestGasLimit())

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	require.NoError(err)

	// Test1: validator min gas price check
	// set min gas price above the transaction gas limit, so the tx should fail
	minGasPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	ctx := suite.ctx.
		WithMinGasPrices([]sdk.DecCoin{minGasPrice}).
		WithIsCheckTx(true)
	_, _, err = ante.FeeAndPriority(ctx, tx)
	require.ErrorIs(sdkerrors.ErrInsufficientFee, err)

	// Test2: min gas price not checked in DeliverTx
	ctx = suite.ctx.WithIsCheckTx(false)
	requiredFee, _, err := ante.FeeAndPriority(ctx, tx)
	require.NoError(err)
	require.True(fee.IsEqual(requiredFee))

	// Test3: should not error when min gas price is same or lower than the fee
	ctx = ctx.WithMinGasPrices(sdk.NewDecCoinsFromCoins(fee...))
	requiredFee, _, err = ante.FeeAndPriority(ctx, tx)
	require.NoError(err)
	require.True(fee.IsEqual(requiredFee))

	ctx = ctx.WithMinGasPrices([]sdk.DecCoin{sdk.NewDecCoin(fee[0].Denom, fee[0].Amount.QuoRaw(2))})
	requiredFee, _, err = ante.FeeAndPriority(ctx, tx)
	require.NoError(err)
	require.True(fee.IsEqual(requiredFee))

	// Test4: ensure no fees for oracle msgs
	require.NoError(suite.txBuilder.SetMsgs(
		oracletypes.NewMsgAggregateExchangeRatePrevote(oracletypes.AggregateVoteHash{}, addr1, sdk.ValAddress(addr1)),
		oracletypes.NewMsgAggregateExchangeRateVote("", "", addr1, sdk.ValAddress(addr1)),
	))
	oracleTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	require.NoError(err)
	require.True(ante.IsOracleTx(oracleTx.GetMsgs()))

	ctx = suite.ctx.WithIsCheckTx(true)
	requiredFee, _, err = ante.FeeAndPriority(ctx, oracleTx)
	require.NoError(err, "Decorator should not require fees for oracle tx")
	require.True(requiredFee.IsZero(), "fee should be zero, got: %s", requiredFee)
}
