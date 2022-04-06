package ante

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	oracletypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// SpamPreventionDecorator defines a custom Umee AnteHandler decorator that is
// responsible for preventing oracle message spam. Specifically, it prohibits
// oracle feeders from submitting multiple oracle messages in a single block.
type SpamPreventionDecorator struct {
	oracleKeeper     OracleKeeper
	oraclePrevoteMap map[string]int64
	oracleVoteMap    map[string]int64
	mu               sync.Mutex
}

func NewSpamPreventionDecorator(oracleKeeper OracleKeeper) *SpamPreventionDecorator {
	return &SpamPreventionDecorator{
		oracleKeeper:     oracleKeeper,
		oraclePrevoteMap: make(map[string]int64),
		oracleVoteMap:    make(map[string]int64),
		mu:               sync.Mutex{},
	}
}

func (spd *SpamPreventionDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	if ctx.IsCheckTx() && !simulate {
		if err := spd.CheckOracleSpam(ctx, tx.GetMsgs()); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// CheckOracleSpam performs the check of whether or not we've seen an oracle
// message from an oracle feeder in the current block or not. If we have, we
// return an error which prohibits the transaction from being processed.
func (spd *SpamPreventionDecorator) CheckOracleSpam(ctx sdk.Context, msgs []sdk.Msg) error {
	spd.mu.Lock()
	defer spd.mu.Unlock()

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
