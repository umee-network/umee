package keeper

import (
	"math"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/x/peggy/types"
)

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	storeKey   sdk.StoreKey // Unexposed key to access store from sdk.Context
	paramSpace paramtypes.Subspace

	cdc            codec.BinaryCodec // The wire codec for binary encoding/decoding.
	bankKeeper     types.BankKeeper
	accountKeeper  types.AccountKeeper
	SlashingKeeper types.SlashingKeeper
	StakingKeeper  types.StakingKeeper

	attestationHandler interface {
		Handle(sdk.Context, types.EthereumClaim) error
	}
}

// NewKeeper returns a new instance of the peggy keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	accKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper,
	bankKeeper types.BankKeeper,
	slashingKeeper types.SlashingKeeper,
) Keeper {

	k := Keeper{
		cdc:            cdc,
		paramSpace:     paramSpace,
		storeKey:       storeKey,
		accountKeeper:  accKeeper,
		StakingKeeper:  stakingKeeper,
		bankKeeper:     bankKeeper,
		SlashingKeeper: slashingKeeper,
	}

	k.attestationHandler = NewAttestationHandler(bankKeeper, k)

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return k
}

/////////////////////////////
//     VALSET REQUESTS     //
/////////////////////////////

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetValsetRequest returns a new instance of the Peggy BridgeValidatorSet
// i.e. {"nonce": 1, "memebers": [{"eth_addr": "foo", "power": 11223}]}
func (k *Keeper) SetValsetRequest(ctx sdk.Context) *types.Valset {
	valset := k.GetCurrentValset(ctx)

	// If none of the bonded validators has registered eth key, then valset.Members = 0.
	if len(valset.Members) == 0 {
		return nil
	}

	k.StoreValset(ctx, valset)
	// Store the checkpoint as a legit past valset
	checkpoint := valset.GetCheckpoint(k.GetPeggyID(ctx))
	k.SetPastEthSignatureCheckpoint(ctx, checkpoint)

	ctx.EventManager().EmitTypedEvent(&types.EventMultisigUpdateRequest{
		BridgeContract: k.GetBridgeContractAddress(ctx).Hex(),
		BridgeChainId:  k.GetBridgeChainID(ctx),
		MultisigId:     valset.Nonce,
		Nonce:          valset.Nonce,
	})
	return valset
}

// StoreValset is for storing a valiator set at a given height
func (k *Keeper) StoreValset(ctx sdk.Context, valset *types.Valset) {
	store := ctx.KVStore(k.storeKey)
	valset.Height = uint64(ctx.BlockHeight())
	store.Set(types.GetValsetKey(valset.Nonce), k.cdc.MustMarshal(valset))
	k.SetLatestValsetNonce(ctx, valset.Nonce)
}

//  SetLatestValsetNonce sets the latest valset nonce
func (k *Keeper) SetLatestValsetNonce(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LatestValsetNonce, types.UInt64Bytes(nonce))
}

// StoreValsetUnsafe is for storing a valiator set at a given height
func (k *Keeper) StoreValsetUnsafe(ctx sdk.Context, valset *types.Valset) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetValsetKey(valset.Nonce), k.cdc.MustMarshal(valset))
	k.SetLatestValsetNonce(ctx, valset.Nonce)
}

// HasValsetRequest returns true if a valset defined by a nonce exists
func (k *Keeper) HasValsetRequest(ctx sdk.Context, nonce uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetValsetKey(nonce))
}

// DeleteValset deletes the valset at a given nonce from state
func (k *Keeper) DeleteValset(ctx sdk.Context, nonce uint64) {
	ctx.KVStore(k.storeKey).Delete(types.GetValsetKey(nonce))
}

// GetLatestValsetNonce returns the latest valset nonce
func (k *Keeper) GetLatestValsetNonce(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LatestValsetNonce)

	if len(bytes) == 0 {
		return 0
	}

	return types.UInt64FromBytes(bytes)
}

// GetValset returns a valset by nonce
func (k *Keeper) GetValset(ctx sdk.Context, nonce uint64) *types.Valset {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValsetKey(nonce))
	if bz == nil {
		return nil
	}

	var valset types.Valset
	k.cdc.MustUnmarshal(bz, &valset)

	return &valset
}

// IterateValsets retruns all valsetRequests
func (k *Keeper) IterateValsets(ctx sdk.Context, cb func(key []byte, val *types.Valset) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ValsetRequestKey)
	iter := prefixStore.ReverseIterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var valset types.Valset
		k.cdc.MustUnmarshal(iter.Value(), &valset)
		// cb returns true to stop early
		if cb(iter.Key(), &valset) {
			break
		}
	}
}

// GetValsets returns all the validator sets in state
func (k *Keeper) GetValsets(ctx sdk.Context) (out []*types.Valset) {
	k.IterateValsets(ctx, func(_ []byte, val *types.Valset) bool {
		out = append(out, val)
		return false
	})

	sort.Sort(types.Valsets(out))

	return
}

// GetLatestValset returns the latest validator set in state
func (k *Keeper) GetLatestValset(ctx sdk.Context) (out *types.Valset) {
	latestValsetNonce := k.GetLatestValsetNonce(ctx)
	out = k.GetValset(ctx, latestValsetNonce)

	return
}

// setLastSlashedValsetNonce sets the latest slashed valset nonce
func (k *Keeper) SetLastSlashedValsetNonce(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastSlashedValsetNonce, types.UInt64Bytes(nonce))
}

// GetLastSlashedValsetNonce returns the latest slashed valset nonce
func (k *Keeper) GetLastSlashedValsetNonce(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastSlashedValsetNonce)

	if len(bytes) == 0 {
		return 0
	}

	return types.UInt64FromBytes(bytes)
}

// SetLastUnbondingBlockHeight sets the last unbonding block height
func (k *Keeper) SetLastUnbondingBlockHeight(ctx sdk.Context, unbondingBlockHeight uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastUnbondingBlockHeight, types.UInt64Bytes(unbondingBlockHeight))
}

// GetLastUnbondingBlockHeight returns the last unbonding block height
func (k *Keeper) GetLastUnbondingBlockHeight(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastUnbondingBlockHeight)

	if len(bytes) == 0 {
		return 0
	}

	return types.UInt64FromBytes(bytes)
}

// GetUnslashedValsets returns all the unslashed validator sets in state
func (k *Keeper) GetUnslashedValsets(ctx sdk.Context, maxHeight uint64) (out []*types.Valset) {
	lastSlashedValsetNonce := k.GetLastSlashedValsetNonce(ctx)

	k.IterateValsetBySlashedValsetNonce(ctx, lastSlashedValsetNonce, maxHeight, func(_ []byte, valset *types.Valset) bool {
		if valset.Nonce > lastSlashedValsetNonce {
			out = append(out, valset)
		}
		return false
	})

	return
}

// IterateValsetBySlashedValsetNonce iterates through all valset by last slashed valset nonce in ASC order
func (k *Keeper) IterateValsetBySlashedValsetNonce(
	ctx sdk.Context,
	lastSlashedValsetNonce uint64,
	maxHeight uint64,
	cb func(k []byte, v *types.Valset) (stop bool),
) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ValsetRequestKey)
	iter := prefixStore.Iterator(types.UInt64Bytes(lastSlashedValsetNonce), types.UInt64Bytes(maxHeight))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		valset := types.Valset{}
		k.cdc.MustUnmarshal(iter.Value(), &valset)

		if cb(iter.Key(), &valset) {
			break
		}
	}
}

/////////////////////////////
//     VALSET CONFIRMS     //
/////////////////////////////

// GetValsetConfirm returns a valset confirmation by a nonce and validator address
func (k *Keeper) GetValsetConfirm(ctx sdk.Context, nonce uint64, validator sdk.AccAddress) *types.MsgValsetConfirm {
	store := ctx.KVStore(k.storeKey)
	entity := store.Get(types.GetValsetConfirmKey(nonce, validator))
	if entity == nil {
		return nil
	}

	valset := types.MsgValsetConfirm{}
	k.cdc.MustUnmarshal(entity, &valset)

	return &valset
}

// SetValsetConfirm sets a valset confirmation
func (k *Keeper) SetValsetConfirm(ctx sdk.Context, valset *types.MsgValsetConfirm) []byte {
	store := ctx.KVStore(k.storeKey)
	addr, err := sdk.AccAddressFromBech32(valset.Orchestrator)
	if err != nil {
		panic(err)
	}

	key := types.GetValsetConfirmKey(valset.Nonce, addr)
	store.Set(key, k.cdc.MustMarshal(valset))

	return key
}

// GetValsetConfirms returns all validator set confirmations by nonce
func (k *Keeper) GetValsetConfirms(ctx sdk.Context, nonce uint64) (valsetConfirms []*types.MsgValsetConfirm) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ValsetConfirmKey)
	start, end := PrefixRange(types.UInt64Bytes(nonce))
	iterator := prefixStore.Iterator(start, end)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		valset := types.MsgValsetConfirm{}

		k.cdc.MustUnmarshal(iterator.Value(), &valset)
		valsetConfirms = append(valsetConfirms, &valset)
	}

	return valsetConfirms
}

// IterateValsetConfirmByNonce iterates through all valset confirms by validator set nonce in ASC order
func (k *Keeper) IterateValsetConfirmByNonce(ctx sdk.Context, nonce uint64, cb func(k []byte, v *types.MsgValsetConfirm) (stop bool)) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ValsetConfirmKey)
	iter := prefixStore.Iterator(PrefixRange(types.UInt64Bytes(nonce)))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		valset := types.MsgValsetConfirm{}
		k.cdc.MustUnmarshal(iter.Value(), &valset)

		if cb(iter.Key(), &valset) {
			break
		}
	}
}

/////////////////////////////
//      BATCH CONFIRMS     //
/////////////////////////////

// GetBatchConfirm returns a batch confirmation given its nonce, the token contract, and a validator address
func (k *Keeper) GetBatchConfirm(ctx sdk.Context, nonce uint64, tokenContract common.Address, validator sdk.AccAddress) *types.MsgConfirmBatch {
	store := ctx.KVStore(k.storeKey)
	entity := store.Get(types.GetBatchConfirmKey(tokenContract, nonce, validator))
	if entity == nil {
		return nil
	}

	batch := types.MsgConfirmBatch{}
	k.cdc.MustUnmarshal(entity, &batch)

	return &batch
}

// SetBatchConfirm sets a batch confirmation by a validator
func (k *Keeper) SetBatchConfirm(ctx sdk.Context, batch *types.MsgConfirmBatch) []byte {
	// convert eth signer to hex string lol
	batch.EthSigner = common.HexToAddress(batch.EthSigner).Hex()
	tokenContract := common.HexToAddress(batch.TokenContract)
	store := ctx.KVStore(k.storeKey)

	acc, err := sdk.AccAddressFromBech32(batch.Orchestrator)
	if err != nil {
		panic(err)
	}

	key := types.GetBatchConfirmKey(tokenContract, batch.Nonce, acc)
	store.Set(key, k.cdc.MustMarshal(batch))

	return key
}

// IterateBatchConfirmByNonceAndTokenContract iterates through all batch confirmations
func (k *Keeper) IterateBatchConfirmByNonceAndTokenContract(
	ctx sdk.Context,
	nonce uint64,
	tokenContract common.Address,
	cb func(k []byte, v *types.MsgConfirmBatch) (stop bool),
) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.BatchConfirmKey)
	prefix := append(tokenContract.Bytes(), types.UInt64Bytes(nonce)...)
	iter := prefixStore.Iterator(PrefixRange(prefix))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		confirm := types.MsgConfirmBatch{}
		k.cdc.MustUnmarshal(iter.Value(), &confirm)

		if cb(iter.Key(), &confirm) {
			break
		}
	}
}

// GetBatchConfirmByNonceAndTokenContract returns the batch confirms
func (k *Keeper) GetBatchConfirmByNonceAndTokenContract(ctx sdk.Context, nonce uint64, tokenContract common.Address) (out []*types.MsgConfirmBatch) {
	k.IterateBatchConfirmByNonceAndTokenContract(ctx, nonce, tokenContract, func(_ []byte, msg *types.MsgConfirmBatch) (stop bool) {
		out = append(out, msg)
		return false
	})

	return
}

/////////////////////////////
//    ADDRESS DELEGATION   //
/////////////////////////////

// SetOrchestratorValidator sets the Orchestrator key for a given validator
func (k *Keeper) SetOrchestratorValidator(ctx sdk.Context, val sdk.ValAddress, orch sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetOrchestratorAddressKey(orch), val.Bytes())
}

// GetOrchestratorValidator returns the validator key associated with an orchestrator key
func (k *Keeper) GetOrchestratorValidator(ctx sdk.Context, orch sdk.AccAddress) (sdk.ValAddress, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetOrchestratorAddressKey(orch))
	if bz == nil {
		return nil, false
	}

	return sdk.ValAddress(bz), true
}

/////////////////////////////
//       ETH ADDRESS       //
/////////////////////////////

// SetEthAddressForValidator sets the ethereum address for a given validator
func (k *Keeper) SetEthAddressForValidator(ctx sdk.Context, validator sdk.ValAddress, ethAddr common.Address) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetEthAddressByValidatorKey(validator), ethAddr.Bytes())
	store.Set(types.GetValidatorByEthAddressKey(ethAddr), validator.Bytes())
}

// GetEthAddressByValidator returns the eth address for a given peggy validator
func (k *Keeper) GetEthAddressByValidator(ctx sdk.Context, validator sdk.ValAddress) (common.Address, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetEthAddressByValidatorKey(validator))
	if bz == nil {
		return common.Address{}, false
	}

	return common.BytesToAddress(bz), true
}

// GetValidatorByEthAddress returns the validator for a given eth address
func (k *Keeper) GetValidatorByEthAddress(ctx sdk.Context, ethAddr common.Address) (validator stakingtypes.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	valAddr := store.Get(types.GetValidatorByEthAddressKey(ethAddr))
	if valAddr == nil {
		return stakingtypes.Validator{}, false
	}

	validator, found = k.StakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return stakingtypes.Validator{}, false
	}

	return validator, true
}

// GetCurrentValset gets powers from the store and normalizes them
// into an integer percentage with a resolution of uint32 Max meaning
// a given validators 'Peggy power' is computed as
// Cosmos power for that validator / total cosmos power = x / uint32 Max
// where x is the voting power on the Peggy contract. This allows us
// to only use integer division which produces a known rounding error
// from truncation equal to the ratio of the validators
// Cosmos power / total cosmos power ratio, leaving us at uint32 Max - 1
// total voting power. This is an acceptable rounding error since floating
// point may cause consensus problems if different floating point unit
// implementations are involved.
//
// 'total cosmos power' has an edge case, if a validator has not set their
// Ethereum key they are not included in the total. If they where control
// of the bridge could be lost in the following situation.
//
// If we have 100 total power, and 100 total power joins the validator set
// the new validators hold more than 33% of the bridge power, if we generate
// and submit a valset and they don't have their eth keys set they can never
// update the validator set again and the bridge and all its' funds are lost.
// For this reason we exclude validators with unset eth keys from validator sets
func (k *Keeper) GetCurrentValset(ctx sdk.Context) *types.Valset {

	validators := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
	// allocate enough space for all validators, but len zero, we then append
	// so that we have an array with extra capacity but the correct length depending
	// on how many validators have keys set.
	bridgeValidators := make([]*types.BridgeValidator, 0, len(validators))
	var totalPower uint64
	for _, validator := range validators {
		val := validator.GetOperator()
		p := uint64(k.StakingKeeper.GetLastValidatorPower(ctx, val))

		if ethAddress, found := k.GetEthAddressByValidator(ctx, val); found {
			bv := &types.BridgeValidator{Power: p, EthereumAddress: ethAddress.Hex()}
			bridgeValidators = append(bridgeValidators, bv)
			totalPower += p
		}
	}

	// normalize power values
	for i := range bridgeValidators {
		bridgeValidators[i].Power = sdk.NewUint(bridgeValidators[i].Power).MulUint64(math.MaxUint32).QuoUint64(totalPower).Uint64()
	}

	// get the reward from the params store
	reward := k.GetParams(ctx).ValsetReward
	var rewardToken common.Address
	var rewardAmount sdk.Int
	if reward.Denom == "" {
		// the case where a validator has 'no reward'. The 'no reward' value is interpreted as having a zero
		// address for the ERC20 token and a zero value for the reward amount. Since we store a coin with the
		// params, a coin with a blank denom and/or zero amount is interpreted in this way.
		rewardToken = common.Address{0x0000000000000000000000000000000000000000}
		rewardAmount = sdk.NewIntFromUint64(0)

	} else {
		rewardToken, rewardAmount = k.RewardToERC20Lookup(ctx, reward)
	}
	// TODO: make the nonce an incrementing one (i.e. fetch last nonce from state, increment, set here)
	return types.NewValset(uint64(ctx.BlockHeight()), uint64(ctx.BlockHeight()), bridgeValidators, rewardAmount, rewardToken)
}

/////////////////////////////
//       PARAMETERS        //
/////////////////////////////

// GetParams returns the parameters from the store
func (k *Keeper) GetParams(ctx sdk.Context) *types.Params {
	params := new(types.Params)

	k.paramSpace.GetParamSet(ctx, params)

	return params
}

// SetParams sets the parameters in the store
func (k *Keeper) SetParams(ctx sdk.Context, params *types.Params) {
	k.paramSpace.SetParamSet(ctx, params)
}

// GetBridgeContractAddress returns the bridge contract address on ETH
func (k *Keeper) GetBridgeContractAddress(ctx sdk.Context) common.Address {
	var bridgeContractAddressHex string
	k.paramSpace.Get(ctx, types.ParamsStoreKeyBridgeContractAddress, &bridgeContractAddressHex)

	return common.HexToAddress(bridgeContractAddressHex)
}

// GetBridgeChainID returns the chain id of the ETH chain we are running against
func (k *Keeper) GetBridgeChainID(ctx sdk.Context) uint64 {
	var bridgeChainID uint64

	k.paramSpace.Get(ctx, types.ParamsStoreKeyBridgeContractChainID, &bridgeChainID)

	return bridgeChainID
}

func (k *Keeper) GetPeggyID(ctx sdk.Context) string {
	var peggyID string
	k.paramSpace.Get(ctx, types.ParamsStoreKeyPeggyID, &peggyID)

	return peggyID
}

func (k *Keeper) setPeggyID(ctx sdk.Context, v string) {

	k.paramSpace.Set(ctx, types.ParamsStoreKeyPeggyID, v)
}

func (k *Keeper) UnpackAttestationClaim(attestation *types.Attestation) (types.EthereumClaim, error) {
	var msg types.EthereumClaim

	err := k.cdc.UnpackAny(attestation.Claim, &msg)
	if err != nil {
		err = errors.Wrap(err, "failed to unpack EthereumClaim")
		return nil, err
	}

	return msg, nil
}

// GetOrchestratorAddresses iterates both the EthAddress and Orchestrator address indexes to produce
// a vector of MsgSetOrchestratorAddresses entires containing all the delgate keys for state
// export / import. This may seem at first glance to be excessively complicated, why not combine
// the EthAddress and Orchestrator address indexes and simply iterate one thing? The answer is that
// even though we set the Eth and Orchestrator address in the same place we use them differently we
// always go from Orchestrator address to Validator address and from validator address to Ethereum address
// we want to keep looking up the validator address for various reasons, so a direct Orchestrator to Ethereum
// address mapping will mean having to keep two of the same data around just to provide lookups.
//
// For the time being this will serve
func (k *Keeper) GetOrchestratorAddresses(ctx sdk.Context) []*types.MsgSetOrchestratorAddresses {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.EthAddressByValidatorKey)
	iter := store.Iterator(PrefixRange(prefix))
	defer iter.Close()

	ethAddresses := make(map[string]common.Address)

	for ; iter.Valid(); iter.Next() {
		// the 'key' contains both the prefix and the value, so we need
		// to cut off the starting bytes, if you don't do this a valid
		// cosmos key will be made out of EthAddressByValidatorKey + the startin bytes
		// of the actual key
		key := iter.Key()[len(types.EthAddressByValidatorKey):]
		value := iter.Value()
		ethAddress := common.BytesToAddress(value)
		validatorAccount := sdk.AccAddress(key)
		ethAddresses[validatorAccount.String()] = ethAddress
	}

	store = ctx.KVStore(k.storeKey)
	prefix = types.KeyOrchestratorAddress
	iter = store.Iterator(PrefixRange(prefix))
	defer iter.Close()

	orchestratorAddresses := make(map[string]sdk.AccAddress)

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()[len(types.KeyOrchestratorAddress):]
		value := iter.Value()
		orchestratorAccount := sdk.AccAddress(key)
		validatorAccount := sdk.AccAddress(value)
		orchestratorAddresses[validatorAccount.String()] = orchestratorAccount
	}

	var result []*types.MsgSetOrchestratorAddresses

	for validatorAccount, ethAddress := range ethAddresses {
		orchestratorAccount, ok := orchestratorAddresses[validatorAccount]
		if !ok {
			panic("cannot find validator account in orchestrator addresses mapping")
		}

		result = append(result, &types.MsgSetOrchestratorAddresses{
			Sender:       validatorAccount,
			Orchestrator: orchestratorAccount.String(),
			EthAddress:   ethAddress.Hex(),
		})
	}

	// we iterated over a map, so now we have to sort to ensure the
	// output here is deterministic, eth address chosen for no particular
	// reason
	sort.Slice(result[:], func(i, j int) bool {
		return result[i].EthAddress < result[j].EthAddress
	})

	return result
}

// DeserializeValidatorIterator returns validators from the validator iterator.
// Adding here in gravity keeper as cdc is not available inside endblocker.
func (k *Keeper) DeserializeValidatorIterator(vals []byte) stakingtypes.ValAddresses {
	validators := stakingtypes.ValAddresses{}
	k.cdc.MustUnmarshal(vals, &validators)
	return validators
}

// PrefixRange turns a prefix into a (start, end) range. The start is the given prefix value and
// the end is calculated by adding 1 bit to the start value. Nil is not allowed as prefix.
// 		Example: []byte{1, 3, 4} becomes []byte{1, 3, 5}
// 				 []byte{15, 42, 255, 255} becomes []byte{15, 43, 0, 0}
//
// In case of an overflow the end is set to nil.
//		Example: []byte{255, 255, 255, 255} becomes nil
// MARK finish-batches: this is where some crazy shit happens
func PrefixRange(prefix []byte) ([]byte, []byte) {
	if prefix == nil {
		panic("nil key not allowed")
	}

	// special case: no prefix is whole range
	if len(prefix) == 0 {
		return nil, nil
	}

	// copy the prefix and update last byte
	end := make([]byte, len(prefix))
	copy(end, prefix)
	l := len(end) - 1
	end[l]++

	// wait, what if that overflowed?....
	for end[l] == 0 && l > 0 {
		l--
		end[l]++
	}

	// okay, funny guy, you gave us FFF, no end to this range...
	if l == 0 && end[0] == 0 {
		end = nil
	}

	return prefix, end
}
