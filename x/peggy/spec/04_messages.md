<!--
order: 4
title: Messages
-->

# Messages

This is a reference document for Peggy message types. For code reference and exact arguments see the [proto definitions](https://github.com/InjectiveLabs/injective-core/blob/master/proto/injective/peggy/v1/msgs.proto). 

## User messages

These are messages sent on the Injective Chain peggy module. See [workflow](./02_workflow.md) for a more detailed summary of the entire deposit and withdraw process.

### SendToEth

```go

// MsgSendToEth
// This is the message that a user calls when they want to bridge an asset
// it will later be removed when it is included in a batch and successfully
// submitted tokens are removed from the users balance immediately
// -------------
// AMOUNT:
// the coin to send across the bridge, note the restriction that this is a
// single coin not a set of coins that is normal in other Injective messages
// FEE:
// the fee paid for the bridge, distinct from the fee paid to the chain to
// actually send this message in the first place. So a successful send has
// two layers of fees for the user
type MsgSendToEth struct {
	Sender    string     
	EthDest   string     
	Amount    types.Coin 
	BridgeFee types.Coin 
}

```
SendToEth allows the user to specify an Ethereum destination, a token to send to Ethereum and a fee denominated in that same token
to pay the relayer. Note that this transaction will contain two fees. One fee amount to submit to the Injective Chain, that can be paid
in any token and one fee amount for the Ethereum relayer that must be paid in the same token that is being bridged.

### CancelSendToEth
```go
// This call allows the sender (and only the sender)
// to cancel a given MsgSendToEth and receive a refund
// of the tokens
type MsgCancelSendToEth struct {
	TransactionId uint64 
	Sender        string 
}

```
CancelSendToEth allows a user to retrieve a transaction that is in the batch pool but has not yet been packaged into a transaction batch by a relayer running [RequestBatch](./04_messages.md#RequestBatch). 

### SubmitBadSignatureEvidence

```go
// This call allows anyone to submit evidence that a
// validator has signed a valset or batch that never
// existed. Subject contains the batch or valset.
type MsgSubmitBadSignatureEvidence struct {
	Subject   *types1.Any 
	Signature string      
	Sender    string      
}
```

SubmitBadSignatureEvidence allows anyone to submit evidence that a validator has signed a valset or batch that never existed.

## Relayer Messages

These are messages run by relayers. Relayers are unpermissioned and simply work to move things from the Injective Chain to Ethereum.

### RequestBatch

```go
// MsgRequestBatch
// this is a message anyone can send that requests a batch of transactions to
// send across the bridge be created for whatever block height this message is
// included in. This acts as a coordination point, the handler for this message
// looks at the AddToOutgoingPool tx's in the store and generates a batch, also
// available in the store tied to this message. The validators then grab this
// batch, sign it, submit the signatures with a MsgConfirmBatch before a relayer
// can finally submit the batch
// -------------
type MsgRequestBatch struct {
	Orchestrator string 
	Denom        string
}
```

Relayers use `QueryPendingSendToEth` in [query.proto](https://github.com/InjectiveLabs/injective-core/blob/master/proto/injective/peggy/v1/query.proto) to query the potential fees for a batch of each
token type. When they find a batch that they wish to relay they send in a RequestBatch message and the Peggy module creates a batch.

This then triggers the Ethereum Signers to send in ConfirmBatch messages, which the signatures required to submit the batch to the Ethereum chain.

At this point any relayer can package these signatures up into a transaction and send them to Ethereum.

As noted above this message is unpermissioned and it is safe to allow anyone to call this message at any time. 

## Oracle Messages

All validators run two processes in addition to their Injective node. An Ethereum oracle and Ethereum signer, these are bundled into a single Orchestrator binary for ease of use.

The oracle observes the Ethereum chain for events from the [Peggy.sol](https://github.com/InjectiveLabs/peggo/blob/master/solidity/contracts/Peggy.sol) contract before submitting them as messages to the Injective Chain.

### DepositClaim
```go

// EthereumBridgeDepositClaim
// When more than 66% of the active validator set has
// claimed to have seen the deposit enter the ethereum blockchain coins are
// issued to the Injective address in question
// -------------
type MsgDepositClaim struct {
	EventNonce     uint64                                
	BlockHeight    uint64                                
	TokenContract  string                                
	Amount         sdk.Int 
	EthereumSender string                                 
	CosmosReceiver string                                 
	Orchestrator   string                                 
}
```
Deposit claims represent a `SendToCosmosEvent` emitted by the Peggy contract. After 2/3 of the validators confirm a deposit claim,  the representative tokens will be issued to the specified `CosmosReceiver` Injective Chain account.

### WithdrawClaim

```go
// WithdrawClaim claims that a batch of withdrawal
// operations on the bridge contract was executed.
type MsgWithdrawClaim struct {
	EventNonce    uint64 
	BlockHeight   uint64 
	BatchNonce    uint64 
	TokenContract string 
	Orchestrator  string 
}
```
Withdraw claims represent a `TransactionBatchExecutedEvent` from the Peggy contract. When this passes the oracle vote the batch in state is cleaned up and tokens are burned/locked.

### ValsetUpdateClaim

```go

// This informs the peggy module that a validator
// set has been updated.
type MsgValsetUpdatedClaim struct {
	EventNonce   uint64                           
	ValsetNonce  uint64                           
	BlockHeight  uint64                           
	Members      []*BridgeValidator               
	RewardAmount sdk.Int 
	RewardToken  string                                 
	Orchestrator string                                 
}
```
claim representing a `ValsetUpdatedEvent` from the Peggy contract. When this passes the oracle vote reward amounts are tallied and minted.

### ERC20DeployedClaim
```go

// ERC20DeployedClaim allows the peggy module
// to learn about an ERC-20 that someone deployed
// to represent a Cosmos asset
type MsgERC20DeployedClaim struct {
	EventNonce    uint64 
	BlockHeight   uint64 
	CosmosDenom   string 
	TokenContract string 
	Name          string 
	Symbol        string 
	Decimals      uint64 
	Orchestrator  string 
}
```
claim representing a `ERC20DeployedEvent` from the Peggy contract. When this passes the oracle vote it is checked for accuracy and adopted or rejected as the ERC-20 representation of a Cosmos SDK based asset. 

## Ethereum Signer Messages

All validators run two processes in addition to their Injective Chain node. An Ethereum oracle and Ethereum signer, these are bundled into a single Orchestrator binary for ease of use.

The Ethereum signer watches several [query endpoints](https://github.com/InjectiveLabs/injective-core/blob/master/proto/injective/peggy/v1/query.proto) and it's only job is to submit a signature for anything that appears on those endpoints. For this reason the validator must provide a secure RPC to an Injective Chain node following chain consensus. Or they risk being tricked into signing the wrong thing.

### ConfirmBatch
```go

// MsgConfirmBatch
// When validators observe a MsgRequestBatch they form a batch by ordering
// transactions currently in the txqueue in order of highest to lowest fee,
// cutting off when the batch either reaches a hardcoded maximum size (to be
// decided, probably around 100) or when transactions stop being profitable
// This message includes the batch as well as an Ethereum signature over this batch by the validator
// -------------
type MsgConfirmBatch struct {
	Nonce         uint64 
	TokenContract string 
	EthSigner     string 
	Orchestrator  string 
	Signature     string 
}
```
Submits an Ethereum signature over a batch appearing in the `LastPendingBatchRequestByAddr` query.

### ValsetConfirm
```go

// MsgValsetConfirm
// this is the message sent by the validators when they wish to submit their
// signatures over the validator set at a given block height. A validator must
// first call MsgSetEthAddress to set their Ethereum address to be used for
// signing. Then someone (anyone) must make a ValsetRequest the request is
// essentially a messaging mechanism to determine which block all validators
// should submit signatures over. Finally validators sign the validator set,
// powers, and Ethereum addresses of the entire validator set at the height of a
// ValsetRequest and submit that signature with this message.
type MsgValsetConfirm struct {
	Nonce        uint64 
	Orchestrator string 
	EthAddress   string 
	Signature    string 
}
```
Submits an Ethereum signature over a batch appearing in the `LastPendingValsetRequestByAddr` query.

## Validator Messages

These are messages sent directly using the validator's message key.

### SetOrchestratorAddresses

```go

// MsgSetOrchestratorAddresses
// this message allows validators to delegate their voting responsibilities
// to a given key. This key is then used as an optional authentication method
// for sigining oracle claims
// VALIDATOR
// The validator field is a injvaloper1... string (i.e. sdk.ValAddress)
// that references a validator in the active set
// ORCHESTRATOR
// The orchestrator field is a inj1... string  (i.e. sdk.AccAddress) that
// references the key that is being delegated to
// ETH_ADDRESS
// This is a hex encoded 0x Ethereum public key that will be used by this validator
// on Ethereum
type MsgSetOrchestratorAddresses struct {
	Sender       string
	Orchestrator string
	EthAddress   string
}
```
This message sets the Orchestrator's delegate keys. 

