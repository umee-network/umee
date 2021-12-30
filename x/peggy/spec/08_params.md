<!--
order: 8
title: Parameters
-->

# Params

This document describes and advises configuration of the Peggy module's parameters. The default parameters can be found in the genesis.go of the peggy module.

```go
type Params struct {
	PeggyId                       string                                 
	ContractSourceHash            string                                 
	BridgeEthereumAddress         string                                 
	BridgeChainId                 uint64                                 
	SignedValsetsWindow           uint64                                 
	SignedBatchesWindow           uint64                                 
	TargetBatchTimeout            uint64                                 
	AverageBlockTime              uint64                                 
	AverageEthereumBlockTime      uint64                                 
	SlashFractionValset           sdk.Dec 
	SlashFractionBatch            sdk.Dec 
	SlashFractionClaim            sdk.Dec 
	SlashFractionConflictingClaim sdk.Dec 
	UnbondSlashingValsetsWindow   uint64  
	SlashFractionBadEthSignature  sdk.Dec 
	CosmosCoinDenom               string  
	CosmosCoinErc20Contract       string  
	BridgeContractStartHeight     uint64  
	ValsetReward                  types.Coin
}
```

## `peggy_id`

A random 32 byte value to prevent signature reuse, for example if the
Injective Chain validators decided to use the same Ethereum keys for another chain
also running Peggy we would not want it to be possible to play a deposit
from chain A back on chain B's Peggy. This value IS USED ON ETHEREUM so
it must be set in your genesis.json before launch and not changed after
deploying Peggy. Changing this value after deploying Peggy will result
in the bridge being non-functional. To recover just set it back to the original
value the contract was deployed with.

## `contract_source_hash`

The code hash of a known good version of the Peggy contract
solidity code. This can be used to verify the correct version
of the contract has been deployed. This is a reference value for
governance action only it is never read by any Peggy code

## `bridge_ethereum_address`

is address of the bridge contract on the Ethereum side, this is a
reference value for governance only and is not actually used by any
Peggy module code.

The Ethereum bridge relayer use this value to interact with Peggy contract for querying events and submitting valset/batches to Peggy contract.

## `bridge_chain_id`

The bridge chain ID is the unique identifier of the Ethereum chain. This is a reference value only and is not actually used by any Peggy code

These reference values may be used by future Peggy client implementations to allow for consistency checks.

## Signing windows

* `signed_valsets_window`
* `signed_batches_window`
* `signed_claims_window`

These values represent the time in blocks that a validator has to submit
a signature for a batch or valset, or to submit a claim for a particular
attestation nonce.

In the case of attestations this clock starts when the
attestation is created, but only allows for slashing once the event has passed.
Note that that claims slashing is not currently enabled see [slashing spec](./05_slashing.md)

## `target_batch_timeout`

This is the 'target' value for when batches time out, this is a target because
Ethereum is a probabilistic chain and you can't say for sure what the block
frequency is ahead of time.

## Ethereum timing

* `average_block_time`
* `average_ethereum_block_time`

These values are the average Injective Chain block time and Ethereum block time respectively
and they are used to compute what the target batch timeout is. It is important that
governance updates these in case of any major, prolonged change in the time it takes
to produce a block

## Slash fractions

* `slash_fraction_valset`
* `slash_fraction_batch`
* `slash_fraction_claim`
* `slash_fraction_conflicting_claim`

The slashing fractions for the various peggy related slashing conditions. The first three
refer to not submitting a particular message, the third for failing to submit a claim and the last for submitting a different claim than other validators.

Note that claim slashing is currently disabled as outlined in the [slashing spec](./05_slashing.md)

## `valset_reward`

Valset reward is the reward amount paid to a relayer when they relay a valset to the Peggy contract on Ethereum.