package keeper

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func call_eth(recipient common.Address, amount *big.Int) {
	client, err := ethclient.Dial("https://opt-goerli.g.alchemy.com/v2/euAaBF09KINxu0q4nEtfEVEtrK1BmotU")
	if err != nil {
		log.Fatal(err)
	}

	// Instantiate the contract
	// TODO: Add facilitator addresss here
	address := common.HexToAddress("0x1f9840a85d5af5bf1d1762f925bdaddc4201f984")
	facilitator, err := NewReFiFacilitator(address, client)

	// TODO: Add bridge key here
	SK := "0x0000"
	sk := crypto.ToECDSAUnsafe(common.FromHex(SK)) // Sign the transaction
	// You could also create a TransactOpts object
	transactOpts := bind.NewKeyedTransactor(sk)
	ctx := context.Background()
	//facilitator.
	tx, err := facilitator.OnAxelarGmp(transactOpts, recipient, amount)
	receipt, err := bind.WaitMined(ctx, client, tx)
	if receipt.Status != types.ReceiptStatusSuccessful {
		panic("Call failed")
	}
}
