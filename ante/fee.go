package ante

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	oracletypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// MaxOracleMsgGasUsage defines the maximum gas allowed for an oracle transaction.
const MaxOracleMsgGasUsage = uint64(100000)

// feeAndPriority ensures tx has enough fee coins to pay for the gas at the CheckTx time
// to early remove transactions from the mempool without enough attached fee.
// The validator min fee check is ignored if the tx contains only oracle messages and
// tx gas limit is <= MaxOracleMsgGasUsage. Essentially, validators can provide price
// transactison for free as long as the gas per message is in the MaxOracleMsgGasUsage limit.
func feeAndPriority(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.ErrTxDecode.Wrap("Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()
	isOracle := isOracleTx(msgs)
	chargeForOracle := !isOracle || gas > uint64(len(msgs))*MaxOracleMsgGasUsage

	if ctx.IsCheckTx() && chargeForOracle {
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
				return nil, 0, sdkerrors.ErrInsufficientFee.Wrapf(
					"insufficient fees; got: %s required: %s", feeCoins, requiredFees)
			}
		}
	}

	priority := getTxPriority(feeCoins, isOracle)
	return feeCoins, priority, nil
}

// isOracleTx checks if all messages are oracle messages
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

// getTxPriority returns naive tx priority based on the lowest fee amount (regardless of the
// denom) and oracle tx check.
func getTxPriority(fee sdk.Coins, isOracle bool) int64 {
	var priority int64
	for _, c := range fee {
		// TODO: we should better compare amounts
		// https://github.com/umee-network/umee/issues/510
		p := int64(math.MaxInt64)
		if c.Amount.IsInt64() {
			p = c.Amount.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}
	if isOracle {
		// TODO: this is a naive version.
		// Proper solution will be implemented in https://github.com/umee-network/umee/issues/510
		priority += 100000
	}
	return priority
}
