package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/umee-network/umee/x/peggy/types"
)

// NormalizeGenesis takes care of formatting in the internal structures, as they're used as values
// in the keeper eventually, while having raw strings in them.
func NormalizeGenesis(data *types.GenesisState) {
	data.Params.BridgeEthereumAddress = common.HexToAddress(data.Params.BridgeEthereumAddress).Hex()

	for _, valset := range data.Valsets {
		for _, member := range valset.Members {
			member.EthereumAddress = common.HexToAddress(member.EthereumAddress).Hex()
		}
	}

	for _, valsetConfirm := range data.ValsetConfirms {
		valsetConfirm.EthAddress = common.HexToAddress(valsetConfirm.EthAddress).Hex()
	}

	for _, batch := range data.Batches {
		batch.TokenContract = common.HexToAddress(batch.TokenContract).Hex()

		for _, outgoingTx := range batch.Transactions {
			outgoingTx.DestAddress = common.HexToAddress(outgoingTx.DestAddress).Hex()
			outgoingTx.Erc20Fee.Contract = common.HexToAddress(outgoingTx.Erc20Fee.Contract).Hex()
			outgoingTx.Erc20Token.Contract = common.HexToAddress(outgoingTx.Erc20Token.Contract).Hex()
		}
	}

	for _, batchConfirm := range data.BatchConfirms {
		batchConfirm.EthSigner = common.HexToAddress(batchConfirm.EthSigner).Hex()
		batchConfirm.TokenContract = common.HexToAddress(batchConfirm.TokenContract).Hex()
	}

	for _, orchestrator := range data.OrchestratorAddresses {
		orchestrator.EthAddress = common.HexToAddress(orchestrator.EthAddress).Hex()
	}

	for _, token := range data.Erc20ToDenoms {
		token.Erc20 = common.HexToAddress(token.Erc20).Hex()
	}
}

// InitGenesis starts a chain from a genesis state
func InitGenesis(ctx sdk.Context, k Keeper, data *types.GenesisState) {
	NormalizeGenesis(data)

	k.SetParams(ctx, data.Params)

	for _, valset := range data.Valsets {
		k.StoreValsetUnsafe(ctx, valset)
	}

	for _, valsetConfirm := range data.ValsetConfirms {
		k.SetValsetConfirm(ctx, valsetConfirm)
	}

	for _, batch := range data.Batches {
		k.StoreBatchUnsafe(ctx, batch)
	}

	for _, batchConfirm := range data.BatchConfirms {
		k.SetBatchConfirm(ctx, batchConfirm)
	}

	// reset pool transactions in state
	for _, tx := range data.UnbatchedTransfers {
		if err := k.setPoolEntry(ctx, tx); err != nil {
			panic(err)
		}
	}

	// reset attestations in state
	for _, attestation := range data.Attestations {
		claim, err := k.UnpackAttestationClaim(attestation)
		if err != nil {
			panic("couldn't UnpackAttestationClaim")
		}

		k.SetAttestation(ctx, claim.GetEventNonce(), claim.ClaimHash(), attestation)
	}

	k.setLastObservedEventNonce(ctx, data.LastObservedNonce)
	k.SetLastObservedEthereumBlockHeight(ctx, data.LastObservedEthereumHeight)
	k.SetLastOutgoingBatchID(ctx, data.LastOutgoingBatchId)
	k.SetLastOutgoingPoolID(ctx, data.LastOutgoingPoolId)
	k.SetLastObservedValset(ctx, data.LastObservedValset)

	for _, attestation := range data.Attestations {
		claim, err := k.UnpackAttestationClaim(attestation)
		if err != nil {
			panic("couldn't UnpackAttestationClaim")
		}

		// reconstruct the latest event nonce for every validator
		// if somehow this genesis state is saved when all attestations
		// have been cleaned up GetLastEventNonceByValidator handles that case
		//
		// if we where to save and load the last event nonce for every validator
		// then we would need to carry that state forever across all chain restarts
		// but since we've already had to handle the edge case of new validators joining
		// while all attestations have already been cleaned up we can do this instead and
		// not carry around every validators event nonce counter forever.
		for _, vote := range attestation.Votes {
			val, err := sdk.ValAddressFromBech32(vote)
			if err != nil {
				panic(err)
			}

			lastEvent := k.GetLastEventByValidator(ctx, val)
			if claim.GetEventNonce() > lastEvent.EthereumEventNonce {
				k.setLastEventByValidator(ctx, val, claim.GetEventNonce(), claim.GetBlockHeight())
			}
		}
	}

	// reset delegate keys in state
	for _, keys := range data.OrchestratorAddresses {
		if err := keys.ValidateBasic(); err != nil {
			panic("invalid delegate key in genesis")
		}

		validatorAccountAddress, _ := sdk.AccAddressFromBech32(keys.Sender)
		valAddress := sdk.ValAddress(validatorAccountAddress.Bytes())
		orchestrator, _ := sdk.AccAddressFromBech32(keys.Orchestrator)

		// set the orchestrator Cosmos address
		k.SetOrchestratorValidator(ctx, valAddress, orchestrator)

		// set the orchestrator Ethereum address
		k.SetEthAddressForValidator(ctx, valAddress, common.HexToAddress(keys.EthAddress))
	}

	// populate state with cosmos originated denom-erc20 mapping
	for _, item := range data.Erc20ToDenoms {
		k.SetCosmosOriginatedDenomToERC20(ctx, item.Denom, common.HexToAddress(item.Erc20))
	}
}

// ExportGenesis exports all the state needed to restart the chain
// from the current state of the chain
func ExportGenesis(ctx sdk.Context, k Keeper) types.GenesisState {
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
		unbatchedTransfers              = k.GetPoolTransactions(ctx)
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

	return types.GenesisState{
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
		UnbatchedTransfers:         unbatchedTransfers,
		LastOutgoingBatchId:        lastOutgoingBatchID,
		LastOutgoingPoolId:         lastOutgoingPoolID,
		LastObservedValset:         *lastObservedValset,
	}
}
