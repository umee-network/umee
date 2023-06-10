package keeper

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

func call_eth() {
	client, err := ethclient.Dial("https://opt-goerli.g.alchemy.com/v2/euAaBF09KINxu0q4nEtfEVEtrK1BmotU")
	if err != nil {
		log.Fatal(err)
	}

	//client.CallContract()
}
