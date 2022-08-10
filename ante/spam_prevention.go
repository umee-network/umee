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
		var err error
		switch msg := msg.(type) {
		case *oracletypes.MsgAggregateExchangeRatePrevote:
			err = spd.validate(ctx, msg.Feeder, msg.Validator, spd.oraclePrevoteMap, curHeight, "pre-vote")
		case *oracletypes.MsgAggregateExchangeRateVote:
			err = spd.validate(ctx, msg.Feeder, msg.Validator, spd.oracleVoteMap, curHeight, "vote")
		default:
			// non oracle msg: stop validation!
			return nil
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (spd *SpamPreventionDecorator) validate(
	ctx sdk.Context,
	feeder,
	validator string,
	cache map[string]int64,
	curHeight int64,
	txType string,
) error {
	feederAddr, err := sdk.AccAddressFromBech32(feeder)
	if err != nil {
		return err
	}
	valAddr, err := sdk.ValAddressFromBech32(validator)
	if err != nil {
		return err
	}
	if err = spd.oracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr); err != nil {
		return err
	}
	if lastSubmitted, ok := cache[validator]; ok && lastSubmitted == curHeight {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"validator has already submitted a %s message at the current height", txType)
	}
	cache[validator] = curHeight
	return nil
}
