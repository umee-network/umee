package ante

import (
	"math"

	gbtypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	oracletypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// MaxMsgGasUsage defines the maximum gas allowed for an oracle transaction.
const MaxMsgGasUsage = uint64(100000)

// FeeAndPriority ensures tx has enough fee coins to pay for the gas at the CheckTx time
// to early remove transactions from the mempool without enough attached fee.
// The validator min fee check is ignored if the tx contains only oracle messages and
// tx gas limit is <= MaxOracleMsgGasUsage. Essentially, validators can provide price
// transactison for free as long as the gas per message is in the MaxOracleMsgGasUsage limit.
func FeeAndPriority(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.ErrTxDecode.Wrap("Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()
	isOracleOrGravity := IsOracleOrGravityTx(msgs)
	chargeFees := !isOracleOrGravity || gas > uint64(len(msgs))*MaxMsgGasUsage

	if ctx.IsCheckTx() && chargeFees {
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

	priority := getTxPriority(feeCoins, isOracleOrGravity)
	if !chargeFees {
		return sdk.Coins{}, priority, nil
	}
	return feeCoins, priority, nil
}

// IsOracleOrGravityTx checks if all messages are oracle messages
func IsOracleOrGravityTx(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		switch msg.(type) {
		case *oracletypes.MsgAggregateExchangeRatePrevote,
			*oracletypes.MsgAggregateExchangeRateVote:
			continue

		// TODO: remove messages which should not be "free":
		case *gbtypes.MsgValsetConfirm,
			*gbtypes.MsgRequestBatch,
			*gbtypes.MsgConfirmBatch,
			*gbtypes.MsgERC20DeployedClaim,
			*gbtypes.MsgConfirmLogicCall,
			*gbtypes.MsgLogicCallExecutedClaim,
			*gbtypes.MsgSendToCosmosClaim,
			*gbtypes.MsgExecuteIbcAutoForwards,
			*gbtypes.MsgBatchSendToEthClaim,
			*gbtypes.MsgValsetUpdatedClaim,
			*gbtypes.MsgSubmitBadSignatureEvidence:
			continue

		default:
			return false
		}
	}

	return true
}

// getTxPriority returns naive tx priority based on the lowest fee amount (regardless of the
// denom) and oracle tx check.
func getTxPriority(fee sdk.Coins, isOracleOrGravity bool) int64 {
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
	if isOracleOrGravity {
		// TODO: this is a naive version.
		// Proper solution will be implemented in https://github.com/umee-network/umee/issues/510
		priority += 100000
	}
	return priority
}
