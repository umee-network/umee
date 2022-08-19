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
	require := suite.Require()
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	msgs := testdata.NewTestMsg(addr1)
	require.NoError(suite.txBuilder.SetMsgs(msgs))
	fee := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	gasLimit := 200000
	suite.txBuilder.SetFeeAmount(fee)
	suite.txBuilder.SetGasLimit(uint64(gasLimit))

	// Test1: validator min gas price check
	// Ante should fail when validator min gas price is above the transaction gas limit
	minGasPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDecFromInt(fee[0].Amount).QuoInt64(int64(gasLimit/2)))
	ctx := suite.ctx.
		WithMinGasPrices([]sdk.DecCoin{minGasPrice}).
		WithIsCheckTx(true)
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	require.NoError(err)
	_, _, err = ante.FeeAndPriority(ctx, tx)
	require.ErrorIs(sdkerrors.ErrInsufficientFee, err)

	// Test2: min gas price not checked in DeliverTx
	ctx = suite.ctx.WithIsCheckTx(false)
	suite.checkFeeAnte(tx, fee, ctx)

	// Test3: should not error when min gas price is same or lower than the fee
	ctx = ctx.WithMinGasPrices(sdk.NewDecCoinsFromCoins(fee...))
	suite.checkFeeAnte(tx, fee, ctx)

	ctx = ctx.WithMinGasPrices([]sdk.DecCoin{sdk.NewDecCoin(fee[0].Denom, fee[0].Amount.QuoRaw(2))})
	suite.checkFeeAnte(tx, fee, ctx)

	// Test4: ensure no fees for oracle msgs
	require.NoError(suite.txBuilder.SetMsgs(
		oracletypes.NewMsgAggregateExchangeRatePrevote(oracletypes.AggregateVoteHash{}, addr1, sdk.ValAddress(addr1)),
		oracletypes.NewMsgAggregateExchangeRateVote("", "", addr1, sdk.ValAddress(addr1)),
	))
	suite.txBuilder.SetFeeAmount(sdk.Coins{})
	suite.txBuilder.SetGasLimit(0)
	oracleTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	require.NoError(err)
	require.True(oracleTx.GetFee().IsZero(), "got: %s", oracleTx.GetFee())
	require.Equal(uint64(0), oracleTx.GetGas())
	require.True(ante.IsOracleOrGravityTx(oracleTx.GetMsgs()))

	suite.checkFeeAnte(oracleTx, sdk.Coins{}, suite.ctx.WithIsCheckTx(true))
	suite.checkFeeAnte(oracleTx, sdk.Coins{}, suite.ctx.WithIsCheckTx(false))
}

func (suite *IntegrationTestSuite) checkFeeAnte(tx sdk.Tx, feeExpected sdk.Coins, ctx sdk.Context) {
	require := suite.Require()
	fee, _, err := ante.FeeAndPriority(ctx, tx)
	require.NoError(err)
	if len(feeExpected) == 0 {
		require.True(fee.IsZero(), "fee should be zero, got: %s", fee)
	} else {
		require.True(fee.IsEqual(feeExpected), "Fee expected %s, got: %s", feeExpected, fee)
	}
}
