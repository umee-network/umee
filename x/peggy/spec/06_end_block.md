<!--
order: 5
title: End-Block
-->

# End-Block

Upon the end of each block, the following operation are performed to update queues and validator set changes. 

## Slashing

Multiple slashing conditions (validator set, batch and claim slashing) are checked in the Endblocker. 

### Validator Slashing

A validator is slashed for not signing over a validator set. The Cosmos SDK allows active validator sets to change from block to block, for this reason we need to store multiple validator sets within a single unbonding period. This allows validators to not be slashed. 

A validator will be slashed or missing a single confirmation signing.

### Batch Slashing

A validator is slashed for not signing over a batch request. A validator will be slashed for missing a batch request. 

## Attestation

Iterates through all attestations currently being voted on. Once an attestation nonce one higher than the previous one, we stop searching for an attestation and call `TryAttestation`. Once an attestation at a specific nonce has enough votes all the other attestations will be skipped and the `lastObservedEventNonce` incremented.

## Cleanup

Cleanup loops through batches in order to clean up the timed out transactions.

### Batches

When a batch of transactions are created they have a specified height of the opposing chain for when the batch becomes invalid. When this happens we must remove them from the store. At the end of every block, we loop through the store of batches checking the timeout heights. 
