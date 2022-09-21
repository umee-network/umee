package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	signing "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/umee-network/umee/v3/ante"
	appparams "github.com/umee-network/umee/v3/app/params"
	"github.com/umee-network/umee/v3/util/coin"
	oracletypes "github.com/umee-network/umee/v3/x/oracle/types"
)

func (suite *IntegrationTestSuite) TestFeeAndPriority() {
	require := suite.Require()
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	msgs := testdata.NewTestMsg(addr1)
	require.NoError(suite.txBuilder.SetMsgs(msgs))
	minGas := appparams.ProtocolMinGasPrice
	mkFee := func(factor string) sdk.Coins {
		return coin.NewDecBld(minGas).Scale(int64(appparams.DefaultGasLimit)).ScaleStr(factor).ToCoins()
	}
	mkGas := func(denom, factor string) sdk.DecCoins {
		if denom == "" {
			denom = minGas.Denom
		}
		f := sdk.MustNewDecFromStr(factor)
		return sdk.DecCoins{sdk.NewDecCoinFromDec(denom, minGas.Amount.Mul(f))}
	}
	mkTx := func(fee sdk.Coins) signing.Tx {
		suite.txBuilder.SetFeeAmount(fee)
		tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
		require.NoError(err)
		return tx
	}

	suite.txBuilder.SetGasLimit(appparams.DefaultGasLimit)
	// we set fee to 2*gasLimit*minGasPrice
	fee := mkFee("2")
	tx := mkTx(fee)

	//
	// Test CheckTX
	//
	ctx := suite.ctx.
		WithMinGasPrices(sdk.DecCoins{minGas}).
		WithIsCheckTx(true)

	// min-gas-settings should work
	suite.checkFeeAnte(tx, fee, ctx)

	// should work when exact fee is provided
	suite.checkFeeAnte(tx, fee, ctx.WithMinGasPrices(mkGas("", "2")))

	// TODO: comment back in when min gas price is nonzero

	// should fail when not enough fee is provided
	// suite.checkFeeFailed(tx, ctx.WithMinGasPrices(mkGas("", "3")))

	// should fail when other denom is required
	// suite.checkFeeFailed(tx, ctx.WithMinGasPrices(mkGas("other", "1")))

	// should fail when some fee doesn't include all gas denoms
	// ctx = ctx.WithMinGasPrices(sdk.DecCoins{minGas,sdk.NewDecCoinFromDec("other", sdk.NewDec(10))})
	// suite.checkFeeFailed(tx, ctx)

	//
	// Test DeliverTx
	//
	ctx = suite.ctx.
		WithMinGasPrices(sdk.DecCoins{minGas}).
		WithIsCheckTx(false)

	// ctx.MinGasPrice shouldn't matter
	suite.checkFeeAnte(tx, fee, ctx)
	suite.checkFeeAnte(tx, fee, ctx.WithMinGasPrices(mkGas("", "3")))
	suite.checkFeeAnte(tx, fee, ctx.WithMinGasPrices(mkGas("other", "1")))
	suite.checkFeeAnte(tx, fee, ctx.WithMinGasPrices(sdk.DecCoins{}))

	// should fail when not enough fee is provided
	// suite.checkFeeFailed(mkTx(mkFee("0.5")), ctx)
	// suite.checkFeeFailed(mkTx(sdk.Coins{}), ctx)

	// should  work when more fees are applied
	fee = append(fee, sdk.NewInt64Coin("other", 10))
	suite.checkFeeAnte(mkTx(fee), fee, ctx)

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

func (suite *IntegrationTestSuite) checkFeeFailed(tx sdk.Tx, ctx sdk.Context) {
	_, _, err := ante.FeeAndPriority(ctx, tx)
	suite.Require().ErrorIs(err, sdkerrors.ErrInsufficientFee)
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
