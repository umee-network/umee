<!--
order: 7
title: Events
-->

# Events

The peggy module emits the following events:

## EndBlocker

### EventAttestationObserved
| Type        | Attribute Key    | Attribute Value    |
|-------------|------------------|--------------------|
| observation | module           | peggy              |
| observation | attestation_type | {attestation_type} |
| observation | bridge_contract  | {bridge_contract}  |
| observation | bridge_chain_id  | {bridge_chain_id}  |
| observation | attestation_id   | {attestation_id}   |
| observation | nonce            | {nonce}            |
  
## Handler

### EventSetOrchestratorAddresses

| Type    | Attribute Key        | Attribute Value    |
|---------|----------------------|--------------------|
| message | module               | peggy     |
| message | set_operator_address | {operator_address} |

### EventSendToEth

| Type    | Attribute Key  | Attribute Value |
|---------|----------------|-----------------|
| message | module         | peggy     |
| message | outgoing_tx_id | {tx_id}         |


### EventBridgeWithdrawalReceived
| Type                | Attribute Key   | Attribute Value   |
|---------------------|-----------------|-------------------|
| withdrawal_received | module          | peggy             |
| withdrawal_received | bridge_contract | {bridge_contract} |
| withdrawal_received | bridge_chain_id | {bridge_chain_id} |
| withdrawal_received | outgoing_tx_id  | {outgoing_tx_id}  |
| withdrawal_received | nonce           | {nonce}           |

### EventBridgeWithdrawCanceled
| Type                 | Attribute Key   | Attribute Value   |
|----------------------|-----------------|-------------------|
| withdrawal_cancelled | module          | peggy             |
| withdrawal_cancelled | bridge_contract | {bridge_contract} |
| withdrawal_cancelled | bridge_chain_id | {bridge_chain_id} |


### EventOutgoingBatch

| Type           | Attribute Key      | Attribute Value   |
|----------------|--------------------|-------------------|
| outgoing_batch | module             | peggy             |
| outgoing_batch | bridge_contract    | {bridge_contract} |
| outgoing_batch | bridge_chain_id    | {bridge_chain_id} |
| outgoing_batch | outgoing_batch_id  | {outgoing_batch_id}|
| outgoing_batch | nonce              | {nonce}           |

### EventOutgoingBatchCanceled
| Type                     | Attribute Key   | Attribute Value   |
|--------------------------|-----------------|-------------------|
| outgoing_batch_cancelled | module          | peggy             |
| outgoing_batch_cancelled | bridge_contract | {bridge_contract} |
| outgoing_batch_cancelled | bridge_chain_id | {bridge_chain_id} |
| outgoing_batch_cancelled | outgoing_batch_id  | {outgoing_batch_id}  |
| outgoing_batch_cancelled | nonce           | {nonce}           |

### EventValsetConfirm

| Type    | Attribute Key        | Attribute Value    |
|---------|----------------------|--------------------|
| message | module               | peggy     |
| message | valset_confirm_key | {valset_confirm_key} |


### EventConfirmBatch

| Type    | Attribute Key     | Attribute Value     |
|---------|-------------------|---------------------|
| message | module            | peggy       |
| message | batch_confirm_key | {batch_confirm_key} |

### EventDepositClaim

| Type    | Attribute Key  | Attribute Value   |
|---------|----------------|-------------------|
| message | module         | peggy     |
| message | attestation_id | {attestation_key} |


### EventWithdrawClaim

| Type    | Attribute Key  | Attribute Value   |
|---------|----------------|-------------------|
| message | module         | peggy    |
| message | attestation_id | {attestation_key} |

### EventERC20DeployedClaim
| Type    | Attribute Key  | Attribute Value      |
|---------|----------------|----------------------|
| message | module         | peggy |
| message | attestation_id | {attestation_key}    |

### EventValsetUpdateClaim
| Type    | Attribute Key  | Attribute Value      |
|---------|----------------|----------------------|
| message | module         | peggy |
| message | attestation_id | {attestation_key}    |

