package keeper

import (
	"context"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/InjectiveLabs/injective-core/metrics"

	"github.com/umee-network/umee/x/peggy/types"
)

var _ types.QueryServer = &Keeper{}

const maxValsetRequestsReturned = 5
const MaxResults = 100 // todo: impl pagination

// Params queries the params of the peggy module
func (k *Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	var params types.Params
	k.paramSpace.GetParamSet(sdk.UnwrapSDKContext(c), &params)
	return &types.QueryParamsResponse{Params: params}, nil

}

// CurrentValset queries the CurrentValset of the peggy module
func (k *Keeper) CurrentValset(c context.Context, req *types.QueryCurrentValsetRequest) (*types.QueryCurrentValsetResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	return &types.QueryCurrentValsetResponse{Valset: k.GetCurrentValset(sdk.UnwrapSDKContext(c))}, nil
}

// ValsetRequest queries the ValsetRequest of the peggy module
func (k *Keeper) ValsetRequest(c context.Context, req *types.QueryValsetRequestRequest) (*types.QueryValsetRequestResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	return &types.QueryValsetRequestResponse{Valset: k.GetValset(sdk.UnwrapSDKContext(c), req.Nonce)}, nil
}

// ValsetConfirm queries the ValsetConfirm of the peggy module
func (k *Keeper) ValsetConfirm(c context.Context, req *types.QueryValsetConfirmRequest) (*types.QueryValsetConfirmResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address invalid")
	}

	return &types.QueryValsetConfirmResponse{Confirm: k.GetValsetConfirm(sdk.UnwrapSDKContext(c), req.Nonce, addr)}, nil
}

// ValsetConfirmsByNonce queries the ValsetConfirmsByNonce of the peggy module
func (k *Keeper) ValsetConfirmsByNonce(c context.Context, req *types.QueryValsetConfirmsByNonceRequest) (*types.QueryValsetConfirmsByNonceResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	confirms := make([]*types.MsgValsetConfirm, 0)

	k.IterateValsetConfirmByNonce(sdk.UnwrapSDKContext(c), req.Nonce, func(_ []byte, valset *types.MsgValsetConfirm) (stop bool) {
		confirms = append(confirms, valset)

		return false
	})

	return &types.QueryValsetConfirmsByNonceResponse{Confirms: confirms}, nil
}

// LastValsetRequests queries the LastValsetRequests of the peggy module
func (k *Keeper) LastValsetRequests(c context.Context, req *types.QueryLastValsetRequestsRequest) (*types.QueryLastValsetRequestsResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	valReq := k.GetValsets(sdk.UnwrapSDKContext(c))
	valReqLen := len(valReq)
	retLen := 0

	if valReqLen < maxValsetRequestsReturned {
		retLen = valReqLen
	} else {
		retLen = maxValsetRequestsReturned
	}

	return &types.QueryLastValsetRequestsResponse{Valsets: valReq[0:retLen]}, nil
}

// LastPendingValsetRequestByAddr queries the LastPendingValsetRequestByAddr of the peggy module
func (k *Keeper) LastPendingValsetRequestByAddr(c context.Context, req *types.QueryLastPendingValsetRequestByAddrRequest) (*types.QueryLastPendingValsetRequestByAddrResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address invalid")
	}

	pendingValsetReq := make([]*types.Valset, 0)
	k.IterateValsets(sdk.UnwrapSDKContext(c), func(_ []byte, val *types.Valset) bool {
		// foundConfirm is true if the operatorAddr has signed the valset we are currently looking at
		foundConfirm := k.GetValsetConfirm(sdk.UnwrapSDKContext(c), val.Nonce, addr) != nil
		// if this valset has NOT been signed by operatorAddr, store it in pendingValsetReq
		// and exit the loop
		if !foundConfirm {
			pendingValsetReq = append(pendingValsetReq, val)
		}
		// if we have more than 100 unconfirmed requests in
		// our array we should exit, TODO pagination
		if len(pendingValsetReq) > 100 {
			return true
		}
		// return false to continue the loop
		return false
	})

	return &types.QueryLastPendingValsetRequestByAddrResponse{Valsets: pendingValsetReq}, nil
}

// BatchFees queries the batch fees from unbatched pool
func (k *Keeper) BatchFees(c context.Context, req *types.QueryBatchFeeRequest) (*types.QueryBatchFeeResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	return &types.QueryBatchFeeResponse{BatchFees: k.GetAllBatchFees(sdk.UnwrapSDKContext(c))}, nil
}

// LastPendingBatchRequestByAddr queries the LastPendingBatchRequestByAddr of the peggy module
func (k *Keeper) LastPendingBatchRequestByAddr(c context.Context, req *types.QueryLastPendingBatchRequestByAddrRequest) (*types.QueryLastPendingBatchRequestByAddrResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address invalid")
	}

	var pendingBatchReq *types.OutgoingTxBatch
	k.IterateOutgoingTXBatches(sdk.UnwrapSDKContext(c), func(_ []byte, batch *types.OutgoingTxBatch) (stop bool) {
		foundConfirm := k.GetBatchConfirm(sdk.UnwrapSDKContext(c), batch.BatchNonce, common.HexToAddress(batch.TokenContract), addr) != nil
		if !foundConfirm {
			pendingBatchReq = batch
			return true
		}

		return false
	})

	return &types.QueryLastPendingBatchRequestByAddrResponse{Batch: pendingBatchReq}, nil
}

// OutgoingTxBatches queries the OutgoingTxBatches of the peggy module
func (k *Keeper) OutgoingTxBatches(c context.Context, req *types.QueryOutgoingTxBatchesRequest) (*types.QueryOutgoingTxBatchesResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	batches := make([]*types.OutgoingTxBatch, 0)
	k.IterateOutgoingTXBatches(sdk.UnwrapSDKContext(c), func(_ []byte, batch *types.OutgoingTxBatch) bool {
		batches = append(batches, batch)
		return len(batches) == MaxResults
	})

	return &types.QueryOutgoingTxBatchesResponse{Batches: batches}, nil
}

// BatchRequestByNonce queries the BatchRequestByNonce of the peggy module
func (k *Keeper) BatchRequestByNonce(c context.Context, req *types.QueryBatchRequestByNonceRequest) (*types.QueryBatchRequestByNonceResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	if err := types.ValidateEthAddress(req.ContractAddress); err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, err.Error())
	}

	foundBatch := k.GetOutgoingTXBatch(sdk.UnwrapSDKContext(c), common.HexToAddress(req.ContractAddress), req.Nonce)
	if foundBatch == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Can not find tx batch")
	}

	return &types.QueryBatchRequestByNonceResponse{Batch: foundBatch}, nil
}

// BatchConfirms returns the batch confirmations by nonce and token contract
func (k *Keeper) BatchConfirms(c context.Context, req *types.QueryBatchConfirmsRequest) (*types.QueryBatchConfirmsResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	confirms := make([]*types.MsgConfirmBatch, 0)
	k.IterateBatchConfirmByNonceAndTokenContract(sdk.UnwrapSDKContext(c), req.Nonce, common.HexToAddress(req.ContractAddress),
		func(_ []byte, batch *types.MsgConfirmBatch) (stop bool) {
			confirms = append(confirms, batch)
			return false
		})

	return &types.QueryBatchConfirmsResponse{Confirms: confirms}, nil
}

// LastEventByAddr returns the last event for the given validator address, this allows eth oracles to figure out where they left off
func (k *Keeper) LastEventByAddr(c context.Context, req *types.QueryLastEventByAddrRequest) (*types.QueryLastEventByAddrResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(c)
	var ret types.QueryLastEventByAddrResponse

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, req.Address)
	}

	validator, found := k.GetOrchestratorValidator(ctx, addr)
	if !found {
		metrics.ReportFuncError(k.svcTags)
		return nil, sdkerrors.Wrap(types.ErrUnknown, "address")
	}

	lastClaimEvent := k.GetLastEventByValidator(ctx, validator)
	ret.LastClaimEvent = &lastClaimEvent

	return &ret, nil
}

// DenomToERC20 queries the Cosmos Denom that maps to an Ethereum ERC20
func (k *Keeper) DenomToERC20(c context.Context, req *types.QueryDenomToERC20Request) (*types.QueryDenomToERC20Response, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(c)
	cosmosOriginated, erc20, err := k.DenomToERC20Lookup(ctx, req.Denom)

	var ret types.QueryDenomToERC20Response
	ret.Erc20 = erc20.Hex()
	ret.CosmosOriginated = cosmosOriginated

	return &ret, err
}

// ERC20ToDenom queries the ERC20 contract that maps to an Ethereum ERC20 if any
func (k *Keeper) ERC20ToDenom(c context.Context, req *types.QueryERC20ToDenomRequest) (*types.QueryERC20ToDenomResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(c)
	cosmosOriginated, name := k.ERC20ToDenomLookup(ctx, common.HexToAddress(req.Erc20))

	var ret types.QueryERC20ToDenomResponse
	ret.Denom = name
	ret.CosmosOriginated = cosmosOriginated

	return &ret, nil
}

func (k *Keeper) GetDelegateKeyByValidator(c context.Context, req *types.QueryDelegateKeysByValidatorAddress) (*types.QueryDelegateKeysByValidatorAddressResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(c)

	valAddress, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, err
	}

	valAccountAddr := sdk.AccAddress(valAddress.Bytes())
	keys := k.GetOrchestratorAddresses(ctx)

	for _, key := range keys {
		senderAddr, err := sdk.AccAddressFromBech32(key.Sender)
		if err != nil {
			metrics.ReportFuncError(k.svcTags)
			return nil, err
		}
		if valAccountAddr.Equals(senderAddr) {
			return &types.QueryDelegateKeysByValidatorAddressResponse{EthAddress: key.EthAddress, OrchestratorAddress: key.Orchestrator}, nil
		}
	}

	metrics.ReportFuncError(k.svcTags)
	return nil, sdkerrors.Wrap(types.ErrInvalid, "No validator")
}

func (k *Keeper) GetDelegateKeyByOrchestrator(c context.Context, req *types.QueryDelegateKeysByOrchestratorAddress) (*types.QueryDelegateKeysByOrchestratorAddressResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(c)
	keys := k.GetOrchestratorAddresses(ctx)

	_, err := sdk.AccAddressFromBech32(req.OrchestratorAddress)
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, err
	}

	for _, key := range keys {
		if req.OrchestratorAddress == key.Orchestrator {
			return &types.QueryDelegateKeysByOrchestratorAddressResponse{ValidatorAddress: key.Sender, EthAddress: key.EthAddress}, nil
		}
	}

	metrics.ReportFuncError(k.svcTags)
	return nil, sdkerrors.Wrap(types.ErrInvalid, "No validator")
}

func (k *Keeper) GetDelegateKeyByEth(c context.Context, req *types.QueryDelegateKeysByEthAddress) (*types.QueryDelegateKeysByEthAddressResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(c)

	keys := k.GetOrchestratorAddresses(ctx)
	if err := types.ValidateEthAddress(req.EthAddress); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid eth address")
	}

	for _, key := range keys {
		if req.EthAddress == key.EthAddress {
			return &types.QueryDelegateKeysByEthAddressResponse{
				ValidatorAddress:    key.Sender,
				OrchestratorAddress: key.Orchestrator}, nil
		}
	}

	metrics.ReportFuncError(k.svcTags)
	return nil, sdkerrors.Wrap(types.ErrInvalid, "No validator")
}

func (k *Keeper) GetPendingSendToEth(c context.Context, req *types.QueryPendingSendToEth) (*types.QueryPendingSendToEthResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(c)
	batches := k.GetOutgoingTxBatches(ctx)
	unbatchedTx := k.GetPoolTransactions(ctx)
	senderAddress := req.SenderAddress

	res := &types.QueryPendingSendToEthResponse{}
	res.TransfersInBatches = make([]*types.OutgoingTransferTx, 0)
	res.UnbatchedTransfers = make([]*types.OutgoingTransferTx, 0)

	for _, batch := range batches {
		for _, tx := range batch.Transactions {
			if tx.Sender == senderAddress {
				res.TransfersInBatches = append(res.TransfersInBatches, tx)
			}
		}
	}

	for _, tx := range unbatchedTx {
		if strings.EqualFold(tx.Sender, senderAddress) {
			res.UnbatchedTransfers = append(res.UnbatchedTransfers, tx)

		}
	}

	return res, nil
}

func (k *Keeper) PeggyModuleState(c context.Context, req *types.QueryModuleStateRequest) (*types.QueryModuleStateResponse, error) {
	metrics.ReportFuncCall(k.grpcTags)
	doneFn := metrics.ReportFuncTiming(k.grpcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(c)

	var (
		p                               = k.GetParams(ctx)
		batches                         = k.GetOutgoingTxBatches(ctx)
		valsets                         = k.GetValsets(ctx)
		attmap                          = k.GetAttestationMapping(ctx)
		vsconfs                         = []*types.MsgValsetConfirm{}
		batchconfs                      = []*types.MsgConfirmBatch{}
		attestations                    = []*types.Attestation{}
		orchestratorAddresses           = k.GetOrchestratorAddresses(ctx)
		lastObservedEventNonce          = k.GetLastObservedEventNonce(ctx)
		lastObservedEthereumBlockHeight = k.GetLastObservedEthereumBlockHeight(ctx)
		erc20ToDenoms                   = []*types.ERC20ToDenom{}
		unbatched_transfers             = k.GetPoolTransactions(ctx)
	)

	// export valset confirmations from state
	for _, vs := range valsets {
		vsconfs = append(vsconfs, k.GetValsetConfirms(ctx, vs.Nonce)...)
	}

	// export batch confirmations from state
	for _, batch := range batches {
		batchconfs = append(batchconfs, k.GetBatchConfirmByNonceAndTokenContract(ctx, batch.BatchNonce, common.HexToAddress(batch.TokenContract))...)
	}

	// sort attestation map keys since map iteration is non-deterministic
	attestationHeights := make([]uint64, 0, len(attmap))
	for k := range attmap {
		attestationHeights = append(attestationHeights, k)
	}
	sort.SliceStable(attestationHeights, func(i, j int) bool {
		return attestationHeights[i] < attestationHeights[j]
	})

	for _, height := range attestationHeights {
		attestations = append(attestations, attmap[height]...)
	}

	// export erc20 to denom relations
	k.IterateERC20ToDenom(ctx, func(_ []byte, erc20ToDenom *types.ERC20ToDenom) bool {
		erc20ToDenoms = append(erc20ToDenoms, erc20ToDenom)
		return false
	})

	lastOutgoingBatchID := k.GetLastOutgoingBatchID(ctx)
	lastOutgoingPoolID := k.GetLastOutgoingPoolID(ctx)
	lastObservedValset := k.GetLastObservedValset(ctx)

	state := types.GenesisState{
		Params:                     p,
		LastObservedNonce:          lastObservedEventNonce,
		LastObservedEthereumHeight: lastObservedEthereumBlockHeight.EthereumBlockHeight,
		Valsets:                    valsets,
		ValsetConfirms:             vsconfs,
		Batches:                    batches,
		BatchConfirms:              batchconfs,
		Attestations:               attestations,
		OrchestratorAddresses:      orchestratorAddresses,
		Erc20ToDenoms:              erc20ToDenoms,
		UnbatchedTransfers:         unbatched_transfers,
		LastOutgoingBatchId:        lastOutgoingBatchID,
		LastOutgoingPoolId:         lastOutgoingPoolID,
		LastObservedValset:         *lastObservedValset,
	}

	res := &types.QueryModuleStateResponse{
		State: &state,
	}

	return res, nil
}
