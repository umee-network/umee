package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutgoingTxBatchCheckpointGold1(t *testing.T) {
	senderAddr, err := sdk.AccAddressFromHex("527FBEE652609AB150F0AEE9D61A2F76CFC4A73E")
	require.NoError(t, err)
	var (
		erc20Addr = common.HexToAddress("0x835973768750b3ed2d5c3ef5adcd5edb44d12ad4")
	)

	src := OutgoingTxBatch{
		BatchNonce: 1,
		//
		BatchTimeout: 2111,
		Transactions: []*OutgoingTransferTx{
			{
				Id:          0x1,
				Sender:      senderAddr.String(),
				DestAddress: common.HexToAddress("0x9fc9c2dfba3b6cf204c37a5f690619772b926e39").Hex(),
				Erc20Token: &ERC20Token{
					Amount:   sdk.NewInt(0x1),
					Contract: erc20Addr.Hex(),
				},
				Erc20Fee: &ERC20Token{
					Amount:   sdk.NewInt(0x1),
					Contract: erc20Addr.Hex(),
				},
			},
		},
		TokenContract: erc20Addr.Hex(),
	}

	ourHash := src.GetCheckpoint("foo")

	// hash from bridge contract
	goldHash := common.HexToHash("0xa3a7ee0a363b8ad2514e7ee8f110d7449c0d88f3b0913c28c1751e6e0079a9b2")
	// The function used to compute the "gold hash" above is in /solidity/test/updateValsetAndSubmitBatch.ts
	// Be aware that every time that you run the above .ts file, it will use a different tokenContractAddress and thus compute
	// a different hash.
	assert.Equal(t, goldHash, ourHash)
}
