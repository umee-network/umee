package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ModuleName is the name of the module
	ModuleName = "peggy"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the module name router key
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName
)

var (
	// EthAddressByValidatorKey indexes cosmos validator account addresses
	// i.e. cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn
	EthAddressByValidatorKey = []byte{0x1}

	// ValidatorByEthAddressKey indexes ethereum addresses
	// i.e. 0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B
	ValidatorByEthAddressKey = []byte{0xfb}

	// ValsetRequestKey indexes valset requests by nonce
	ValsetRequestKey = []byte{0x2}

	// ValsetConfirmKey indexes valset confirmations by nonce and the validator account address
	// i.e cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn
	ValsetConfirmKey = []byte{0x3}

	// OracleAttestationKey attestation details by nonce and validator address
	// i.e. cosmosvaloper1ahx7f8wyertuus9r20284ej0asrs085case3kn
	// An attestation can be thought of as the 'event to be executed' while
	// the Claims are an individual validator saying that they saw an event
	// occur the Attestation is 'the event' that multiple claims vote on and
	// eventually executes
	OracleAttestationKey = []byte{0x5}

	// OutgoingTXPoolKey indexes the last nonce for the outgoing tx pool
	OutgoingTXPoolKey = []byte{0x6}

	// SecondIndexOutgoingTXFeeKey indexes fee amounts by token contract address
	SecondIndexOutgoingTXFeeKey = []byte{0x9}

	// OutgoingTXBatchKey indexes outgoing tx batches under a nonce and token address
	OutgoingTXBatchKey = []byte{0xa}

	// OutgoingTXBatchBlockKey indexes outgoing tx batches under a block height and token address
	OutgoingTXBatchBlockKey = []byte{0xb}

	// BatchConfirmKey indexes validator confirmations by token contract address
	BatchConfirmKey = []byte{0xe1}

	// LastEventNonceByValidatorKey indexes lateset event nonce by validator
	LastEventNonceByValidatorKey = []byte{0xe2}

	// LastEventByValidatorKey indexes lateset claim event by validator
	LastEventByValidatorKey = []byte{0xf1}

	// LastObservedEventNonceKey indexes the latest event nonce
	LastObservedEventNonceKey = []byte{0xf2}

	// SequenceKeyPrefix indexes different txids
	SequenceKeyPrefix = []byte{0x7}

	// KeyLastTXPoolID indexes the lastTxPoolID
	KeyLastTXPoolID = append(SequenceKeyPrefix, []byte("lastTxPoolId")...)

	// KeyLastOutgoingBatchID indexes the lastBatchID
	KeyLastOutgoingBatchID = append(SequenceKeyPrefix, []byte("lastBatchId")...)

	// KeyOrchestratorAddress indexes the validator keys for an orchestrator
	KeyOrchestratorAddress = []byte{0xe8}

	// LastObservedEthereumBlockHeightKey indexes the latest Ethereum block height
	LastObservedEthereumBlockHeightKey = []byte{0xf9}

	// DenomToERC20Key prefixes the index of Cosmos originated asset denoms to ERC20s
	DenomToERC20Key = []byte{0xf3}

	// ERC20ToDenomKey prefixes the index of Cosmos originated assets ERC20s to denoms
	ERC20ToDenomKey = []byte{0xf4}

	// LastSlashedValsetNonce indexes the latest slashed valset nonce
	LastSlashedValsetNonce = []byte{0xf5}

	// LatestValsetNonce indexes the latest valset nonce
	LatestValsetNonce = []byte{0xf6}

	// LastSlashedBatchBlock indexes the latest slashed batch block height
	LastSlashedBatchBlock = []byte{0xf7}

	// LastUnbondingBlockHeight indexes the last validator unbonding block height
	LastUnbondingBlockHeight = []byte{0xf8}

	// LastObservedValsetNonceKey indexes the latest observed valset nonce
	// HERE THERE BE DRAGONS, do not use this value as an up to date validator set
	// on Ethereum it will always lag significantly and may be totally wrong at some
	// times.
	LastObservedValsetKey = []byte{0xfa}

	// PastEthSignatureCheckpointKey indexes eth signature checkpoints that have existed
	PastEthSignatureCheckpointKey = []byte{0x1b}
)

// GetOrchestratorAddressKey returns the following key format
// prefix
// [0xe8][cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn]
func GetOrchestratorAddressKey(orc sdk.AccAddress) []byte {
	return append(KeyOrchestratorAddress, orc.Bytes()...)
}

// GetEthAddressByValidatorKey returns the following key format
// prefix              cosmos-validator
// [0x0][cosmosvaloper1ahx7f8wyertuus9r20284ej0asrs085case3kn]
func GetEthAddressByValidatorKey(validator sdk.ValAddress) []byte {
	return append(EthAddressByValidatorKey, validator.Bytes()...)
}

// GetValidatorByEthAddressKey returns the following key format
// prefix              cosmos-validator
// [0xf9][0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B]
func GetValidatorByEthAddressKey(ethAddress common.Address) []byte {
	return append(ValidatorByEthAddressKey, ethAddress.Bytes()...)
}

// GetValsetKey returns the following key format
// prefix    nonce
// [0x0][0 0 0 0 0 0 0 1]
func GetValsetKey(nonce uint64) []byte {
	return append(ValsetRequestKey, UInt64Bytes(nonce)...)
}

// GetValsetConfirmKey returns the following key format
// prefix   nonce                    validator-address
// [0x0][0 0 0 0 0 0 0 1][cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn]
// MARK finish-batches: this is where the key is created in the old (presumed working) code
func GetValsetConfirmKey(nonce uint64, validator sdk.AccAddress) []byte {
	return append(ValsetConfirmKey, append(UInt64Bytes(nonce), validator.Bytes()...)...)
}

// GetAttestationKey returns the following key format
// prefix     nonce                             claim-details-hash
// [0x5][0 0 0 0 0 0 0 1][fd1af8cec6c67fcf156f1b61fdf91ebc04d05484d007436e75342fc05bbff35a]
// An attestation is an event multiple people are voting on, this function needs the claim
// details because each Attestation is aggregating all claims of a specific event, lets say
// validator X and validator y where making different claims about the same event nonce
// Note that the claim hash does NOT include the claimer address and only identifies an event
func GetAttestationKey(eventNonce uint64, claimHash []byte) []byte {
	key := make([]byte, len(OracleAttestationKey)+len(UInt64Bytes(0))+len(claimHash))
	copy(key[0:], OracleAttestationKey)
	copy(key[len(OracleAttestationKey):], UInt64Bytes(eventNonce))
	copy(key[len(OracleAttestationKey)+len(UInt64Bytes(0)):], claimHash)
	return key
}

// GetAttestationKeyWithHash returns the following key format
// prefix     nonce                             claim-details-hash
// [0x5][0 0 0 0 0 0 0 1][fd1af8cec6c67fcf156f1b61fdf91ebc04d05484d007436e75342fc05bbff35a]
// An attestation is an event multiple people are voting on, this function needs the claim
// details because each Attestation is aggregating all claims of a specific event, lets say
// validator X and validator y where making different claims about the same event nonce
// Note that the claim hash does NOT include the claimer address and only identifies an event
func GetAttestationKeyWithHash(eventNonce uint64, claimHash []byte) []byte {
	key := make([]byte, len(OracleAttestationKey)+len(UInt64Bytes(0))+len(claimHash))
	copy(key[0:], OracleAttestationKey)
	copy(key[len(OracleAttestationKey):], UInt64Bytes(eventNonce))
	copy(key[len(OracleAttestationKey)+len(UInt64Bytes(0)):], claimHash)
	return key
}

// GetOutgoingTxPoolKey returns the following key format
// prefix     id
// [0x6][0 0 0 0 0 0 0 1]
func GetOutgoingTxPoolKey(id uint64) []byte {
	buf := make([]byte, 0, len(OutgoingTXPoolKey)+8)
	buf = append(buf, OutgoingTXPoolKey...)
	buf = append(buf, sdk.Uint64ToBigEndian(id)...)

	return buf
}

// GetOutgoingTxBatchKey returns the following key format
// prefix     nonce                     eth-contract-address
// [0xa][0 0 0 0 0 0 0 1][0xc783df8a850f42e7F7e57013759C285caa701eB6]
func GetOutgoingTxBatchKey(tokenContract common.Address, nonce uint64) []byte {
	buf := make([]byte, 0, len(OutgoingTXBatchKey)+8+ETHContractAddressLen)
	buf = append(buf, OutgoingTXBatchKey...)
	buf = append(buf, UInt64Bytes(nonce)...)
	buf = append(buf, tokenContract.Bytes()...)

	return buf
}

// GetOutgoingTxBatchBlockKey returns the following key format
// prefix     blockheight
// [0xb][0 0 0 0 2 1 4 3]
func GetOutgoingTxBatchBlockKey(block uint64) []byte {
	return append(OutgoingTXBatchBlockKey, UInt64Bytes(block)...)
}

// GetBatchConfirmKey returns the following key format
// prefix           eth-contract-address                BatchNonce                       Validator-address
// [0xe1][0xc783df8a850f42e7F7e57013759C285caa701eB6][0 0 0 0 0 0 0 1][cosmosvaloper1ahx7f8wyertuus9r20284ej0asrs085case3kn]
// TODO this should be a sdk.ValAddress
func GetBatchConfirmKey(tokenContract common.Address, batchNonce uint64, validator sdk.AccAddress) []byte {
	buf := make([]byte, 0, len(BatchConfirmKey)+ETHContractAddressLen+8+len(validator))
	buf = append(buf, BatchConfirmKey...)
	buf = append(buf, tokenContract.Bytes()...)
	buf = append(buf, UInt64Bytes(batchNonce)...)
	buf = append(buf, validator.Bytes()...)

	return buf
}

// GetFeeSecondIndexKey returns the following key format
// prefix            eth-contract-address            					fee_amount
// [0x9][0xc783df8a850f42e7F7e57013759C285caa701eB6][0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
func GetFeeSecondIndexKey(tokenContract common.Address, fee *ERC20Token) []byte {
	buf := make([]byte, 0, len(SecondIndexOutgoingTXFeeKey)+ETHContractAddressLen+32)
	buf = append(buf, SecondIndexOutgoingTXFeeKey...)
	buf = append(buf, tokenContract.Bytes()...)

	// sdk.BigInt represented as a zero-extended big-endian byte slice (32 bytes)
	amount := make([]byte, 32)
	amount = fee.Amount.BigInt().FillBytes(amount)
	buf = append(buf, amount...)

	return buf
}

// GetLastEventNonceByValidatorKey indexes lateset event nonce by validator
// GetLastEventNonceByValidatorKey returns the following key format
// prefix              cosmos-validator
// [0x0][cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn]
func GetLastEventNonceByValidatorKey(validator sdk.ValAddress) []byte {
	buf := make([]byte, 0, len(LastEventNonceByValidatorKey)+len(validator))
	buf = append(buf, LastEventNonceByValidatorKey...)
	buf = append(buf, validator.Bytes()...)

	return buf
}

// GetLastEventByValidatorKey indexes lateset event by validator
// GetLastEventByValidatorKey returns the following key format
// prefix              cosmos-validator
// [0x0][cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn]
func GetLastEventByValidatorKey(validator sdk.ValAddress) []byte {
	buf := make([]byte, 0, len(LastEventByValidatorKey)+len(validator))
	buf = append(buf, LastEventByValidatorKey...)
	buf = append(buf, validator.Bytes()...)

	return buf
}

func GetCosmosDenomToERC20Key(denom string) []byte {
	buf := make([]byte, 0, len(DenomToERC20Key)+len(denom))
	buf = append(buf, DenomToERC20Key...)
	buf = append(buf, denom...)

	return buf
}

func GetERC20ToCosmosDenomKey(tokenContract common.Address) []byte {
	buf := make([]byte, 0, len(ERC20ToDenomKey)+ETHContractAddressLen)
	buf = append(buf, ERC20ToDenomKey...)
	buf = append(buf, tokenContract.Bytes()...)

	return buf
}

// GetPastEthSignatureCheckpointKey returns the following key format
// prefix    checkpoint
// [0x0][ checkpoint bytes ]
func GetPastEthSignatureCheckpointKey(checkpoint common.Hash) []byte {
	return append(PastEthSignatureCheckpointKey, checkpoint[:]...)
}
