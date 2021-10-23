package peggy_test

import (
	"bytes"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/umee-network/umee/x/peggy"
	"github.com/umee-network/umee/x/peggy/testpeggy"
	"github.com/umee-network/umee/x/peggy/types"

	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
)

func TestHandleMsgSendToEth(t *testing.T) {
	var (
		userCosmosAddr, _           = sdk.AccAddressFromBech32("cosmos1990z7dqsvh8gthw9pa5sn4wuy2xrsd80mg5z6y")
		blockTime                   = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		blockHeight           int64 = 200
		denom                       = "peggy0xB5E9944950C97acab395a324716D186632789712"
		startingCoinAmount, _       = sdk.NewIntFromString("150000000000000000000") // 150 ETH worth, required to reach above u64 limit (which is about 18 ETH)
		sendAmount, _               = sdk.NewIntFromString("50000000000000000000")  // 50 ETH
		feeAmount, _                = sdk.NewIntFromString("5000000000000000000")   // 5 ETH
		startingCoins               = sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}
		sendingCoin                 = sdk.NewCoin(denom, sendAmount)
		feeCoin                     = sdk.NewCoin(denom, feeAmount)
		ethDestination              = "0x3c9289da00b02dC623d0D8D907619890301D26d4"
	)

	// we start by depositing some funds into the users balance to send
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	h := NewHandler(input.PeggyKeeper)
	input.BankKeeper.MintCoins(ctx, types.ModuleName, startingCoins)
	input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userCosmosAddr, startingCoins)
	balance1 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}, balance1)

	// send some coins
	msg := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err := h(ctx, msg)
	require.NoError(t, err)
	balance2 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount.Sub(sendAmount).Sub(feeAmount))}, balance2)

	// do the same thing again and make sure it works twice
	msg1 := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err1 := h(ctx, msg1)
	require.NoError(t, err1)
	balance3 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	finalAmount3 := startingCoinAmount.Sub(sendAmount).Sub(sendAmount).Sub(feeAmount).Sub(feeAmount)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, finalAmount3)}, balance3)

	// now we should be out of coins and error
	msg2 := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err2 := h(ctx, msg2)
	require.Error(t, err2)
	balance4 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, finalAmount3)}, balance4)
}

func TestMsgDepositClaimSingleValidator(t *testing.T) {
	var (
		myOrchestratorAddr sdk.AccAddress = make([]byte, v040auth.AddrLen)
		myCosmosAddr, _                   = sdk.AccAddressFromBech32("cosmos16ahjkfqxpp6lvfy9fpfnfjg39xr96qett0alj5")
		myValAddr                         = sdk.ValAddress(myOrchestratorAddr) // revisit when proper mapping is impl in keeper
		myNonce                           = uint64(1)
		anyETHAddr                        = "0xf9613b532673Cc223aBa451dFA8539B87e1F666D"
		tokenETHAddr                      = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		myBlockTime                       = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		amountA, _                        = sdk.NewIntFromString("50000000000000000000")  // 50 ETH
		amountB, _                        = sdk.NewIntFromString("100000000000000000000") // 100 ETH
	)
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	input.PeggyKeeper.StakingKeeper = testpeggy.NewStakingKeeperMock(myValAddr)
	input.PeggyKeeper.SetOrchestratorValidator(ctx, myValAddr, myOrchestratorAddr)
	h := NewHandler(input.PeggyKeeper)

	myErc20 := types.ERC20Token{
		Amount:   amountA,
		Contract: tokenETHAddr,
	}

	ethClaim := types.MsgDepositClaim{
		EventNonce:     myNonce,
		TokenContract:  myErc20.Contract,
		Amount:         myErc20.Amount,
		EthereumSender: anyETHAddr,
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   myOrchestratorAddr.String(),
	}

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err := h(ctx, &ethClaim)
	NewBlockHandler(input.PeggyKeeper).EndBlocker(ctx)
	require.NoError(t, err)

	// and attestation persisted
	a := input.PeggyKeeper.GetAttestation(ctx, myNonce, ethClaim.ClaimHash())
	require.NotNil(t, a)
	// and vouchers added to the account
	balance := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("peggy0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", amountA)}, balance)

	// Test to reject duplicate deposit
	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = h(ctx, &ethClaim)
	NewBlockHandler(input.PeggyKeeper).EndBlocker(ctx)
	// then
	require.Error(t, err)
	balance = input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("peggy0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", amountA)}, balance)

	// Test to reject skipped nonce
	ethClaim = types.MsgDepositClaim{
		EventNonce:     uint64(3),
		TokenContract:  tokenETHAddr,
		Amount:         amountA,
		EthereumSender: anyETHAddr,
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   myOrchestratorAddr.String(),
	}

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = h(ctx, &ethClaim)
	NewBlockHandler(input.PeggyKeeper).EndBlocker(ctx)
	// then
	require.Error(t, err)
	balance = input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("peggy0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", amountA)}, balance)

	// Test to finally accept consecutive nonce
	ethClaim = types.MsgDepositClaim{
		EventNonce:     uint64(2),
		Amount:         amountA,
		TokenContract:  tokenETHAddr,
		EthereumSender: anyETHAddr,
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   myOrchestratorAddr.String(),
	}

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = h(ctx, &ethClaim)
	NewBlockHandler(input.PeggyKeeper).EndBlocker(ctx)

	// then
	require.NoError(t, err)
	balance = input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("peggy0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", amountB)}, balance)
}

func TestMsgDepositClaimsMultiValidator(t *testing.T) {
	var (
		orchestratorAddr1, _ = sdk.AccAddressFromBech32("cosmos1dg55rtevlfxh46w88yjpdd08sqhh5cc3xhkcej")
		orchestratorAddr2, _ = sdk.AccAddressFromBech32("cosmos164knshrzuuurf05qxf3q5ewpfnwzl4gj4m4dfy")
		orchestratorAddr3, _ = sdk.AccAddressFromBech32("cosmos193fw83ynn76328pty4yl7473vg9x86alq2cft7")
		myCosmosAddr, _      = sdk.AccAddressFromBech32("cosmos16ahjkfqxpp6lvfy9fpfnfjg39xr96qett0alj5")
		valAddr1             = sdk.ValAddress(orchestratorAddr1) // revisit when proper mapping is impl in keeper
		valAddr2             = sdk.ValAddress(orchestratorAddr2) // revisit when proper mapping is impl in keeper
		valAddr3             = sdk.ValAddress(orchestratorAddr3) // revisit when proper mapping is impl in keeper
		myNonce              = uint64(1)
		anyETHAddr           = common.HexToAddress("0xf9613b532673cc223aba451dfa8539b87e1f666d")
		tokenETHAddr         = common.HexToAddress("0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e")
		myBlockTime          = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
	)
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	input.PeggyKeeper.StakingKeeper = testpeggy.NewStakingKeeperMock(valAddr1, valAddr2, valAddr3)
	input.PeggyKeeper.SetOrchestratorValidator(ctx, valAddr1, orchestratorAddr1)
	input.PeggyKeeper.SetOrchestratorValidator(ctx, valAddr2, orchestratorAddr2)
	input.PeggyKeeper.SetOrchestratorValidator(ctx, valAddr3, orchestratorAddr3)
	h := NewHandler(input.PeggyKeeper)

	myErc20 := types.ERC20Token{
		Amount:   sdk.NewInt(12),
		Contract: tokenETHAddr.Hex(),
	}

	ethClaim1 := types.MsgDepositClaim{
		EventNonce:     myNonce,
		TokenContract:  myErc20.Contract,
		Amount:         myErc20.Amount,
		EthereumSender: anyETHAddr.Hex(),
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   orchestratorAddr1.String(),
	}
	ethClaim2 := types.MsgDepositClaim{
		EventNonce:     myNonce,
		TokenContract:  myErc20.Contract,
		Amount:         myErc20.Amount,
		EthereumSender: anyETHAddr.Hex(),
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   orchestratorAddr2.String(),
	}
	ethClaim3 := types.MsgDepositClaim{
		EventNonce:     myNonce,
		TokenContract:  myErc20.Contract,
		Amount:         myErc20.Amount,
		EthereumSender: anyETHAddr.Hex(),
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   orchestratorAddr3.String(),
	}

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err := h(ctx, &ethClaim1)
	NewBlockHandler(input.PeggyKeeper).EndBlocker(ctx)
	require.NoError(t, err)
	// and attestation persisted
	a1 := input.PeggyKeeper.GetAttestation(ctx, myNonce, ethClaim1.ClaimHash())
	require.NotNil(t, a1)
	// and vouchers not yet added to the account
	balance1 := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.NotEqual(t, sdk.Coins{sdk.NewInt64Coin("peggy0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", 12)}, balance1)

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = h(ctx, &ethClaim2)
	NewBlockHandler(input.PeggyKeeper).EndBlocker(ctx)
	require.NoError(t, err)

	// and attestation persisted
	a2 := input.PeggyKeeper.GetAttestation(ctx, myNonce, ethClaim1.ClaimHash())
	require.NotNil(t, a2)
	// and vouchers now added to the account
	balance2 := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewInt64Coin("peggy0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", 12)}, balance2)

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = h(ctx, &ethClaim3)
	NewBlockHandler(input.PeggyKeeper).EndBlocker(ctx)
	require.NoError(t, err)

	// and attestation persisted
	a3 := input.PeggyKeeper.GetAttestation(ctx, myNonce, ethClaim1.ClaimHash())
	require.NotNil(t, a3)
	// and no additional added to the account
	balance3 := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewInt64Coin("peggy0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", 12)}, balance3)
}

func TestMsgSetOrchestratorAddresses(t *testing.T) {
	var (
		ethAddress                             = common.HexToAddress("0xb462864E395d88d6bc7C5dd5F3F5eb4cc2599255")
		cosmosAddress           sdk.AccAddress = bytes.Repeat([]byte{0x1}, v040auth.AddrLen)
		validatorAccountAddress sdk.AccAddress = bytes.Repeat([]byte{0x2}, v040auth.AddrLen)
		blockTime                              = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		blockHeight             int64          = 200
	)
	validatorAddr := sdk.ValAddress(validatorAccountAddress)
	input := testpeggy.CreateTestEnv(t)
	input.PeggyKeeper.StakingKeeper = testpeggy.NewStakingKeeperMock(validatorAddr)
	ctx := input.Context
	h := NewHandler(input.PeggyKeeper)
	ctx = ctx.WithBlockTime(blockTime)

	msg := types.NewMsgSetOrchestratorAddress(validatorAccountAddress, cosmosAddress, ethAddress)
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err := h(ctx, msg)
	require.NoError(t, err)

	getEthAddress, _ := input.PeggyKeeper.GetEthAddressByValidator(ctx, validatorAddr)
	assert.Equal(t, ethAddress, getEthAddress)

	getOrchestratorValidator, _ := input.PeggyKeeper.GetOrchestratorValidator(ctx, cosmosAddress)
	assert.Equal(t, validatorAddr, getOrchestratorValidator)
}
