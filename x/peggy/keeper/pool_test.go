package keeper_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/x/peggy/testpeggy"
	"github.com/umee-network/umee/x/peggy/types"
)

func TestAddToOutgoingPool(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	var (
		mySender, _         = sdk.AccAddressFromBech32("cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn")
		myReceiver          = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr = common.HexToAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5")
	)
	// mint some voucher first
	allVouchers := sdk.Coins{types.NewERC20Token(99999, myTokenContractAddr).PeggyCoin()}
	err := input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)
	fmt.Println("Error", err)
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	input.PeggyKeeper.SetLastOutgoingPoolID(ctx, uint64(0))
	// when
	for i, v := range []uint64{2, 3, 2, 1} {
		amount := types.NewERC20Token(uint64(i+100), myTokenContractAddr).PeggyCoin()
		fee := types.NewERC20Token(v, myTokenContractAddr).PeggyCoin()
		r, err := input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
	}
	// then
	var got []*types.OutgoingTransferTx
	input.PeggyKeeper.IterateOutgoingPoolByFee(ctx, myTokenContractAddr, func(_ uint64, tx *types.OutgoingTransferTx) bool {
		got = append(got, tx)
		return false
	})
	exp := []*types.OutgoingTransferTx{
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
	assert.Equal(t, exp, got)
}

func TestTotalBatchFeeInPool(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context

	// token1
	var (
		mySender, _     = sdk.AccAddressFromBech32("cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn")
		myReceiver      = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContract = common.HexToAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5")
	)
	// mint some voucher first
	allVouchers := sdk.Coins{types.NewERC20Token(99999, myTokenContract).PeggyCoin()}
	err := input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// create outgoing pool
	for i, v := range []uint64{2, 3, 2, 1} {
		amount := types.NewERC20Token(uint64(i+100), myTokenContract).PeggyCoin()
		fee := types.NewERC20Token(v, myTokenContract).PeggyCoin()
		r, err2 := input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err2)
		t.Logf("___ response: %#v", r)
	}

	// token 2 - Only top 100
	var (
		myToken2ContractAddr = common.HexToAddress("0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0")
	)
	// mint some voucher first
	allVouchers = sdk.Coins{types.NewERC20Token(18446744073709551615, myToken2ContractAddr).PeggyCoin()}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// Add

	// create outgoing pool
	for i := 0; i < 110; i++ {
		amount := types.NewERC20Token(uint64(i+100), myToken2ContractAddr).PeggyCoin()
		fee := types.NewERC20Token(uint64(5), myToken2ContractAddr).PeggyCoin()
		r, err := input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
	}

	batchFees := input.PeggyKeeper.GetAllBatchFees(ctx)
	/*
		tokenFeeMap should be
		map[0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5:8 0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0:500]
		**/
	assert.Equal(t, batchFees[0].TotalFees.BigInt(), big.NewInt(int64(8)))
	assert.Equal(t, batchFees[1].TotalFees.BigInt(), big.NewInt(int64(500)))

}

func TestLastOutgoingBatchID(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context

	// initially it should be zero
	batchID := input.PeggyKeeper.GetLastOutgoingBatchID(ctx)
	assert.Equal(t, uint64(0), batchID)
	// set
	inputBatchID := uint64(100)
	input.PeggyKeeper.SetLastOutgoingBatchID(ctx, inputBatchID)

	// get
	batchID = input.PeggyKeeper.GetLastOutgoingBatchID(ctx)
	assert.Equal(t, inputBatchID, batchID)

	// Auto increment
	input.PeggyKeeper.AutoIncrementID(ctx, types.KeyLastOutgoingBatchID)
	batchID = input.PeggyKeeper.GetLastOutgoingBatchID(ctx)
	assert.Equal(t, inputBatchID+1, batchID)
}
