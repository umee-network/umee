package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	oracletypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// MaxOracleMsgGasUsage defines the maximum gas allowed for an oracle transaction.
const MaxOracleMsgGasUsage = uint64(100000)

// MempoolFeeDecorator defines a custom Umee AnteHandler decorator that is
// responsible for allowing oracle transactions from oracle feeders to bypass
// the minimum fee CheckTx check. However, if an oracle transaction's gas limit
// is beyond the accepted threshold, the minimum fee check is still applied.
//
// For non-oracle transactions, the minimum fee check is applied.
type MempoolFeeDecorator struct{}

func NewMempoolFeeDecorator() MempoolFeeDecorator {
	return MempoolFeeDecorator{}
}

func (mfd MempoolFeeDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	// Only check for minimum fees if the execution mode is CheckTx and the tx does
	// not contain oracle messages. If the tx does contain oracle messages, it's
	// total gas must be less than or equal to a constant, otherwise minimum fees
	// are checked.
	if ctx.IsCheckTx() && !simulate &&
		!(isOracleTx(msgs) && gas <= uint64(len(msgs))*MaxOracleMsgGasUsage) {
		minGasPrices := ctx.MinGasPrices()
		if !minGasPrices.IsZero() {
			requiredFees := make(sdk.Coins, len(minGasPrices))

			// Determine the required fees by multiplying each required minimum gas
			// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
			glDec := sdk.NewDec(int64(gas))
			for i, gp := range minGasPrices {
				fee := gp.Amount.Mul(glDec)
				requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
			}

			if !feeCoins.IsAnyGTE(requiredFees) {
				return ctx, sdkerrors.Wrapf(
					sdkerrors.ErrInsufficientFee,
					"insufficient fees; got: %s required: %s",
					feeCoins,
					requiredFees,
				)
			}
		}
	}

	return next(ctx, tx, simulate)
}

func isOracleTx(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		switch msg.(type) {
		case *oracletypes.MsgAggregateExchangeRatePrevote:
			continue

		case *oracletypes.MsgAggregateExchangeRateVote:
			continue

		default:
			return false
		}
	}

	return true
}
