<!--
order: 10
title: Future Improvements
-->

# Future Improvements



### Native Ethereum Signing

Validators run a required `Eth Signer` in the peggo orchestrator because we can not yet insert this sort of simple signature logic into Cosmos SDK based chains without significant modification to Tendermint. This may be possible in the future with [modifications to Tendermint](https://github.com/tendermint/tendermint/issues/6066).

It should be noted that both [PEGGYSLASH-02](./05_slashing.md) and [PEGGYSLASH-03](./05_slashing.md) could be eliminated with no loss of security if it where possible to perform the Ethereum signatures inside the consensus code. This is a pretty limited feature addition to Tendermint that would make Peggy far less prone to slashing.

### Improved Validator Incentives

Currently validators in Peggy have only one carrot - the extra activity brought to the chain by a functioning bridge.

There are on the other hand a lot of negative incentives (sticks) that the validators must watch out for. These are outlined in the [slashing spec](./05_slashing.md).

One negative incentive that is not covered under slashing is the cost of submitting oracle submissions and signatures. Currently these operations are not incentivized, but still cost the validators fees to submit. This isn't a severe issue considering the relatively cheap transaction fees on the Injective Chain currently, but of course is an important factor to consider as transaction fees rise. 

Some positive incentives for correctly participating in the operation of the bridge should be under consideration. In addition to eliminating the fees for mandatory submissions.
