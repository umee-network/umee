package simulation

import (
	"crypto/ecdsa"
	"math/rand"

	gravitykeeper "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/keeper"
	"github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/ethereum/go-ethereum/crypto"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation operation weights constants
const (
	OpWeightExec = "op_weight_simulate_replace"
)

// operations weight
const (
	DefaultWeightReplace = 100
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	sk stakingkeeper.Keeper,
	ak distrtypes.AccountKeeper,
	bk bankkeeper.Keeper,
	k gravitykeeper.Keeper,
	appCdc cdctypes.AnyUnpacker,
) simulation.WeightedOperations {

	var weightReplace int

	appParams.GetOrGenerate(cdc, OpWeightExec, &weightReplace, nil,
		func(_ *rand.Rand) {
			weightReplace = DefaultWeightReplace
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightReplace,
			SimulateValidatorReplace(k, sk, ak, bk, appCdc),
		),
	}
}

func generateEthAddress() string {
	privateKey, _ := crypto.GenerateKey()
	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return address
}

// SimulateValidatorReplace loops through the validator set and updates gravity info
func SimulateValidatorReplace(
	k gravitykeeper.Keeper,
	sk stakingkeeper.Keeper,
	ak distrtypes.AccountKeeper,
	bk bankkeeper.Keeper,
	cdc cdctypes.AnyUnpacker,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		vals := sk.GetValidatorSet()
		vals.IterateValidators(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
			account := sdk.AccAddress(validator.GetOperator())
			// check if the validator does not have an existing key
			_, foundExistingEthAddress := k.GetEthAddressByValidator(ctx, validator.GetOperator())
			_, foundExistingOrchAddress := k.GetOrchestratorValidator(ctx, account)
			if !foundExistingEthAddress || !foundExistingOrchAddress {
				ethAddr, _ := types.NewEthAddress(generateEthAddress())
				simAccount, _ := simtypes.FindAccount(accs, account)
				spendable := bk.SpendableCoins(ctx, account)
				msg := types.NewMsgSetOrchestratorAddress(validator.GetOperator(), account, *ethAddr)
				txCtx := simulation.OperationInput{
					R:               r,
					App:             app,
					TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
					Cdc:             nil,
					Msg:             msg,
					MsgType:         msg.Type(),
					Context:         ctx,
					SimAccount:      simAccount,
					AccountKeeper:   ak,
					Bankkeeper:      bk,
					ModuleName:      types.ModuleName,
					CoinsSpentInMsg: spendable,
				}
				_, _, err := simulation.GenAndDeliverTxWithRandFees(txCtx)
				if err != nil {
					panic("unable to update gravity validator set")
				}
			}
			return false
		})
		return simtypes.NewOperationMsgBasic("gravity", "MsgSetOrchestratorAddress", "validators updated", true, nil), nil, nil
	}
}
