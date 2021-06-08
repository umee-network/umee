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
a custom `x/transfer` module that is based on the [ICS-20](https://github.com/cosmos/ibc/tree/master/spec/app/ics-020-fungible-token-transfer)
standard. However, instead of creating a completely [custom IBC module](https://github.com/cosmos/ibc-go/blob/v1.0.0-alpha1/docs/custom.md),
we will embed the the `ICS-20` module in `x/transfer` and override the `OnRecvPacket`
to implement custom logic.

Namely, we will mimic `OnRecvPacket` from `ICS-20` in all cases except where
minted vouchers are sent to the receiving address. Instead of sending minted
vouchers to the receiving address, we will execute a hook, `OnMintedVouchers`,
defined in the `x/umee` module.

The `x/umee` module will contain the core business logic of Umee's capital
facilities. When a user transfers ATOM tokens to Umee and the `OnMintedVouchers`
hook is executed by the custom `x/transfer` module, the `x/umee` module will
perform the following:

- Automatically send source ATOM tokens to the dedicated `ModuleAccount` pool
  address instead of a the original transfer's receiving address.
- Mark the transfer's amount in a pending queue, `PendingDelegationQueue`, to be
  later used for cross-chain delegation on the source ATOM chain.
- Track the sender and amount in the module's state.
- Mint derivative meTokens and send them to the source sender's Umee account.

The `x/umee` module's `ModuleAccount` will be represented as an [Interchain Account](https://github.com/cosmos/ibc/tree/master/spec/app/ics-027-interchain-accounts). This account will exist and be managed on the Umee
network while being registered on the destination ATOM chain as an `Interchain Account`.
The `ModuleAccount` will have permission to delegate funds on behalf of Umee users
via the `Interchain Account` ICS standard.

At the beginning of each `BeginBlock`, the `x/umee` module will check if the
`ModuleAccount` pool address has been registered on the destination ATOM chain.
If not, in `BeginBlock` we will register the account via `RegisterIBCAccount`.

```go
func (keeper Keeper) RegisterIBCAccount(ctx sdk.Context, owner sdk.AccAddress, srcPort, srcChannel string) error {
  salt := keeper.GetIncrementalSalt(ctx)
  err := keeper.iaKeeper.TryRegisterIBCAccount(ctx, sourcePort, sourceChannel, []byte(salt))
  if err != nil {
    return err
  }

  keeper.PushAddressToRegistrationQueue(ctx, sourcePort, sourceChannel, owner)
  ctx.EventManager().EmitEvent(sdk.NewEvent("register-interchain-account",sdk.NewAttribute("salt", salt)))

  return nil
}
```

Note, we push the account into a queue via, `PushAddressToRegistrationQueue`, so
that we can implement a hook where on packet ACK, the `x/umee` module can store
and mark the account as registered<sup>1</sup>.

Once the `ModuleAccount` pool address is registered, we are now able to perform
cross-chain delegations on the destination ATOM chain. During each `EndBlock`
execution in the `x/umee` module, we will construct a series of messages that will
be executed on the source ATOM chain via `TryRunTx` as defined by the
`Interchain Account` ICS standard. Finally, the `PendingDelegationQueue` will be
cleared out.

The first message will be to send all ATOM tokens currently in the `PendingDelegationQueue`
to the Interchain account registered on the ATOM source chain. The next series of
messages will be delegation messages to a set of validators on the ATOM source
chain. This set of validators and the delegation ratio for each validator, will
be controlled by Umee governance and will be maintained as a parameter in the
`x/umee` module.

```go
type ValidatorDelegation struct {
  Validator string
  Ratio     string
}

type ValidatorDelegationSet struct {
  Validators []ValidatorDelegation
}
```

Example:

```json
{
  "validators": [
    {
      "validator": "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0",
      "ratio": "0.50"
    },
    {
      "validator": "cosmosvaloper14kn0kk33szpwus9nh8n87fjel8djx0y070ymmj",
      "ratio": "0.50"
    }
  ]
}
```

Once the transaction is relayed to the ATOM source chain, `OnTxSucceeded` will
be executed by the `Interchain Account` module. The `x/umee` module will implement
a hook such that when this method is called, we mark the cross-chain delegation
as complete. Note, subsequent cross-chain delegations in `EndBlock` cannot be
made until the previous one has completed.

Note, Umee will also make use of epoch-based staking, however, this will have no
bearing on when cross-chain delegations can be made.

> <sup>1</sup> Once the [Interchain Account](https://github.com/cosmos/ibc/tree/master/spec/app/ics-027-interchain-accounts)
> ICS standard is updated to use a single channel per account, we will no longer
> need to use a queue and the `Interchain Account` ICS standard will take care of
> this for us.

### Validator Delegation Set Parameter Changes

Users, via Umee's governance, can decide to change or update the validator
delegation set via a `ParameterChangeProposal`. When this happens, we need to
ensure the cross-chain delegations accurately reflects the new set of validators.

## Open Questions

1. How will Umee handle situations where one or more validators in the validator
   delegation set leaves the validator set in the ATOM chain?
2. How will Umee handle situations where one or more validators in the validator
   delegation set gets slashed for misbehavior in the ATOM chain?
3. How will users reclaim their original ATOMs?
4. How will Umee handle the situation where the cross-chain delegation cannot be
   made for indefinite period of time due to faulty relayers, liveness issues or
   other errors on the ATOM-based chain?

## Consequences

### Positive

- Allows for cross-chain delegations in a decentralized manner.

### Negative

### Neutral

- Requires a relayer process to listen for unique events and broadcast
  delegation transactions.

## References

- [ICS-20](https://github.com/cosmos/ibc/tree/master/spec/app/ics-020-fungible-token-transfer)
- [Interchain Accounts](https://github.com/cosmos/ibc/tree/master/spec/app/ics-027-interchain-accounts)
- [Custom IBC Modules](https://github.com/cosmos/ibc-go/blob/v1.0.0-alpha1/docs/custom.md)
