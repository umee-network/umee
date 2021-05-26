# ADR 001: Umee Module

## Changelog

- April 28, 2021: Initial Draft (@alexanderbez)

## Status

Proposed

## Context

Umee is a Universal Capital Facility that can collateralize assets on one
blockchain towards borrowing assets on another blockchain. The platform specializes
in allowing staked assets from POS blockchains to be used as collateral for
borrowing across blockchains. The platform uses a combination of algorithmically
determined interest rates based on market driven conditions.

For the initial MVP implementation of the Umee network, we require the ability
for users to be able to send ATOM tokens to a dedicated pool in the Umee network
via IBC. ATOM tokens deposited into the Umee pool will mint a derivative meToken
in a one-to-one ratio, i.e. one ATOM mints one meToken derivative.

Validators on the Umee network will take these deposited ATOM tokens and delegate
them to a set of governance-controlled validators on the ATOM source chain,
e.g. the Cosmos Hub.

A user that sent ATOM tokens to the Umee network can then take their derivative
meTokens and send them to Ethereum via a bridge where a synthetic ERC-20 version
of the meTokens are minted. Once sent, the meTokens are locked in Umee and the
user can then freely trade and operate within Ethereum's DeFi ecosystem with the
synthetic ERC-20 meTokens.

## Decision

To support users sending ATOM tokens to the Umee network via IBC, we will create
a `x/umee` module that contains a dedicated `ModuleAccount` that represents the
ATOM pool. In addition, we will implement a [custom IBC module](https://github.com/cosmos/ibc-go/blob/v1.0.0-alpha1/docs/custom.md).
The custom IBC module will handle receiving custom packets on a custom Umee port.
The custom module will be nearly identical to the ICS-20 module, except that it
will perform the following:

- Automatically send source ATOM tokens to the dedicated `ModuleAccount` pool
  address instead of a user supplied address.
- Mark the transfer's amount in a pending queue, `PendingDelegationQueue`, to be
  later used for delegation on the source chain.
- Track the sender and amount in the module's state.
- Mint derivative meTokens and send them to the source sender's Umee account.

At the beginning of a staking epoch, on the order of a few hours (to be determined),
E<sub>1</sub> in `BeginBlock`, the `x/umee` module will perform the following:

- Move the funds from the `PendingDelegationQueue` into another queue,
  `PrevPendingDelegationQueue`. This queue is locked as funds cannot be sent to
  it. In addition, these are the funds that will be delegated to the source chain
  during the current staking epoch.
- Clear out `PendingDelegationQueue`.

The validators in the current staking epoch, E<sub>1</sub>, will then construct,
sign, and broadcast multisig transactions to the Umee network. Once enough
transactions are received by the `x/umee` module, the `x/umee` module will store
the constructed multisig transaction and emit an event signalling a delegation
transaction can be made to the source chain. Delegations will be made to
validators controlled by Umee governance.

The multisig will be constructed such that all constituents are the validators
in the current epoch. The threshold of the multisig will be the total number of
validators in the epoch. This means all validators will have to sign and broadcast
their part of the multisig. If any signature is missing, a cross delegation
cannot be made. We do not envision a penalty for missing multisig transactions
but this can be revised in the future in addition to an incentive mechanism.

The multisig account must exist on native ATOM chain, which can already exist
from a previous epoch or if a validator set changes in an epoch. When an account
does not exist on the native ATOM chain, it will have to be created by Umee via
IBC by sending some tokens to that account.

A separate relayer process will not only be responsible for relaying IBC packets
to and from the source chain and Umee, but it will also be responsible for
listening for these delegation events, and broadcasting the constructed multisig
transaction to the source chain.

Any new transfers during E<sub>1</sub> will be processed in the same manner during
E<sub>2</sub>.

## Consequences

### Positive

- Allows for cross-chain delegations in a decentralized manner.

### Negative

- Requires the maintenance of a validator controlled multisig account each epoch.
- Requires a relayer process to listen for unique events and broadcast
  delegation transactions.
- Transferred ATOMs to the Umee network are not delegated until the next epoch
  and thus do not earn rewards until the next epoch.
  
### Neutral

## References
