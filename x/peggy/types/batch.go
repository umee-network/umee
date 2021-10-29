package types

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// GetCheckpoint gets the checkpoint signature from the given outgoing tx batch
func (b OutgoingTxBatch) GetCheckpoint(peggyIDstring string) common.Hash {

	abi, err := abi.JSON(strings.NewReader(OutgoingBatchTxCheckpointABIJSON))
	if err != nil {
		panic("Bad ABI constant!")
	}

	// the contract argument is not a arbitrary length array but a fixed length 32 byte
	// array, therefore we have to utf8 encode the string (the default in this case) and
	// then copy the variable length encoded data into a fixed length array. This function
	// will panic if peggyId is too long to fit in 32 bytes
	peggyID, err := strToFixByteArray(peggyIDstring)
	if err != nil {
		panic(err)
	}

	// Create the methodName argument which salts the signature
	methodNameBytes := []uint8("transactionBatch")
	var batchMethodName [32]uint8
	copy(batchMethodName[:], methodNameBytes[:])

	// Run through the elements of the batch and serialize them
	txAmounts := make([]*big.Int, len(b.Transactions))
	txDestinations := make([]common.Address, len(b.Transactions))
	txFees := make([]*big.Int, len(b.Transactions))
	for i, tx := range b.Transactions {
		txAmounts[i] = tx.Erc20Token.Amount.BigInt()
		txDestinations[i] = common.HexToAddress(tx.DestAddress)
		txFees[i] = tx.Erc20Fee.Amount.BigInt()
	}

	// the methodName needs to be the same as the 'name' above in the checkpointAbiJson
	// but other than that it's a constant that has no impact on the output. This is because
	// it gets encoded as a function name which we must then discard.
	abiEncodedBatch, err := abi.Pack("submitBatch",
		peggyID,
		batchMethodName,
		txAmounts,
		txDestinations,
		txFees,
		big.NewInt(int64(b.BatchNonce)),
		common.HexToAddress(b.TokenContract),
		big.NewInt(int64(b.BatchTimeout)),
	)

	// this should never happen outside of test since any case that could crash on encoding
	// should be filtered above.
	if err != nil {
		panic(fmt.Sprintf("Error packing checkpoint! %s/n", err))
	}

	hash := crypto.Keccak256Hash(abiEncodedBatch[4:])
	return hash
}
