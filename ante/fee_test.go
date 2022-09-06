package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/ante"
	appparams "github.com/umee-network/umee/v3/app/params"
	oracletypes "github.com/umee-network/umee/v3/x/oracle/types"
)

func (suite *IntegrationTestSuite) TestFeeAndPriority() {
	require := suite.Require()
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	msgs := testdata.NewTestMsg(addr1)
	require.NoError(suite.txBuilder.SetMsgs(msgs))
	gasLimit := 200000
	minG := appparams.MinMinGasPrice
	mkFee := func(factor string) sdk.Coins {
		f := sdk.MustNewDecFromStr(factor)
		return sdk.NewCoins(sdk.NewCoin(appparams.BondDenom,
			minG.Amount.MulInt64(int64(gasLimit)).Mul(f).Ceil().RoundInt()))
	}
	mkGas := func(denom, factor string) sdk.DecCoins {
		if denom == "" {
			denom = minG.Denom
		}
		f := sdk.MustNewDecFromStr(factor)
		return sdk.DecCoins{sdk.NewDecCoinFromDec(denom, minG.Amount.Mul(f))}
	}

	// we set fee to 2*gasLimit*minGasPrice
	fee := mkFee("2")
	suite.txBuilder.SetFeeAmount(fee)
	suite.txBuilder.SetGasLimit(uint64(gasLimit))
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	require.NoError(err)

	// Test CheckTX //

	// min-gas-settings should work
	ctx := suite.ctx.
		WithMinGasPrices(sdk.DecCoins{appparams.MinMinGasPrice}).
		WithIsCheckTx(true)
	suite.checkFeeAnte(tx, fee, ctx)

	// should work when exact fee is provided
	suite.checkFeeAnte(tx, fee, ctx.WithMinGasPrices(mkGas("", "2")))

	// should fail when not enough fee is provided
	suite.checkFeeFailed(tx, ctx.WithMinGasPrices(mkGas("", "3")))

	// should fail when other denom is required
	suite.checkFeeFailed(tx, ctx.WithMinGasPrices(mkGas("other", "1")))

	// should fail when some fee doesn't include all gas denoms
	ctx = ctx.WithMinGasPrices(sdk.DecCoins{appparams.MinMinGasPrice,
		sdk.NewDecCoinFromDec("other", sdk.NewDec(10))})
	_, _, err = ante.FeeAndPriority(ctx, tx)
	require.ErrorIs(sdkerrors.ErrInsufficientFee, err)

	/*
		// Test2: validator min gas price check during the
		// Ante should fail when validator min gas price is above the transaction gas limit
		ctx = ctx.WithMinGasPrices(sdk.NewDecCoinsFromCoins(fee...))
		_, _, err = ante.FeeAndPriority(ctx, tx)
		require.ErrorIs(sdkerrors.ErrInsufficientFee, err)

		// Test2: min gas price not checked in DeliverTx
		ctx = suite.ctx.WithIsCheckTx(false)
		suite.checkFeeAnte(tx, fee, ctx)

		ctx = ctx.WithMinGasPrices([]sdk.DecCoin{sdk.NewDecCoin(fee[0].Denom, fee[0].Amount.QuoRaw(2))})
		suite.checkFeeAnte(tx, fee, ctx)


	*/

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
	suite.Require().ErrorIs(sdkerrors.ErrInsufficientFee, err)
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
