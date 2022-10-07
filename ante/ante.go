package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	ibcante "github.com/cosmos/ibc-go/v5/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v5/modules/core/keeper"
)

type HandlerOptions struct {
	AccountKeeper   cosmosante.AccountKeeper
	BankKeeper      types.BankKeeper
	FeegrantKeeper  cosmosante.FeegrantKeeper
	OracleKeeper    OracleKeeper
	IBCKeeper       *ibckeeper.Keeper
	SignModeHandler signing.SignModeHandler
	SigGasConsumer  cosmosante.SignatureVerificationGasConsumer
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("account keeper is required for ante builder")
	}
	if options.BankKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("bank keeper is required for ante builder")
	}
	if options.OracleKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("oracle keeper is required for ante builder")
	}
	if options.SignModeHandler == nil {
		return nil, sdkerrors.ErrLogic.Wrap("sign mode handler is required for ante builder")
	}

	return sdk.ChainAnteDecorators(
		cosmosante.NewSetUpContextDecorator(),            // outermost AnteDecorator. SetUpContext must be called first
		cosmosante.NewExtensionOptionsDecorator(nil),     // nil=reject extensions
		NewSpamPreventionDecorator(options.OracleKeeper), // spam prevention
		cosmosante.NewValidateBasicDecorator(),
		cosmosante.NewTxTimeoutHeightDecorator(),
		cosmosante.NewValidateMemoDecorator(options.AccountKeeper),
		cosmosante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		cosmosante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, FeeAndPriority),
		// SetPubKeyDecorator must be called before all signature verification decorators
		cosmosante.NewSetPubKeyDecorator(options.AccountKeeper),
		cosmosante.NewValidateSigCountDecorator(options.AccountKeeper),
		cosmosante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		cosmosante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		cosmosante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	), nil
}
