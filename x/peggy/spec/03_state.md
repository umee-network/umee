<!--
order: 3
title: State
-->

# State

## Params

Params is a module-wide configuration structure that stores system parameters and defines overall functioning of the peggy module. Detailed specification for each parameter can be found in the [Parameters section](08_params.md). 

- Params: `Paramsspace("peggy") -> legacy_amino(params)`

### Validator Info

#### Ethereum Address by Validator 

Stores each Validator's corresponding delegate Ethereum address indexed by the validator's account address. 

| key          | Value | Type   | Encoding               |
|--------------|-------|--------|------------------------|
| `[]byte{0x1} + []byte(validatorAddr)` | Ethereum address | `common.Address` | Protobuf encoded |

#### Validator by Ethereum Address

Stores each Validator's account address indexed by Ethereum address. 

| key          | Value | Type   | Encoding               |
|--------------|-------|--------|------------------------|
| `[]byte{0xfb} + []byte(ethAddress)` | Validator address | `sdk.ValAddress` | Protobuf encoded |

### OutgoingTxBatch

Stored in two possible ways, first with a height and second without (unsafe). Unsafe is used for testing and export and import of state.
Currently [Peggy.sol](https://github.com/InjectiveLabs/peggo/blob/master/solidity/contracts/Peggy.sol) is hardcoded to only accept batches with a single token type and only pay rewards in that same token type.

```go
// OutgoingTxBatch represents a batch of transactions going from Peggy to ETH
type OutgoingTxBatch struct {
	BatchNonce    uint64               
	BatchTimeout  uint64               
	Transactions  []*OutgoingTransferTx 
	TokenContract string                
	Block         uint64               
}
```

| key          | Value | Type   | Encoding               |
|--------------|-------|--------|------------------------|
| `[]byte{0xa} + []byte(tokenContract) + nonce (big endian encoded)` | A batch of outgoing transactions | `types.OutgoingTxBatch` | Protobuf encoded |

### ValidatorSet

This is the validator set of the bridge.

Stored in two possible ways, first with a height and second without (unsafe). Unsafe is used for testing and export and import of state.

```go
// Valset is the Ethereum Bridge Multsig Set, each peggy validator also
// maintains an ETH key to sign messages, these are used to check signatures on
// ETH because of the significant gas savings
type Valset struct {
	Nonce        uint64                               
	Members      []*BridgeValidator                   
	Height       uint64                               
	RewardAmount sdk.Int 
	// the reward token in it's Ethereum hex address representation
	RewardToken string
}

```

| key          | Value | Type   | Encoding               |
|--------------|-------|--------|------------------------|
| `[]byte{0x2} + nonce (big endian encoded)` | Validator set | `types.Valset` | Protobuf encoded |

### SlashedValsetNonce

The latest validator set slash nonce. This is used to track which validator set needs to be slashed and which already has been.

| Key            | Value | Type   | Encoding               |
|----------------|-------|--------|------------------------|
| `[]byte{0xf5}` | Nonce | uint64 | encoded via big endian |

### ValsetNonce

The latest validator set nonce, this value is updated on every write. 

| key          | Value | Type   | Encoding               |
|--------------|-------|--------|------------------------|
| `[]byte{0xf6}` | Nonce | `uint64` | encoded via big endian |


### Valset Confirmation

When a validator signs over a validator set this is considered a `valSetConfirmation`, these are saved via the current nonce and the orchestrator address. 

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
// -------------
type MsgValsetConfirm struct {
	Nonce        uint64 
	Orchestrator string 
	EthAddress   string 
	Signature    string 
}
```

| Key                                         | Value                  | Type                     | Encoding         |
|---------------------------------------------|------------------------|--------------------------|------------------|
| `[]byte{0x3} + (nonce + []byte(AccAddress)` | Validator Confirmation | `types.MsgValsetConfirm` | Protobuf encoded |

### ConfirmBatch

When a validator confirms a batch it is added to the confirm batch store. It is stored using the orchestrator, token contract and nonce as the key. 

```go
// MsgConfirmBatch
type MsgConfirmBatch struct {
	Nonce         uint64 
	TokenContract string 
	EthSigner     string 
	Orchestrator  string 
	Signature     string
}

```
| Key                                                                 | Value                        | Type                    | Encoding         |
|---------------------------------------------------------------------|------------------------------|-------------------------|------------------|
| `[]byte{0xe1} + []byte(tokenContract) + nonce + []byte(AccAddress)` | Validator Batch Confirmation | `types.MsgConfirmBatch` | Protobuf encoded |

### OrchestratorValidator

When a validator would like to delegate their voting power to another key. The value is stored using the orchestrator address as the key

| Key                                 | Value                                        | Type     | Encoding         |
|-------------------------------------|----------------------------------------------|----------|------------------|
| `[]byte{0xe8} + []byte(AccAddress)` | Orchestrator address assigned by a validator | `[]byte` | Protobuf encoded |

### EthAddress

A validator has an associated counter chain address. 

| Key                                 | Value                                        | Type     | Encoding         |
|-------------------------------------|----------------------------------------------|----------|------------------|
| `[]byte{0x1} + []byte(ValAddress)` | Ethereum address assigned by a validator | `[]byte` | Protobuf encoded |

### OutgoingTransferTx

Sets an outgoing transactions into the applications transaction pool to be included into a batch. 
```go
// OutgoingTransferTx represents an individual send from Peggy to ETH
type OutgoingTransferTx struct {
	Id          uint64     
	Sender      string     
	DestAddress string     
	Erc20Token  *ERC20Token 
	Erc20Fee    *ERC20Token 
}
```

| Key                                 | Value                                        | Type     | Encoding         |
|-------------------------------------|----------------------------------------------|----------|------------------|
| `[]byte{0x6} + id (big endian encoded)` | User created transaction to be included in a batch | `types.OutgoingTransferTx` | Protobuf encoded |

### IDS

### SlashedBlockHeight

Represents the latest slashed block height. There is always only a singe value stored. 

| Key                                 | Value                                        | Type     | Encoding         |
|-------------------------------------|----------------------------------------------|----------|------------------|
| `[]byte{0xf7}` | Latest height a batch slashing occurred | `uint64` | Big endian encoded |

### TokenContract & Denom

A denom that is originally from a counter chain will be from a contract. The token contract and denom are stored in two ways. First, the denom is used as the key and the value is the token contract. Second, the contract is used as the key, the value is the denom the token contract represents. 

| Key                                 | Value                                        | Type     | Encoding         |
|-------------------------------------|----------------------------------------------|----------|------------------|
| `[]byte{0xf3} + []byte(denom)` | Token contract address | `[]byte` | stored in byte format |

| Key                                 | Value                                        | Type     | Encoding         |
|-------------------------------------|----------------------------------------------|----------|------------------|
| `[]byte{0xf4} + []byte(tokenContract)` | Latest height a batch slashing occurred | `[]byte` | stored in byte format |

### LastEventNonce

The last observed event nonce. This is set when `TryAttestation()` is called. There is always only a single value held in this store.

| Key                                 | Value                                        | Type     | Encoding         |
|-------------------------------------|----------------------------------------------|----------|------------------|
| `[]byte{0xf2}` | Last observed event nonce| `uint64` | Big endian encoded |

### LastObservedEthereumHeight

This is the last observed height on ethereum. There will always only be a single value stored in this store.

| Key                                 | Value                                        | Type     | Encoding         |
|-------------------------------------|----------------------------------------------|----------|------------------|
| `[]byte{0xf9}` | Last observed Ethereum Height| `uint64` | Protobuf encoded |


### Attestation

```go
// Attestation is an aggregate of `claims` that eventually becomes `observed` by
// all orchestrators
// EVENT_NONCE:
// EventNonce a nonce provided by the peggy contract that is unique per event fired
// These event nonces must be relayed in order. This is a correctness issue,
// if relaying out of order transaction replay attacks become possible
// OBSERVED:
// Observed indicates that >67% of validators have attested to the event,
// and that the event should be executed by the peggy state machine
type Attestation struct {
	Observed bool       
	Votes    []string   
	Height   uint64     
	Claim    *types.Any 
}
```
| Key                                 | Value                                        | Type     | Encoding         |
|-------------------------------------|----------------------------------------------|----------|------------------|
| `[]byte{0x5} + evenNonce (big endian encoded) + []byte(claimHash)` | Attestation of occurred events/claims| `types.Attestation` | Protobuf encoded |
