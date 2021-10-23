<!--
order: 5
title: Slashing
-->

# Slashing
### Security Concerns

The **Validator Set** is the actual set of keys with stake behind them, which are slashed for double-signs or other
misbehavior. We typically consider the security of a chain to be the security of a _Validator Set_. This varies on
each chain, but is our gold standard. Even IBC offers no more security than the minimum of both involved Validator Sets.

The **Eth bridge relayer** is a binary run alongside the main `injectived` daemon by the validator set. It exists purely as a matter of code organization and is in charge of signing Ethereum transactions, as well as observing events on Ethereum and bringing them into the Injective Chain state. It signs transactions bound for Ethereum with an Ethereum key, and signs over events coming from Ethereum with an Injective Chain account key. We can add slashing conditions to any mis-signed message by any _Eth Signer_ run by the _Validator Set_ and be able to provide the same security as the _Validator Set_, just a different module detecting evidence of malice and deciding how much to slash. If we can prove a transaction signed by any _Eth Signer_ of the _Validator Set_ was illegal or malicious, then we can slash on the Injective Chain side and potentially provide 100% of the security of the _Validator Set_. Note that this also has access to the 3 week unbonding period to allow evidence to slash even if they immediately unbond.

Below are various slashing conditions we use in Peggy.

## PEGGYSLASH-01: Signing fake validator set or tx batch evidence

This slashing condition is intended to stop validators from signing over a validator set and nonce that has never existed on the Injective Chain. It works via an evidence mechanism, where anyone can submit a message containing the signature of a validator over a fake validator set. This is intended to produce the effect that if a cartel of validators is formed with the intention of submitting a fake validator set, one defector can cause them all to be slashed.
```go
// This call allows anyone to submit evidence that a
// validator has signed a valset, batch, or logic call that never
// existed. Subject contains the batch, valset, or logic call.
type MsgSubmitBadSignatureEvidence struct {
	Subject   *types1.Any 
	Signature string      
	Sender    string      
}
```
**Implementation considerations:**

The trickiest part of this slashing condition is determining that a validator set has never existed on Injective. To save space, we will need to clean up old validator sets. We could keep a mapping of validator set hash to true in the KV store, and use that to check if a validator set has ever existed. This is more efficient than storing the whole validator set, but its growth is still unbounded. It might be possible to use other cryptographic methods to cut down on the size of this mapping. It might be OK to prune very old entries from this mapping, but any pruning reduces the deterrence of this slashing condition.

The implemented version of this slashing condition stores a map of hashes for all past events, this is smaller than storing entire batches or validator sets and doesn't have to be accessed as frequently. A possible but not currently implemented efficiency optimization would be to remove hashes from this list after a given period. But this would require storing more metadata about each hash.

Currently automatic evidence submission is not implemented in the relayer. By the time a signature is visible on Ethereum it's already too late for slashing to prevent bridge highjacking or theft of funds. Furthermore since 66% of the validator set is required to perform this action anyways that same controlling majority could simply refuse the evidence. The most common case envisioned for this slashing condition is to break up a cabal of validators trying to take over the bridge by making it more difficult for them to trust one another and actually coordinate such a theft.

The theft would involve exchanging of slashable Ethereum signatures and open up the possibility of a manual submission of this message by any defector in the group.

Currently this is implemented as an ever growing array of hashes in state.

## PEGGYSLASH-02: Failure to sign tx batch

This slashing condition is triggered when a validator does not sign a transaction batch within `SignedBatchesWindow` upon it's creation by the Peggy module. This prevents two bad scenarios-

1. A validator simply does not bother to keep the correct binaries running on their system,
2. A cartel of >1/3 validators unbond and then refuse to sign updates, preventing any batches from getting enough signatures to be submitted to the Peggy Ethereum contract.

## PEGGYSLASH-03: Failure to sign validator set update

This slashing condition is triggered when a validator does not sign a validator set update which is produced by the Peggy  module. This prevents two bad scenarios-

1. A validator simply does not bother to keep the correct binaries running on their system,
2. A cartel of >1/3 validators unbond and then refuse to sign updates, preventing any validator set updates from getting enough signatures to be submitted to the Peggy Ethereum contract. If they prevent validator set updates for longer than the Injective Chain unbonding period, they can no longer be punished for submitting fake validator set updates and tx batches (PEGGYSLASH-01 and PEGGYSLASH-03).

To deal with scenario 2, PEGGYSLASH-03 will also need to slash validators who are no longer validating, but are still in the unbonding period for up to `UnbondSlashingValsetsWindow` blocks. This means that when a validator leaves the validator set, they will need to keep running their equipment for at least `UnbondSlashingValsetsWindow` blocks. This is unusual for the Injective Chain, and may not be accepted by the validators.

The current value of `UnbondSlashingValsetsWindow` is 25,000 blocks, or about 12-14 hours. We have determined this to be a safe value based on the following logic. So long as every validator leaving the validator set signs at least one validator set update that they are not contained in then it is guaranteed to be possible for a relayer to produce a chain of validator set updates to transform the current state on the chain into the present state.

It should be noted that both PEGGYSLASH-02 and PEGGYSLASH-03 could be eliminated with no loss of security if it where possible to perform the Ethereum signatures inside the consensus code. This is a pretty limited feature addition to Tendermint that would make Peggy far less prone to slashing.

## PEGGYSLASH-04: Submitting incorrect Eth oracle claim (Disabled for now)

The Ethereum oracle code (currently mostly contained in attestation.go), is a key part of Peggy. It allows the Peggy module to have knowledge of events that have occurred on Ethereum, such as deposits and executed batches. PEGGYSLASH-03 is intended to punish validators who submit a claim for an event that never happened on Ethereum.

**Implementation considerations**

The only way we know whether an event has happened on Ethereum is through the Ethereum event oracle itself. So to implement this slashing condition, we slash validators who have submitted claims for a different event at the same nonce as an event that was observed by >2/3s of validators.

Although well-intentioned, this slashing condition is likely not advisable for most applications of Peggy. This is because it ties the functioning of the Injective Chain which it is installed on to the correct functioning of the Ethereum chain. If there is a serious fork of the Ethereum chain, different validators behaving honestly may see different events at the same event nonce and be slashed through no fault of their own. Widespread unfair slashing would be very disruptive to the social structure of the Injective Chain.

Maybe PEGGYSLASH-04 is not necessary at all:

The real utility of this slashing condition is to make it so that, if >2/3 of the validators form a cartel to all submit a fake event at a certain nonce, some number of them can defect from the cartel and submit the real event at that nonce. If there are enough defecting cartel members that the real event becomes observed, then the remaining cartel members will be slashed by this condition. However, this would require >1/2 of the cartel members to defect in most conditions.

If not enough of the cartel defects, then neither event will be observed, and the Ethereum oracle will just halt. This is a much more likely scenario than one in which PEGGYSLASH-04 is actually triggered.

Also, PEGGYSLASH-04 will be triggered against the honest validators in the case of a successful cartel. This could act to make it easier for a forming cartel to threaten validators who do not want to join.

## PEGGYSLASH-05: Failure to submit Eth oracle claims (Disabled for now)

This is similar to PEGGYSLASH-04, but it is triggered against validators who do not submit an oracle claim that has been observed. In contrast to PEGGYSLASH-04, PEGGYSLASH-05 is intended to punish validators who stop participating in the oracle completely.

**Implementation considerations**

Unfortunately, PEGGYSLASH-05 has the same downsides as PEGGYSLASH-04 in that it ties the correct operation of the Injective Chain to the Ethereum chain. Also, it likely does not incentivize much in the way of correct behavior. To avoid triggering PEGGYSLASH-05, a validator simply needs to copy claims which are close to becoming observed. This copying of claims could be prevented by a commit-reveal scheme, but it would still be easy for a "lazy validator" to simply use a public Ethereum full node or block explorer, with similar effects on security. Therefore, the real usefulness of PEGGYSLASH-05 is likely minimal

PEGGYSLASH-05 also introduces significant risks. Mostly around forks on the Ethereum chain. For example recently OpenEthereum failed to properly handle the Berlin hardfork, the resulting node 'failure' was totally undetectable to automated tools. It didn't crash so there was no restart to perform, blocks where still being produced although extremely slowly. If this had occurred while Peggy was running with PEGGYSLASH-05 active it would have caused those validators to be removed from the set. Possibly resulting in a very chaotic moment for the chain as dozens of validators where removed for little to no fault of their own.

Without PEGGYSLASH-04 and PEGGYSLASH-05, the Ethereum event oracle only continues to function if >2/3 of the validators voluntarily submit correct claims. Although the arguments against PEGGYSLASH-04 and PEGGYSLASH-05 are convincing, we must decide whether we are comfortable with this fact. Alternatively we must be comfortable with the Injective Chain potentially halting entirely due to Ethereum generated factors.

