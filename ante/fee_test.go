package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	oracletypes "github.com/umee-network/umee/v2/x/oracle/types"
)

func (suite *IntegrationTestSuite) TestMempoolFee() {
	suite.SetupTest()
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
	suite.Require().NoError(err)

	// Test2: min gas price not checked in DeliverTx
	ctx = suite.ctx.WithIsCheckTx(false)
	suite.checkFeeAnte(tx, fee, ctx)

	// Test3: should not error when min gas price is same or lower than the fee
	ctx = ctx.WithMinGasPrices(sdk.NewDecCoinsFromCoins(fee...))
	suite.checkFeeAnte(tx, fee, ctx)

	ctx = ctx.WithMinGasPrices([]sdk.DecCoin{sdk.NewDecCoin(fee[0].Denom, fee[0].Amount.QuoRaw(2))})
	suite.checkFeeAnte(tx, fee, ctx)

	// ensure no fees for oracle msgs
	suite.Require().NoError(suite.txBuilder.SetMsgs(
		oracletypes.NewMsgAggregateExchangeRatePrevote(oracletypes.AggregateVoteHash{}, addr1, sdk.ValAddress(addr1)),
		oracletypes.NewMsgAggregateExchangeRateVote("", "", addr1, sdk.ValAddress(addr1)),
	))
	suite.txBuilder.SetFeeAmount(sdk.Coins{})
	suite.txBuilder.SetGasLimit(0)
	oracleTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	_, err = antehandler(suite.ctx, oracleTx, false)
	suite.Require().NoError(err, "Decorator should not require high price for oracle tx")

	suite.ctx = suite.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NoError(err, "MempoolFeeDecorator returned error in DeliverTx")

	suite.ctx = suite.ctx.WithIsCheckTx(true)

	atomPrice = sdk.NewDecCoinFromDec("atom", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	suite.ctx = suite.ctx.WithMinGasPrices(lowGasPrice)

	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NoError(err, "Decorator should not have errored on fee higher than local gasPrice")
}
