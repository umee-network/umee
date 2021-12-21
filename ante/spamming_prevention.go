package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	oracletypes "github.com/umee-network/umee/x/oracle/types"
)

// SpammingPreventionDecorator will check if the transaction's gas is smaller than
// configured hard cap.
type SpamPreventionDecorator struct {
	oracleKeeper     OracleKeeper
	oraclePrevoteMap map[string]int64
	oracleVoteMap    map[string]int64
}

// NewSpammingPreventionDecorator returns new spamming prevention decorator instance
func NewSpammingPreventionDecorator(oracleKeeper OracleKeeper) SpamPreventionDecorator {
	return SpamPreventionDecorator{
		oracleKeeper:     oracleKeeper,
		oraclePrevoteMap: make(map[string]int64),
		oracleVoteMap:    make(map[string]int64),
	}
}

// AnteHandle handles msg tax fee checking
func (spd SpamPreventionDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	if ctx.IsCheckTx() && !simulate {
		if err := spd.CheckOracleSpamming(ctx, tx.GetMsgs()); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// CheckOracleSpamming check whether the msgs are spamming purpose or not
func (spd *SpamPreventionDecorator) CheckOracleSpamming(ctx sdk.Context, msgs []sdk.Msg) error {
	curHeight := ctx.BlockHeight()
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *oracletypes.MsgAggregateExchangeRatePrevote:
			feederAddr, err := sdk.AccAddressFromBech32(msg.Feeder)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
			if err != nil {
				return err
			}

			err = spd.oracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
			if err != nil {
				return err
			}

			if lastSubmittedHeight, ok := spd.oraclePrevoteMap[msg.Validator]; ok && lastSubmittedHeight == curHeight {
				return sdkerrors.Wrap(
					sdkerrors.ErrInvalidRequest,
					"validator has already submitted a pre-vote message at the current height",
				)
			}

			spd.oraclePrevoteMap[msg.Validator] = curHeight
			continue
		case *oracletypes.MsgAggregateExchangeRateVote:
			feederAddr, err := sdk.AccAddressFromBech32(msg.Feeder)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
			if err != nil {
				return err
			}

			err = spd.oracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
			if err != nil {
				return err
			}

			if lastSubmittedHeight, ok := spd.oracleVoteMap[msg.Validator]; ok && lastSubmittedHeight == curHeight {
				return sdkerrors.Wrap(
					sdkerrors.ErrInvalidRequest,
					"validator has already submitted a vote message at the current height",
				)
			}

			spd.oracleVoteMap[msg.Validator] = curHeight
			continue
		default:
			return nil
		}
	}

	return nil
}
