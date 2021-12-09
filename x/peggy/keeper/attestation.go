package keeper

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/peggy/types"
)

func (k *Keeper) Attest(ctx sdk.Context, claim types.EthereumClaim, anyClaim *codectypes.Any) (*types.Attestation, error) {
	valAddr, found := k.GetOrchestratorValidator(ctx, claim.GetClaimer())
	if !found {
		panic("Could not find ValAddr for delegate key, should be checked by now")
	}

	// Check that the nonce of this event is exactly one higher than the last nonce stored by this validator.
	// We check the event nonce in processAttestation as well,
	// but checking it here gives individual eth signers a chance to retry,
	// and prevents validators from submitting two claims with the same nonce
	lastEvent := k.GetLastEventByValidator(ctx, valAddr)
	if claim.GetEventNonce() != lastEvent.EthereumEventNonce+1 {
		return nil, types.ErrNonContiguousEventNonce
	}

	// Tries to get an attestation with the same eventNonce and claim as the claim that was submitted.
	att := k.GetAttestation(ctx, claim.GetEventNonce(), claim.ClaimHash())

	// If it does not exist, create a new one.
	if att == nil {
		att = &types.Attestation{
			Observed: false,
			Height:   uint64(ctx.BlockHeight()),
			Claim:    anyClaim,
		}
	}

	// Add the validator's vote to this attestation
	att.Votes = append(att.Votes, valAddr.String())

	k.SetAttestation(ctx, claim.GetEventNonce(), claim.ClaimHash(), att)
	k.setLastEventByValidator(ctx, valAddr, claim.GetEventNonce(), claim.GetBlockHeight())

	return att, nil
}

// TryAttestation checks if an attestation has enough votes to be applied to the consensus state
// and has not already been marked Observed, then calls processAttestation to actually apply it to the state,
// and then marks it Observed and emits an event.
func (k *Keeper) TryAttestation(ctx sdk.Context, att *types.Attestation) {
	claim, err := k.UnpackAttestationClaim(att)
	if err != nil {
		panic("could not cast to claim")
	}

	// If the attestation has not yet been Observed, sum up the votes and see if it is ready to apply to the state.
	// This conditional stops the attestation from accidentally being applied twice.
	if !att.Observed {
		// Sum the current powers of all validators who have voted and see if it passes the current threshold
		totalPower := k.StakingKeeper.GetLastTotalPower(ctx)
		requiredPower := types.AttestationVotesPowerThreshold.Mul(totalPower).Quo(sdk.NewInt(100))
		attestationPower := sdk.ZeroInt()
		for _, validator := range att.Votes {
			val, err := sdk.ValAddressFromBech32(validator)
			if err != nil {
				panic(err)
			}

			validatorPower := k.StakingKeeper.GetLastValidatorPower(ctx, val)
			// Add it to the attestation power's sum
			attestationPower = attestationPower.Add(sdk.NewInt(validatorPower))
			// If the power of all the validators that have voted on the attestation is higher or equal to the threshold,
			// process the attestation, set Observed to true, and break
			if attestationPower.GTE(requiredPower) {
				lastEventNonce := k.GetLastObservedEventNonce(ctx)
				// this check is performed at the next level up so this should never panic
				// outside of programmer error.
				if claim.GetEventNonce() != lastEventNonce+1 {
					panic("attempting to apply events to state out of order")
				}

				k.setLastObservedEventNonce(ctx, claim.GetEventNonce())
				k.SetLastObservedEthereumBlockHeight(ctx, claim.GetBlockHeight())

				att.Observed = true
				k.SetAttestation(ctx, claim.GetEventNonce(), claim.ClaimHash(), att)

				k.processAttestation(ctx, claim)
				k.emitObservedEvent(ctx, att, claim)
				break
			}
		}
	} else {
		// We panic here because this should never happen
		panic("attempting to process observed attestation")
	}
}

// processAttestation actually applies the attestation to the consensus state
func (k *Keeper) processAttestation(ctx sdk.Context, claim types.EthereumClaim) {
	// then execute in a new Tx so that we can store state on failure
	xCtx, commit := ctx.CacheContext()
	if err := k.attestationHandler.Handle(xCtx, claim); err != nil { // execute with a transient storage
		// If the attestation fails, something has gone wrong and we can't recover it. Log and move on
		// The attestation will still be marked "Observed", and validators can still be slashed for not
		// having voted for it.
		k.Logger(ctx).Error(
			"attestation failed",
			"err", err,
			"claim_type", claim.GetType().String(),
			"id", fmt.Sprintf("%X", types.GetAttestationKey(claim.GetEventNonce(), claim.ClaimHash())),
			"nonce", claim.GetEventNonce(),
		)
	} else {
		commit() // persist transient storage
	}
}

// emitObservedEvent emits an event with information about an attestation that has been applied to
// consensus state.
func (k *Keeper) emitObservedEvent(ctx sdk.Context, att *types.Attestation, claim types.EthereumClaim) {
	ctx.EventManager().EmitTypedEvent(&types.EventAttestationObserved{
		AttestationType: claim.GetType(),
		BridgeContract:  k.GetBridgeContractAddress(ctx).Hex(),
		BridgeChainId:   k.GetBridgeChainID(ctx),
		AttestationId:   types.GetAttestationKey(claim.GetEventNonce(), claim.ClaimHash()),
		Nonce:           claim.GetEventNonce(),
	})
}

// SetAttestation sets the attestation in the store
func (k *Keeper) SetAttestation(ctx sdk.Context, eventNonce uint64, claimHash []byte, att *types.Attestation) {
	store := ctx.KVStore(k.storeKey)
	aKey := types.GetAttestationKey(eventNonce, claimHash)
	store.Set(aKey, k.cdc.MustMarshal(att))
}

// GetAttestation return an attestation given a nonce
func (k *Keeper) GetAttestation(ctx sdk.Context, eventNonce uint64, claimHash []byte) *types.Attestation {
	store := ctx.KVStore(k.storeKey)
	aKey := types.GetAttestationKey(eventNonce, claimHash)
	bz := store.Get(aKey)
	if len(bz) == 0 {
		return nil
	}

	var att types.Attestation
	k.cdc.MustUnmarshal(bz, &att)

	return &att
}

// DeleteAttestation deletes an attestation given an event nonce and claim
func (k *Keeper) DeleteAttestation(ctx sdk.Context, att *types.Attestation) {
	claim, err := k.UnpackAttestationClaim(att)
	if err != nil {
		panic("Bad Attestation in DeleteAttestation")
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAttestationKeyWithHash(claim.GetEventNonce(), claim.ClaimHash()))
}

// GetAttestationMapping returns a mapping of eventnonce -> attestations at that nonce
func (k *Keeper) GetAttestationMapping(ctx sdk.Context) (out map[uint64][]*types.Attestation) {
	out = make(map[uint64][]*types.Attestation)
	k.IterateAttestations(ctx, func(_ []byte, attestation *types.Attestation) (stop bool) {
		claim, err := k.UnpackAttestationClaim(attestation)
		if err != nil {
			panic("couldn't UnpackAttestationClaim")
		}

		eventNonce := claim.GetEventNonce()
		out[eventNonce] = append(out[eventNonce], attestation)

		return false
	})

	return
}

// IterateAttestations iterates through all attestations
func (k *Keeper) IterateAttestations(ctx sdk.Context, cb func(k []byte, v *types.Attestation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.OracleAttestationKey

	iter := store.Iterator(PrefixRange(prefix))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		attestation := types.Attestation{}

		k.cdc.MustUnmarshal(iter.Value(), &attestation)

		// cb returns true to stop early
		if cb(iter.Key(), &attestation) {
			return
		}
	}
}

// GetLastObservedValset retrieves the last observed validator set from the store
// WARNING: This value is not an up to date validator set on Ethereum, it is a validator set
// that AT ONE POINT was the one in the Gravity bridge on Ethereum. If you assume that it's up
// to date you may break the bridge
func (k *Keeper) GetLastObservedValset(ctx sdk.Context) *types.Valset {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastObservedValsetKey)

	if len(bytes) == 0 {
		return nil
	}

	valset := types.Valset{}
	k.cdc.MustUnmarshal(bytes, &valset)

	return &valset
}

// SetLastObservedValset updates the last observed validator set in the store
func (k *Keeper) SetLastObservedValset(ctx sdk.Context, valset types.Valset) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastObservedValsetKey, k.cdc.MustMarshal(&valset))
}

// GetLastObservedEventNonce returns the latest observed event nonce
func (k *Keeper) GetLastObservedEventNonce(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastObservedEventNonceKey)

	if len(bytes) == 0 {
		return 0
	}

	return types.UInt64FromBytes(bytes)
}

// GetLastObservedEthereumBlockHeight height gets the block height to of the last observed attestation from
// the store
func (k *Keeper) GetLastObservedEthereumBlockHeight(ctx sdk.Context) types.LastObservedEthereumBlockHeight {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastObservedEthereumBlockHeightKey)

	if len(bytes) == 0 {
		return types.LastObservedEthereumBlockHeight{
			CosmosBlockHeight:   0,
			EthereumBlockHeight: 0,
		}
	}

	height := types.LastObservedEthereumBlockHeight{}
	k.cdc.MustUnmarshal(bytes, &height)

	return height
}

// SetLastObservedEthereumBlockHeight sets the block height in the store.
func (k *Keeper) SetLastObservedEthereumBlockHeight(ctx sdk.Context, ethereumHeight uint64) {
	store := ctx.KVStore(k.storeKey)
	height := types.LastObservedEthereumBlockHeight{
		EthereumBlockHeight: ethereumHeight,
		CosmosBlockHeight:   uint64(ctx.BlockHeight()),
	}

	store.Set(types.LastObservedEthereumBlockHeightKey, k.cdc.MustMarshal(&height))
}

// setLastObservedEventNonce sets the latest observed event nonce
func (k *Keeper) setLastObservedEventNonce(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastObservedEventNonceKey, types.UInt64Bytes(nonce))
}

func (k *Keeper) setLastEventByValidator(ctx sdk.Context, validator sdk.ValAddress, nonce uint64, blockHeight uint64) {
	store := ctx.KVStore(k.storeKey)
	lastClaimEvent := types.LastClaimEvent{
		EthereumEventNonce:  nonce,
		EthereumEventHeight: blockHeight,
	}

	store.Set(types.GetLastEventByValidatorKey(validator), k.cdc.MustMarshal(&lastClaimEvent))
}

// GetLastEventByValidator returns the latest event for a given validator
func (k *Keeper) GetLastEventByValidator(ctx sdk.Context, validator sdk.ValAddress) (lastEvent types.LastClaimEvent) {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.GetLastEventByValidatorKey(validator))

	if len(bytes) == 0 {
		// in the case that we have no existing value this is the first
		// time a validator is submitting a claim. Since we don't want to force
		// them to replay the entire history of all events ever we can't start
		// at zero
		//
		// We could start at the LastObservedEventNonce but if we do that this
		// validator will be slashed, because they are responsible for making a claim
		// on any attestation that has not yet passed the slashing window.
		//
		// Therefore we need to return to them the lowest attestation that is still within
		// the slashing window. Since we delete attestations after the slashing window that's
		// just the lowest observed event in the store. If no claims have been submitted in for
		// params.SignedClaimsWindow we may have no attestations in our nonce. At which point
		// the last observed which is a persistent and never cleaned counter will suffice.
		lowestObservedNonce := k.GetLastObservedEventNonce(ctx)
		lowestObservedHeight := k.GetLastObservedEthereumBlockHeight(ctx)
		peggyParams := k.GetParams(ctx)
		attmap := k.GetAttestationMapping(ctx)

		// when the chain starts from genesis state, as there are no events broadcasted, lowest_observed_nonce will be zero.
		// Bridge relayer has to scan the events from the height at which bridge contract is deployed on ethereum.
		if lowestObservedNonce == 0 {
			lastEvent = types.LastClaimEvent{
				EthereumEventNonce:  lowestObservedNonce,
				EthereumEventHeight: peggyParams.BridgeContractStartHeight,
			}
			return
		}

		// no new claims in params.SignedClaimsWindow, we can return the current value
		// because the validator can't be slashed for an event that has already passed.
		// so they only have to worry about the *next* event to occur
		if len(attmap) == 0 {
			lastEvent = types.LastClaimEvent{
				EthereumEventNonce:  lowestObservedNonce,
				EthereumEventHeight: lowestObservedHeight.EthereumBlockHeight,
			}
			return
		}

		for nonce, atts := range attmap {
			for att := range atts {
				if atts[att].Observed && nonce < lowestObservedNonce {
					claim, err := k.UnpackAttestationClaim(atts[att])
					if err != nil {
						panic("could not cast to claim")
					}
					lastEvent = types.LastClaimEvent{
						EthereumEventNonce:  nonce,
						EthereumEventHeight: claim.GetBlockHeight(),
					}
				}
			}
		}

		return
	} else {
		// Unmarshall last observerd event by validator
		k.cdc.MustUnmarshal(bytes, &lastEvent)
		return
	}
}
