package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/umee-network/umee/x/peggy/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{
		Keeper: keeper,
	}
}

func (k msgServer) SetOrchestratorAddresses(c context.Context, msg *types.MsgSetOrchestratorAddresses) (*types.MsgSetOrchestratorAddressesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	valAccAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	// get orchestrator address if available, otherwise default to validator address
	var orchAddr sdk.AccAddress
	if msg.Orchestrator != "" {
		orchAddr, err = sdk.AccAddressFromBech32(msg.Orchestrator)
		if err != nil {
			return nil, err
		}
	} else {
		orchAddr = valAccAddr
	}

	valAddr := sdk.ValAddress(valAccAddr)
	ethAddr := common.HexToAddress(msg.EthAddress)

	// ensure that the validator exists
	if k.Keeper.StakingKeeper.Validator(ctx, valAddr) == nil {
		return nil, sdkerrors.Wrap(stakingtypes.ErrNoValidatorFound, valAddr.String())
	}

	if _, ok := k.GetOrchestratorValidator(ctx, orchAddr); ok {
		return nil, sdkerrors.Wrapf(types.ErrSetOrchAddresses, "orchestrator address %s in use", orchAddr)
	}

	if _, ok := k.GetValidatorByEthAddress(ctx, ethAddr); ok {
		return nil, sdkerrors.Wrapf(types.ErrSetOrchAddresses, "ethereum address %s in use", ethAddr)
	}

	valAccSeq, err := k.accountKeeper.GetSequence(ctx, valAccAddr)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrSetOrchAddresses, "failed to get sequence for validator account %s", valAccAddr)
	}

	var nonce uint64
	if valAccSeq > 0 {
		// We decrement since we process the message after the ante-handler which
		// increments the nonce.
		nonce = valAccSeq - 1
	}

	signMsgBz := k.cdc.MustMarshal(&types.SetOrchestratorAddressesSignMsg{
		ValidatorAddress: valAddr.String(),
		Nonce:            nonce,
	})

	hash := ethcrypto.Keccak256Hash(signMsgBz)

	if err = types.ValidateEthereumSignature(hash, msg.EthSignature, ethAddr); err != nil {
		return nil, sdkerrors.Wrapf(
			types.ErrSetOrchAddresses,
			"failed to validate delegate keys signature for Ethereum address %X; %s; %d",
			ethAddr, err, nonce,
		)
	}

	k.SetOrchestratorValidator(ctx, valAddr, orchAddr)
	k.SetEthAddressForValidator(ctx, valAddr, ethAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			sdk.NewAttribute(types.AttributeKeySetOrchestratorAddr, orchAddr.String()),
			sdk.NewAttribute(types.AttributeKeySetEthereumAddr, ethAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyValidatorAddr, valAddr.String()),
		),
	)

	return &types.MsgSetOrchestratorAddressesResponse{}, nil
}

// ValsetConfirm handles MsgValsetConfirm
func (k msgServer) ValsetConfirm(c context.Context, msg *types.MsgValsetConfirm) (*types.MsgValsetConfirmResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	valset := k.GetValset(ctx, msg.Nonce)
	if valset == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find valset")
	}

	peggyID := k.GetPeggyID(ctx)
	checkpoint := valset.GetCheckpoint(peggyID)

	sigBytes, err := hex.DecodeString(msg.Signature)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "signature decoding")
	}
	orchaddr, _ := sdk.AccAddressFromBech32(msg.Orchestrator)
	validator, found := k.GetOrchestratorValidator(ctx, orchaddr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	ethAddress, found := k.GetEthAddressByValidator(ctx, validator)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrEmpty, "no eth address found")
	}

	if err = types.ValidateEthereumSignature(checkpoint, sigBytes, ethAddress); err != nil {
		description := fmt.Sprintf(
			"signature verification failed expected sig by %s with peggy-id %s with checkpoint %s found %s",
			ethAddress, peggyID, checkpoint.Hex(), msg.Signature,
		)

		return nil, sdkerrors.Wrap(types.ErrInvalid, description)
	}

	// persist signature
	if k.GetValsetConfirm(ctx, msg.Nonce, orchaddr) != nil {
		return nil, sdkerrors.Wrap(types.ErrDuplicate, "signature duplicate")
	}
	key := k.SetValsetConfirm(ctx, msg)

	ctx.EventManager().EmitTypedEvent(&types.EventValsetConfirm{
		ValsetConfirmKey: key,
	})

	return &types.MsgValsetConfirmResponse{}, nil
}

// SendToEth handles MsgSendToEth
func (k msgServer) SendToEth(c context.Context, msg *types.MsgSendToEth) (*types.MsgSendToEthResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	txID, err := k.AddToOutgoingPool(ctx, sender, common.HexToAddress(msg.EthDest), msg.Amount, msg.BridgeFee)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitTypedEvent(&types.EventSendToEth{
		OutgoingTxId: txID,
	})

	return &types.MsgSendToEthResponse{}, nil
}

// RequestBatch handles MsgRequestBatch
func (k msgServer) RequestBatch(c context.Context, msg *types.MsgRequestBatch) (*types.MsgRequestBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Check if the denom is a peggy coin, if not, check if there is a deployed ERC20 representing it.
	// If not, error out
	_, tokenContract, err := k.DenomToERC20Lookup(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}

	batch, err := k.BuildOutgoingTXBatch(ctx, tokenContract, OutgoingTxBatchSize)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitTypedEvent(&types.EventRequestBatch{
		BatchNonce: batch.BatchNonce,
	})

	return &types.MsgRequestBatchResponse{}, nil
}

// ConfirmBatch handles MsgConfirmBatch
func (k msgServer) ConfirmBatch(c context.Context, msg *types.MsgConfirmBatch) (*types.MsgConfirmBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	tokenContract := common.HexToAddress(msg.TokenContract)

	// fetch the outgoing batch given the nonce
	batch := k.GetOutgoingTXBatch(ctx, tokenContract, msg.Nonce)
	if batch == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find batch")
	}

	peggyID := k.GetPeggyID(ctx)
	checkpoint := batch.GetCheckpoint(peggyID)

	sigBytes, err := hex.DecodeString(msg.Signature)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "signature decoding")
	}

	orchaddr, _ := sdk.AccAddressFromBech32(msg.Orchestrator)
	validator, found := k.GetOrchestratorValidator(ctx, orchaddr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	ethAddress, found := k.GetEthAddressByValidator(ctx, validator)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrEmpty, "eth address not found")
	}

	err = types.ValidateEthereumSignature(checkpoint, sigBytes, ethAddress)
	if err != nil {
		description := fmt.Sprintf(
			"signature verification failed expected sig by %s with peggy-id %s with checkpoint %s found %s",
			ethAddress, peggyID, checkpoint.Hex(), msg.Signature,
		)

		return nil, sdkerrors.Wrap(types.ErrInvalid, description)
	}

	// check if we already have this confirm
	if k.GetBatchConfirm(ctx, msg.Nonce, tokenContract, orchaddr) != nil {
		return nil, sdkerrors.Wrap(types.ErrDuplicate, "duplicate signature")
	}
	key := k.SetBatchConfirm(ctx, msg)

	ctx.EventManager().EmitTypedEvent(&types.EventConfirmBatch{
		BatchConfirmKey: key,
	})

	return nil, nil
}

// DepositClaim handles MsgDepositClaim
// TODO it is possible to submit an old msgDepositClaim (old defined as covering an event nonce that has already been
// executed aka 'observed' and had it's slashing window expire) that will never be cleaned up in the endblocker. This
// should not be a security risk as 'old' events can never execute but it does store spam in the chain.
func (k msgServer) DepositClaim(c context.Context, msg *types.MsgDepositClaim) (*types.MsgDepositClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	orchestrator, _ := sdk.AccAddressFromBech32(msg.Orchestrator)
	validator, found := k.GetOrchestratorValidator(ctx, orchestrator)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	// return an error if the validator isn't in the active set
	val := k.StakingKeeper.Validator(ctx, validator)
	if val == nil || !val.IsBonded() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, "validator not in active set")
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	// Add the claim to the store
	_, err = k.Attest(ctx, msg, any)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create attestation")
	}

	ctx.EventManager().EmitTypedEvent(&types.EventDepositClaim{
		AttestationId: types.GetAttestationKey(msg.EventNonce, msg.ClaimHash()),
	})

	return &types.MsgDepositClaimResponse{}, nil
}

// WithdrawClaim handles MsgWithdrawClaim
// TODO it is possible to submit an old msgWithdrawClaim (old defined as covering an event nonce that has already been
// executed aka 'observed' and had it's slashing window expire) that will never be cleaned up in the endblocker. This
// should not be a security risk as 'old' events can never execute but it does store spam in the chain.
func (k msgServer) WithdrawClaim(c context.Context, msg *types.MsgWithdrawClaim) (*types.MsgWithdrawClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	orchestrator, _ := sdk.AccAddressFromBech32(msg.Orchestrator)
	validator, found := k.GetOrchestratorValidator(ctx, orchestrator)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	// return an error if the validator isn't in the active set
	val := k.StakingKeeper.Validator(ctx, validator)
	if val == nil || !val.IsBonded() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, "validator not in acitve set")
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	// Add the claim to the store
	_, err = k.Attest(ctx, msg, any)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create attestation")
	}

	ctx.EventManager().EmitTypedEvent(&types.EventWithdrawClaim{
		AttestationId: types.GetAttestationKey(msg.EventNonce, msg.ClaimHash()),
	})

	return &types.MsgWithdrawClaimResponse{}, nil
}

// ERC20DeployedClaim handles MsgERC20Deployed
func (k msgServer) ERC20DeployedClaim(c context.Context, msg *types.MsgERC20DeployedClaim) (*types.MsgERC20DeployedClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	orch, _ := sdk.AccAddressFromBech32(msg.Orchestrator)
	validator, found := k.GetOrchestratorValidator(ctx, orch)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	// return an error if the validator isn't in the active set
	val := k.StakingKeeper.Validator(ctx, validator)
	if val == nil || !val.IsBonded() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, "validator not in acitve set")
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	// Add the claim to the store
	_, err = k.Attest(ctx, msg, any)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create attestation")
	}

	ctx.EventManager().EmitTypedEvent(&types.EventERC20DeployedClaim{
		AttestationId: types.GetAttestationKey(msg.EventNonce, msg.ClaimHash()),
	})

	return &types.MsgERC20DeployedClaimResponse{}, nil
}

// ValsetUpdateClaim handles claims for executing a validator set update on Ethereum
func (k msgServer) ValsetUpdateClaim(c context.Context, msg *types.MsgValsetUpdatedClaim) (*types.MsgValsetUpdatedClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	orchaddr, _ := sdk.AccAddressFromBech32(msg.Orchestrator)
	validator, found := k.GetOrchestratorValidator(ctx, orchaddr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	// return an error if the validator isn't in the active set
	val := k.StakingKeeper.Validator(ctx, validator)
	if val == nil || !val.IsBonded() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, "validator not in acitve set")
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	// Add the claim to the store
	_, err = k.Attest(ctx, msg, any)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create attestation")
	}

	ctx.EventManager().EmitTypedEvent(&types.EventValsetUpdateClaim{
		AttestationId: types.GetAttestationKey(msg.EventNonce, msg.ClaimHash()),
	})

	return &types.MsgValsetUpdatedClaimResponse{}, nil
}

func (k msgServer) CancelSendToEth(c context.Context, msg *types.MsgCancelSendToEth) (*types.MsgCancelSendToEthResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = k.RemoveFromOutgoingPoolAndRefund(ctx, msg.TransactionId, sender)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitTypedEvent(&types.EventCancelSendToEth{
		OutgoingTxId: msg.TransactionId,
	})

	return &types.MsgCancelSendToEthResponse{}, nil
}

func (k msgServer) SubmitBadSignatureEvidence(c context.Context, msg *types.MsgSubmitBadSignatureEvidence) (*types.MsgSubmitBadSignatureEvidenceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	err := k.CheckBadSignatureEvidence(ctx, msg)
	ctx.EventManager().EmitTypedEvent(&types.EventSubmitBadSignatureEvidence{
		BadEthSignature:        msg.Signature,
		BadEthSignatureSubject: fmt.Sprint(msg.Subject),
	})

	return &types.MsgSubmitBadSignatureEvidenceResponse{}, err
}
