package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/ante"
	oracletypes "github.com/umee-network/umee/x/oracle/types"
)

func (suite *IntegrationTestSuite) TestMempoolFee() {
	suite.SetupTest()
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewMempoolFeeDecorator()
	antehandler := sdk.ChainAnteDecorators(mfd)
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	// Set high gas price so standard test fee fails
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	highGasPrice := []sdk.DecCoin{atomPrice}
	suite.ctx = suite.ctx.WithMinGasPrices(highGasPrice)

	suite.ctx = suite.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Error(err, "Decorator should have errored on too low fee for local gasPrice")

	// ensure no fees for oracle msgs
	suite.Require().NoError(suite.txBuilder.SetMsgs(
		oracletypes.NewMsgAggregateExchangeRatePrevote(oracletypes.AggregateVoteHash{}, addr1, sdk.ValAddress(addr1)),
		oracletypes.NewMsgAggregateExchangeRateVote("", "", addr1, sdk.ValAddress(addr1)),
	))
	_, err = suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
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
