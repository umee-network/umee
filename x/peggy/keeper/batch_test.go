package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/x/peggy/testpeggy"
	"github.com/umee-network/umee/x/peggy/types"
)

func TestBatches(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	var (
		now                 = time.Now().UTC()
		mySender, _         = sdk.AccAddressFromBech32("umee1dkfhxs87adz9ll6jfr0jr5jet6u8tjaqx4z8rg")
		myReceiver          = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr = common.HexToAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5") // Pickle
		allVouchers         = sdk.NewCoins(
			types.NewERC20Token(99999, myTokenContractAddr).PeggyCoin(),
		)
	)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)

	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	input.PeggyKeeper.SetLastOutgoingPoolID(ctx, uint64(0))
	input.PeggyKeeper.SetLastOutgoingBatchID(ctx, uint64(0))
	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	for i, v := range []uint64{2, 3, 2, 1} {
		amount := types.NewERC20Token(uint64(i+100), myTokenContractAddr).PeggyCoin()
		fee := types.NewERC20Token(v, myTokenContractAddr).PeggyCoin()
		_, err := input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err)
	}

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	firstBatch, err := input.PeggyKeeper.BuildOutgoingTXBatch(ctx, myTokenContractAddr, 2)
	require.NoError(t, err)

	// then batch is persisted
	gotFirstBatch := input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(firstBatch.TokenContract), firstBatch.BatchNonce)
	require.NotNil(t, gotFirstBatch)

	expFirstBatch := &types.OutgoingTxBatch{
		BatchNonce: 1,
		Transactions: []*types.OutgoingTransferTx{
			{
				Id:          2,
				Erc20Fee:    types.NewERC20Token(3, myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver.Hex(),
				Erc20Token:  types.NewERC20Token(101, myTokenContractAddr),
			},
			{
				Id:          1,
				Erc20Fee:    types.NewERC20Token(2, myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver.Hex(),
				Erc20Token:  types.NewERC20Token(100, myTokenContractAddr),
			},
		},
		TokenContract: myTokenContractAddr.Hex(),
		Block:         1234567,
	}
	assert.Equal(t, expFirstBatch, gotFirstBatch)

	// and verify remaining available Tx in the pool
	var gotUnbatchedTx []*types.OutgoingTransferTx
	input.PeggyKeeper.IterateOutgoingPoolByFee(ctx, myTokenContractAddr, func(_ uint64, tx *types.OutgoingTransferTx) bool {
		gotUnbatchedTx = append(gotUnbatchedTx, tx)
		return false
	})
	expUnbatchedTx := []*types.OutgoingTransferTx{
		{
			Id:          3,
			Erc20Fee:    types.NewERC20Token(2, myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewERC20Token(102, myTokenContractAddr),
		},
		{
			Id:          4,
			Erc20Fee:    types.NewERC20Token(1, myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewERC20Token(103, myTokenContractAddr),
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)

	// CREATE SECOND, MORE PROFITABLE BATCH
	// ====================================

	// add some more TX to the pool to create a more profitable batch
	for i, v := range []uint64{4, 5} {

		amount := types.NewERC20Token(uint64(i+100), myTokenContractAddr).PeggyCoin()
		fee := types.NewERC20Token(v, myTokenContractAddr).PeggyCoin()
		_, err = input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err)
	}

	// create the more profitable batch
	ctx = ctx.WithBlockTime(now)
	// tx batch size is 2, so that some of them stay behind
	secondBatch, err := input.PeggyKeeper.BuildOutgoingTXBatch(ctx, myTokenContractAddr, 2)
	require.NoError(t, err)

	// check that the more profitable batch has the right txs in it
	expSecondBatch := &types.OutgoingTxBatch{
		BatchNonce: 2,
		Transactions: []*types.OutgoingTransferTx{
			{
				Id:          6,
				Erc20Fee:    types.NewERC20Token(5, myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver.Hex(),
				Erc20Token:  types.NewERC20Token(101, myTokenContractAddr),
			},
			{
				Id:          5,
				Erc20Fee:    types.NewERC20Token(4, myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver.Hex(),
				Erc20Token:  types.NewERC20Token(100, myTokenContractAddr),
			},
		},
		TokenContract: myTokenContractAddr.Hex(),
		Block:         1234567,
	}

	assert.Equal(t, expSecondBatch, secondBatch)

	// EXECUTE THE MORE PROFITABLE BATCH
	// =================================

	// Execute the batch
	input.PeggyKeeper.OutgoingTxBatchExecuted(ctx, common.HexToAddress(secondBatch.TokenContract), secondBatch.BatchNonce)
	// check batch has been deleted
	gotSecondBatch := input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(secondBatch.TokenContract), secondBatch.BatchNonce)
	require.Nil(t, gotSecondBatch)

	// check that txs from first batch have been freed
	gotUnbatchedTx = nil
	input.PeggyKeeper.IterateOutgoingPoolByFee(ctx, myTokenContractAddr, func(_ uint64, tx *types.OutgoingTransferTx) bool {
		gotUnbatchedTx = append(gotUnbatchedTx, tx)
		return false
	})
	expUnbatchedTx = []*types.OutgoingTransferTx{
		{
			Id:          2,
			Erc20Fee:    types.NewERC20Token(3, myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewERC20Token(101, myTokenContractAddr),
		},
		{
			Id:          1,
			Erc20Fee:    types.NewERC20Token(2, myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewERC20Token(100, myTokenContractAddr),
		},
		{
			Id:          3,
			Erc20Fee:    types.NewERC20Token(2, myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewERC20Token(102, myTokenContractAddr),
		},
		{
			Id:          4,
			Erc20Fee:    types.NewERC20Token(1, myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewERC20Token(103, myTokenContractAddr),
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)
}

// tests that batches work with large token amounts, mostly a duplicate of the above
// tests but using much bigger numbers
func TestBatchesFullCoins(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	var (
		now                 = time.Now().UTC()
		mySender, _         = sdk.AccAddressFromBech32("umee1dkfhxs87adz9ll6jfr0jr5jet6u8tjaqx4z8rg")
		myReceiver          = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr = common.HexToAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5") // Pickle
		totalCoins, _       = sdk.NewIntFromString("1500000000000000000000")                    // 1,500 ETH worth
		oneEth, _           = sdk.NewIntFromString("1000000000000000000")
		allVouchers         = sdk.NewCoins(
			types.NewSDKIntERC20Token(totalCoins, myTokenContractAddr).PeggyCoin(),
		)
	)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	input.PeggyKeeper.SetLastOutgoingPoolID(ctx, uint64(0))
	input.PeggyKeeper.SetLastOutgoingBatchID(ctx, uint64(0))
	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	for _, v := range []uint64{20, 300, 25, 10} {
		vAsSDKInt := sdk.NewIntFromUint64(v)
		amount := types.NewSDKIntERC20Token(oneEth.Mul(vAsSDKInt), myTokenContractAddr).PeggyCoin()
		fee := types.NewSDKIntERC20Token(oneEth.Mul(vAsSDKInt), myTokenContractAddr).PeggyCoin()
		_, err := input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err)
	}

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	firstBatch, err := input.PeggyKeeper.BuildOutgoingTXBatch(ctx, myTokenContractAddr, 2)
	require.NoError(t, err)

	// then batch is persisted
	gotFirstBatch := input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(firstBatch.TokenContract), firstBatch.BatchNonce)
	require.NotNil(t, gotFirstBatch)

	expFirstBatch := &types.OutgoingTxBatch{
		BatchNonce: 1,
		Transactions: []*types.OutgoingTransferTx{
			{
				Id:          2,
				Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(300)), myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver.Hex(),
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(300)), myTokenContractAddr),
			},
			{
				Id:          3,
				Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(25)), myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver.Hex(),
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(25)), myTokenContractAddr),
			},
		},
		TokenContract: myTokenContractAddr.Hex(),
		Block:         1234567,
	}
	assert.Equal(t, expFirstBatch, gotFirstBatch)

	// and verify remaining available Tx in the pool
	var gotUnbatchedTx []*types.OutgoingTransferTx
	input.PeggyKeeper.IterateOutgoingPoolByFee(ctx, myTokenContractAddr, func(_ uint64, tx *types.OutgoingTransferTx) bool {
		gotUnbatchedTx = append(gotUnbatchedTx, tx)
		return false
	})
	expUnbatchedTx := []*types.OutgoingTransferTx{
		{
			Id:          1,
			Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(20)), myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(20)), myTokenContractAddr),
		},
		{
			Id:          4,
			Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(10)), myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(10)), myTokenContractAddr),
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)

	// CREATE SECOND, MORE PROFITABLE BATCH
	// ====================================

	// add some more TX to the pool to create a more profitable batch
	for _, v := range []uint64{4, 5} {
		vAsSDKInt := sdk.NewIntFromUint64(v)
		amount := types.NewSDKIntERC20Token(oneEth.Mul(vAsSDKInt), myTokenContractAddr).PeggyCoin()
		fee := types.NewSDKIntERC20Token(oneEth.Mul(vAsSDKInt), myTokenContractAddr).PeggyCoin()
		_, err := input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err)
	}

	// create the more profitable batch
	ctx = ctx.WithBlockTime(now)
	// tx batch size is 2, so that some of them stay behind
	secondBatch, err := input.PeggyKeeper.BuildOutgoingTXBatch(ctx, myTokenContractAddr, 2)
	require.NoError(t, err)

	// check that the more profitable batch has the right txs in it
	expSecondBatch := &types.OutgoingTxBatch{
		BatchNonce: 2,
		Transactions: []*types.OutgoingTransferTx{
			{
				Id:          1,
				Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(20)), myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver.Hex(),
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(20)), myTokenContractAddr),
			},
			{
				Id:          4,
				Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(10)), myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver.Hex(),
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(10)), myTokenContractAddr),
			},
		},
		TokenContract: myTokenContractAddr.Hex(),
		Block:         1234567,
	}

	assert.Equal(t, expSecondBatch, secondBatch)

	// EXECUTE THE MORE PROFITABLE BATCH
	// =================================

	// Execute the batch
	input.PeggyKeeper.OutgoingTxBatchExecuted(ctx, common.HexToAddress(secondBatch.TokenContract), secondBatch.BatchNonce)
	// check batch has been deleted
	gotSecondBatch := input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(secondBatch.TokenContract), secondBatch.BatchNonce)
	require.Nil(t, gotSecondBatch)

	// check that txs from first batch have been freed
	gotUnbatchedTx = nil
	input.PeggyKeeper.IterateOutgoingPoolByFee(ctx, myTokenContractAddr, func(_ uint64, tx *types.OutgoingTransferTx) bool {
		gotUnbatchedTx = append(gotUnbatchedTx, tx)
		return false
	})
	expUnbatchedTx = []*types.OutgoingTransferTx{
		{
			Id:          2,
			Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(300)), myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(300)), myTokenContractAddr),
		},
		{
			Id:          3,
			Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(25)), myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(25)), myTokenContractAddr),
		},
		{
			Id:          6,
			Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(5)), myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(5)), myTokenContractAddr),
		},
		{
			Id:          5,
			Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(4)), myTokenContractAddr),
			Sender:      mySender.String(),
			DestAddress: myReceiver.Hex(),
			Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(4)), myTokenContractAddr),
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)
}

func TestPoolTxRefund(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	var (
		now                 = time.Now().UTC()
		mySender, _         = sdk.AccAddressFromBech32("umee1dkfhxs87adz9ll6jfr0jr5jet6u8tjaqx4z8rg")
		myReceiver          = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr = common.HexToAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5") // Pickle
		allVouchers         = sdk.NewCoins(
			types.NewERC20Token(414, myTokenContractAddr).PeggyCoin(),
		)
		myDenom = types.NewERC20Token(1, myTokenContractAddr).PeggyCoin().Denom
	)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	input.PeggyKeeper.SetLastOutgoingPoolID(ctx, uint64(0))
	input.PeggyKeeper.SetLastOutgoingBatchID(ctx, uint64(0))
	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	for i, v := range []uint64{2, 3, 2, 1} {
		amount := types.NewERC20Token(uint64(i+100), myTokenContractAddr).PeggyCoin()
		fee := types.NewERC20Token(v, myTokenContractAddr).PeggyCoin()
		_, err := input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err)
	}

	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	_, err := input.PeggyKeeper.BuildOutgoingTXBatch(ctx, myTokenContractAddr, 2)
	require.NoError(t, err)

	// try to refund a tx that's in a batch
	err1 := input.PeggyKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 1, mySender)
	require.Error(t, err1)

	// try to refund a tx that's in the pool
	err2 := input.PeggyKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 4, mySender)
	require.NoError(t, err2)

	// make sure refund was issued
	balances := input.BankKeeper.GetAllBalances(ctx, mySender)
	require.Equal(t, sdk.NewInt(104), balances.AmountOf(myDenom))
}

// TestManyBatches handles test cases around batch execution, specifically executing multiple batches
// out of sequential order, which is exactly what happens on the
func TestManyBatches(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	var (
		now                = time.Now().UTC()
		mySender, _        = sdk.AccAddressFromBech32("umee1dkfhxs87adz9ll6jfr0jr5jet6u8tjaqx4z8rg")
		myReceiver         = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		tokenContractAddr1 = common.HexToAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5")
		tokenContractAddr2 = common.HexToAddress("0xF815240800ddf3E0be80e0d848B13ecaa504BF37")
		tokenContractAddr3 = common.HexToAddress("0xd086dDA7BccEB70e35064f540d07E4baED142cB3")
		tokenContractAddr4 = common.HexToAddress("0x384981B9d133701c4bD445F77bF61C3d80e79D46")
		totalCoins, _      = sdk.NewIntFromString("1500000000000000000000000")
		oneEth, _          = sdk.NewIntFromString("1000000000000000000")
		allVouchers        = sdk.NewCoins(
			types.NewSDKIntERC20Token(totalCoins, tokenContractAddr1).PeggyCoin(),
			types.NewSDKIntERC20Token(totalCoins, tokenContractAddr2).PeggyCoin(),
			types.NewSDKIntERC20Token(totalCoins, tokenContractAddr3).PeggyCoin(),
			types.NewSDKIntERC20Token(totalCoins, tokenContractAddr4).PeggyCoin(),
		)
	)

	// mint vouchers first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	tokens := [4]common.Address{tokenContractAddr1, tokenContractAddr2, tokenContractAddr3, tokenContractAddr4}

	for _, contract := range tokens {
		for v := 1; v < 500; v++ {
			vAsSDKInt := sdk.NewIntFromUint64(uint64(v))
			amount := types.NewSDKIntERC20Token(oneEth.Mul(vAsSDKInt), contract).PeggyCoin()
			fee := types.NewSDKIntERC20Token(oneEth.Mul(vAsSDKInt), contract).PeggyCoin()
			_, err := input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
			require.NoError(t, err)
		}
	}

	// when
	ctx = ctx.WithBlockTime(now)

	var batches []types.OutgoingTxBatch
	for _, contract := range tokens {
		for v := 1; v < 5; v++ {
			batch, err := input.PeggyKeeper.BuildOutgoingTXBatch(ctx, contract, 100)
			batches = append(batches, *batch)
			require.NoError(t, err)
		}
	}
	for _, batch := range batches {
		// then batch is persisted
		gotBatch := input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(batch.TokenContract), batch.BatchNonce)
		require.NotNil(t, gotBatch)
	}

	// EXECUTE BOTH BATCHES
	// =================================

	// shuffle batches to simulate out of order execution on Ethereum
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(batches), func(i, j int) { batches[i], batches[j] = batches[j], batches[i] })

	// Execute the batches, if there are any problems OutgoingTxBatchExecuted will panic
	for _, batch := range batches {
		gotBatch := input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(batch.TokenContract), batch.BatchNonce)
		// we may have already deleted some of the batches in this list by executing later ones
		if gotBatch != nil {
			input.PeggyKeeper.OutgoingTxBatchExecuted(ctx, common.HexToAddress(batch.TokenContract), batch.BatchNonce)
		}
	}
}
