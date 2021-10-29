package peggy_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	. "github.com/umee-network/umee/x/peggy"
	"github.com/umee-network/umee/x/peggy/testpeggy"
	"github.com/umee-network/umee/x/peggy/types"
)

func TestValsetCreationIfNotAvailable(t *testing.T) {
	input, ctx := testpeggy.SetupFiveValChain(t)
	pk := input.PeggyKeeper

	// EndBlocker should set a new validator set if not available
	NewBlockHandler(pk).EndBlocker(ctx)
	require.NotNil(t, pk.GetValset(ctx, uint64(ctx.BlockHeight())))
	valsets := pk.GetValsets(ctx)
	require.True(t, len(valsets) == 1)

}

func TestValsetCreationUponUnbonding(t *testing.T) {
	input, ctx := testpeggy.SetupFiveValChain(t)
	pk := input.PeggyKeeper
	pk.SetValsetRequest(ctx)

	input.Context = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// begin unbonding
	sh := staking.NewHandler(input.StakingKeeper)
	undelegateMsg := testpeggy.NewTestMsgUnDelegateValidator(testpeggy.ValAddrs[0], testpeggy.StakingAmount)
	sh(input.Context, undelegateMsg)

	// Run the staking endblocker to ensure valset is set in state
	staking.EndBlocker(input.Context, input.StakingKeeper)
	NewBlockHandler(pk).EndBlocker(input.Context)

	assert.Equal(t, uint64(input.Context.BlockHeight()), pk.GetLatestValsetNonce(ctx))
}

func TestValsetSlashing_ValsetCreated_Before_ValidatorBonded(t *testing.T) {
	//	Don't slash validators if valset is created before he is bonded.

	input, ctx := testpeggy.SetupFiveValChain(t)
	pk := input.PeggyKeeper
	params := input.PeggyKeeper.GetParams(ctx)

	val := input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[0])
	vs := pk.GetCurrentValset(ctx)
	height := uint64(ctx.BlockHeight()) - (params.SignedValsetsWindow + 1)
	vs.Height = height
	vs.Nonce = height
	pk.StoreValsetUnsafe(ctx, vs)

	NewBlockHandler(pk).EndBlocker(ctx)

	// ensure that the  validator who is bonded after valset is created is not slashed
	val = input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[0])
	require.False(t, val.IsJailed())
}

func TestValsetSlashing_ValsetCreated_After_ValidatorBonded(t *testing.T) {
	//	Slashing Conditions for Bonded Validator

	input, ctx := testpeggy.SetupFiveValChain(t)
	pk := input.PeggyKeeper
	params := input.PeggyKeeper.GetParams(ctx)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(params.SignedValsetsWindow) + 2)
	val := input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[0])
	vs := pk.GetCurrentValset(ctx)
	height := uint64(ctx.BlockHeight()) - (params.SignedValsetsWindow + 1)
	vs.Height = height
	vs.Nonce = height
	pk.StoreValsetUnsafe(ctx, vs)

	for i, val := range testpeggy.AccAddrs {
		if i == 0 {
			// don't sign with first validator
			continue
		}

		valsetConfirm := types.NewMsgValsetConfirm(vs.Nonce, testpeggy.EthAddrs[i].String(), val, "dummysig")
		pk.SetValsetConfirm(ctx, valsetConfirm)
	}

	NewBlockHandler(pk).EndBlocker(ctx)

	// ensure that the  validator who is bonded before valset is created is slashed
	val = input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[0])
	require.True(t, val.IsJailed())

	// ensure that the  validator who attested the valset is not slashed.
	val = input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[1])
	require.False(t, val.IsJailed())

}

func TestValsetSlashing_UnbondingValidator_UnbondWindow_NotExpired(t *testing.T) {
	//	Slashing Conditions for Unbonding Validator

	//  Create 5 validators
	input, ctx := testpeggy.SetupFiveValChain(t)
	val := input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[0])
	fmt.Println("val1  tokens", val.GetTokens().ToDec())

	pk := input.PeggyKeeper
	params := input.PeggyKeeper.GetParams(ctx)

	// Define slashing variables
	validatorStartHeight := ctx.BlockHeight()                                                        // 0
	valsetRequestHeight := validatorStartHeight + 1                                                  // 1
	valUnbondingHeight := valsetRequestHeight + 1                                                    // 2
	valsetRequestSlashedAt := valsetRequestHeight + int64(params.SignedValsetsWindow)                // 11
	validatorUnbondingWindowExpiry := valUnbondingHeight + int64(params.UnbondSlashingValsetsWindow) // 17
	currentBlockHeight := valsetRequestSlashedAt + 1                                                 // 12

	assert.True(t, valsetRequestSlashedAt < currentBlockHeight)
	assert.True(t, valsetRequestHeight < validatorUnbondingWindowExpiry)

	// Create Valset request
	ctx = ctx.WithBlockHeight(valsetRequestHeight)
	vs := pk.GetCurrentValset(ctx)
	vs.Height = uint64(valsetRequestHeight)
	vs.Nonce = uint64(valsetRequestHeight)
	pk.StoreValsetUnsafe(ctx, vs)

	// Start Unbonding validators
	// Validator-1  Unbond slash window is not expired. if not attested, slash
	// Validator-2  Unbond slash window is not expired. if attested, don't slash
	input.Context = ctx.WithBlockHeight(valUnbondingHeight)
	sh := staking.NewHandler(input.StakingKeeper)
	undelegateMsg1 := testpeggy.NewTestMsgUnDelegateValidator(testpeggy.ValAddrs[0], testpeggy.StakingAmount)
	sh(input.Context, undelegateMsg1)
	undelegateMsg2 := testpeggy.NewTestMsgUnDelegateValidator(testpeggy.ValAddrs[1], testpeggy.StakingAmount)
	sh(input.Context, undelegateMsg2)

	for i, val := range testpeggy.AccAddrs {
		if i == 0 {
			// don't sign with first validator
			continue
		}
		valsetConfirm := types.NewMsgValsetConfirm(vs.Nonce, testpeggy.EthAddrs[i].String(), val, "dummysig")
		pk.SetValsetConfirm(ctx, valsetConfirm)
	}
	staking.EndBlocker(input.Context, input.StakingKeeper)

	ctx = ctx.WithBlockHeight(currentBlockHeight)
	NewBlockHandler(pk).EndBlocker(ctx)

	// Assertions
	val1 := input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[0])
	assert.True(t, val1.IsJailed())
	fmt.Println("val1  tokens", val1.GetTokens().ToDec())
	// check if tokens are slashed for val1.

	val2 := input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[1])
	assert.True(t, val2.IsJailed())
	fmt.Println("val2  tokens", val2.GetTokens().ToDec())
	// check if tokens shouldn't be slashed for val2.
}

func TestBatchSlashing(t *testing.T) {
	input, ctx := testpeggy.SetupFiveValChain(t)
	pk := input.PeggyKeeper
	params := pk.GetParams(ctx)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(params.SignedValsetsWindow) + 2)

	// First store a batch
	batch := &types.OutgoingTxBatch{
		BatchNonce:    1,
		Transactions:  []*types.OutgoingTransferTx{},
		TokenContract: testpeggy.TokenContractAddrs[0],
		Block:         uint64(ctx.BlockHeight() - int64(params.SignedBatchesWindow+1)),
	}
	pk.StoreBatchUnsafe(ctx, batch)

	for i, val := range testpeggy.AccAddrs {
		if i == 0 {
			// don't sign with first validator
			continue
		}
		if i == 1 {
			// don't sign with 2nd validator. set val bond height > batch block height
			validator := input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[i])
			valConsAddr, _ := validator.GetConsAddr()
			valSigningInfo := slashingtypes.ValidatorSigningInfo{StartHeight: int64(batch.Block + 1)}
			input.SlashingKeeper.SetValidatorSigningInfo(ctx, valConsAddr, valSigningInfo)
			continue
		}
		pk.SetBatchConfirm(ctx, &types.MsgConfirmBatch{
			Nonce:         batch.BatchNonce,
			TokenContract: testpeggy.TokenContractAddrs[0],
			EthSigner:     testpeggy.EthAddrs[i].String(),
			Orchestrator:  val.String(),
		})
	}

	NewBlockHandler(pk).EndBlocker(ctx)

	// ensure that the  validator is jailed and slashed
	val := input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[0])
	require.True(t, val.IsJailed())

	// ensure that the 2nd  validator is not jailed and slashed
	val2 := input.StakingKeeper.Validator(ctx, testpeggy.ValAddrs[1])
	require.False(t, val2.IsJailed())

	// Ensure that the last slashed valset nonce is set properly
	lastSlashedBatchBlock := input.PeggyKeeper.GetLastSlashedBatchBlock(ctx)
	assert.Equal(t, lastSlashedBatchBlock, batch.Block)

}

func TestValsetEmission(t *testing.T) {
	input, ctx := testpeggy.SetupFiveValChain(t)
	pk := input.PeggyKeeper

	// Store a validator set with a power change as the most recent validator set
	vs := pk.GetCurrentValset(ctx)
	vs.Nonce = vs.Nonce - 1
	delta := float64(types.BridgeValidators(vs.Members).TotalPower()) * 0.05
	vs.Members[0].Power = uint64(float64(vs.Members[0].Power) - delta/2)
	vs.Members[1].Power = uint64(float64(vs.Members[1].Power) + delta/2)
	pk.StoreValset(ctx, vs)

	// EndBlocker should set a new validator set
	NewBlockHandler(pk).EndBlocker(ctx)
	require.NotNil(t, pk.GetValset(ctx, uint64(ctx.BlockHeight())))
	valsets := pk.GetValsets(ctx)
	require.True(t, len(valsets) == 2)
}

func TestValsetSetting(t *testing.T) {
	input, ctx := testpeggy.SetupFiveValChain(t)
	pk := input.PeggyKeeper
	pk.SetValsetRequest(ctx)
	valsets := pk.GetValsets(ctx)
	require.True(t, len(valsets) == 1)
}

/// Test batch timeout
func TestBatchTimeout(t *testing.T) {
	input, ctx := testpeggy.SetupFiveValChain(t)
	pk := input.PeggyKeeper
	params := pk.GetParams(ctx)
	var (
		now                 = time.Now().UTC()
		mySender, _         = sdk.AccAddressFromBech32("cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn")
		myReceiver          = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr = common.HexToAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5") // Pickle
		allVouchers         = sdk.NewCoins(
			types.NewERC20Token(99999, myTokenContractAddr).PeggyCoin(),
		)
	)

	require.Greater(t, params.AverageBlockTime, uint64(0))
	require.Greater(t, params.AverageEthereumBlockTime, uint64(0))

	// mint some vouchers first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// add some TX to the pool
	for i, v := range []uint64{2, 3, 2, 1, 5, 6} {
		amount := types.NewERC20Token(uint64(i+100), myTokenContractAddr).PeggyCoin()
		fee := types.NewERC20Token(v, myTokenContractAddr).PeggyCoin()
		_, err := input.PeggyKeeper.AddToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err)
	}

	// when
	ctx = ctx.WithBlockTime(now)
	ctx = ctx.WithBlockHeight(250)

	// check that we can make a batch without first setting an ethereum block height
	b1, err1 := pk.BuildOutgoingTXBatch(ctx, myTokenContractAddr, 2)
	require.NoError(t, err1)
	require.Equal(t, b1.BatchTimeout, uint64(0))

	pk.SetLastObservedEthereumBlockHeight(ctx, 500)

	b2, err2 := pk.BuildOutgoingTXBatch(ctx, myTokenContractAddr, 2)
	require.NoError(t, err2)
	// this is exactly block 500 plus twelve hours
	require.Equal(t, b2.BatchTimeout, uint64(504))

	// make sure the batches got stored in the first place
	gotFirstBatch := input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(b1.TokenContract), b1.BatchNonce)
	require.NotNil(t, gotFirstBatch)
	gotSecondBatch := input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(b2.TokenContract), b2.BatchNonce)
	require.NotNil(t, gotSecondBatch)

	// when, way into the future
	ctx = ctx.WithBlockTime(now)
	ctx = ctx.WithBlockHeight(9)

	b3, err2 := pk.BuildOutgoingTXBatch(ctx, myTokenContractAddr, 2)
	require.NoError(t, err2)

	NewBlockHandler(pk).EndBlocker(ctx)

	// this had a timeout of zero should be deleted.
	gotFirstBatch = input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(b1.TokenContract), b1.BatchNonce)
	require.Nil(t, gotFirstBatch)
	// make sure the end blocker does not delete these, as the block height has not officially
	// been updated by a relay event
	gotSecondBatch = input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(b2.TokenContract), b2.BatchNonce)
	require.NotNil(t, gotSecondBatch)
	gotThirdBatch := input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(b3.TokenContract), b3.BatchNonce)
	require.NotNil(t, gotThirdBatch)

	pk.SetLastObservedEthereumBlockHeight(ctx, 5000)
	NewBlockHandler(pk).EndBlocker(ctx)

	// make sure the end blocker does delete these, as we've got a new Ethereum block height
	gotFirstBatch = input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(b1.TokenContract), b1.BatchNonce)
	require.Nil(t, gotFirstBatch)
	gotSecondBatch = input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(b2.TokenContract), b2.BatchNonce)
	require.Nil(t, gotSecondBatch)
	gotThirdBatch = input.PeggyKeeper.GetOutgoingTXBatch(ctx, common.HexToAddress(b3.TokenContract), b3.BatchNonce)
	require.NotNil(t, gotThirdBatch)
}
