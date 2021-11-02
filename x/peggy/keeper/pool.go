package keeper

import (
	"encoding/binary"
	"math/big"
	"sort"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/umee-network/umee/x/peggy/types"
)

// AddToOutgoingPool
// - checks a counterpart denominator exists for the given voucher type
// - burns the voucher for transfer amount and fees
// - persists an OutgoingTx
// - adds the TX to the `available` TX pool via a second index
func (k *Keeper) AddToOutgoingPool(ctx sdk.Context, sender sdk.AccAddress, counterpartReceiver common.Address, amount sdk.Coin, fee sdk.Coin) (uint64, error) {
	totalAmount := amount.Add(fee)
	totalInVouchers := sdk.Coins{totalAmount}

	// If the coin is a peggy voucher, burn the coins. If not, check if there is a deployed ERC20 contract representing it.
	// If there is, lock the coins.

	isCosmosOriginated, tokenContract, err := k.DenomToERC20Lookup(ctx, totalAmount.Denom)
	if err != nil {
		return 0, err
	}

	// If it is a cosmos-originated asset we lock it
	if isCosmosOriginated {
		// lock coins in module
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, totalInVouchers); err != nil {
			return 0, err
		}
	} else {
		// If it is an ethereum-originated asset we burn it
		// send coins to module in prep for burn
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, totalInVouchers); err != nil {
			return 0, err
		}

		// burn vouchers to send them back to ETH
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, totalInVouchers); err != nil {
			panic(err)
		}
	}

	// get next tx id from keeper
	nextID := k.AutoIncrementID(ctx, types.KeyLastTXPoolID)
	erc20Fee := types.NewSDKIntERC20Token(fee.Amount, tokenContract)

	// construct outgoing tx, as part of this process we represent
	// the token as an ERC20 token since it is preparing to go to ETH
	// rather than the denom that is the input to this function.
	outgoing := &types.OutgoingTransferTx{
		Id:          nextID,
		Sender:      sender.String(),
		DestAddress: counterpartReceiver.Hex(),
		Erc20Token:  types.NewSDKIntERC20Token(amount.Amount, tokenContract),
		Erc20Fee:    erc20Fee,
	}

	// set the outgoing tx in the pool index
	if err := k.setPoolEntry(ctx, outgoing); err != nil {
		return 0, err
	}

	// add a second index with the fee
	k.appendToUnbatchedTXIndex(ctx, tokenContract, erc20Fee, nextID)

	// todo: add second index for sender so that we can easily query: give pending Tx by sender
	// todo: what about a second index for receiver?

	ctx.EventManager().EmitTypedEvent(&types.EventBridgeWithdrawalReceived{
		BridgeContract: k.GetBridgeContractAddress(ctx).Hex(),
		BridgeChainId:  k.GetBridgeChainID(ctx),
		OutgoingTxId:   nextID,
		Nonce:          nextID,
	})

	return nextID, nil
}

// RemoveFromOutgoingPoolAndRefund
// - checks that the provided tx actually exists
// - deletes the un-batched tx from the pool
// - issues the tokens back to the sender
func (k *Keeper) RemoveFromOutgoingPoolAndRefund(ctx sdk.Context, txId uint64, sender sdk.AccAddress) error {
	// check that we actually have a tx with that id and what it's details are
	tx, err := k.getPoolEntry(ctx, txId)
	if err != nil {
		return err
	}

	txSender, err := sdk.AccAddressFromBech32(tx.Sender)
	if err != nil {
		return err
	}

	if !sender.Equals(txSender) {
		return sdkerrors.Wrapf(types.ErrInvalid, "invalid sender address")
	}

	// An inconsistent entry should never enter the store, but this is the ideal
	// place to exploit it such a bug if it did ever occur, so we should double
	// check to be really sure
	if tx.Erc20Fee.Contract != tx.Erc20Token.Contract {
		return sdkerrors.Wrapf(types.ErrInvalid, "inconsistent tokens to cancel: %s %s", tx.Erc20Fee.Contract, tx.Erc20Token.Contract)
	}

	found := false
	poolTx := k.GetPoolTransactions(ctx)
	for _, pTx := range poolTx {
		if pTx.Id == txId {
			found = true
		}
	}
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalid, "txId %d is not in unbatched pool! Must be in batch!", txId)
	}

	// delete this tx from both indexes
	err = k.removeFromUnbatchedTXIndex(ctx, common.HexToAddress(tx.Erc20Token.Contract), tx.Erc20Fee, txId)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalid, "txId %d not in unbatched index! Must be in a batch!", txId)
	}
	k.removePoolEntry(ctx, txId)

	// reissue the amount and the fee
	totalToRefundCoins := sdk.Coins{}
	isCosmosOriginated, _ := k.ERC20ToDenomLookup(ctx, common.HexToAddress(tx.Erc20Token.Contract))

	// native cosmos coin denom
	totalToRefund := tx.Erc20Token.PeggyCoin()
	totalToRefund.Amount = totalToRefund.Amount.Add(tx.Erc20Fee.Amount)
	totalToRefundCoins = sdk.NewCoins(totalToRefund)

	// If it is a cosmos-originated the coins are in the module (see AddToOutgoingPool) so we can just take them out
	if isCosmosOriginated {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, totalToRefundCoins); err != nil {
			return err
		}
	} else {
		// If it is an ethereum-originated asset we have to mint it (see Handle in attestation_handler.go)
		// mint coins in module for prep to send
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, totalToRefundCoins); err != nil {
			return sdkerrors.Wrapf(err, "mint vouchers coins: %s", totalToRefundCoins)
		}
		if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, totalToRefundCoins); err != nil {
			return sdkerrors.Wrap(err, "transfer vouchers")
		}
	}

	ctx.EventManager().EmitTypedEvent(&types.EventBridgeWithdrawCanceled{
		BridgeContract: k.GetBridgeContractAddress(ctx).Hex(),
		BridgeChainId:  k.GetBridgeChainID(ctx),
	})

	return nil
}

// appendToUnbatchedTXIndex add at the end when tx with same fee exists
func (k *Keeper) appendToUnbatchedTXIndex(ctx sdk.Context, tokenContract common.Address, fee *types.ERC20Token, txID uint64) {
	store := ctx.KVStore(k.storeKey)
	idxKey := types.GetFeeSecondIndexKey(tokenContract, fee)
	var idSet types.IDSet
	if store.Has(idxKey) {
		bz := store.Get(idxKey)
		k.cdc.MustUnmarshal(bz, &idSet)
	}
	idSet.Ids = append(idSet.Ids, txID)
	store.Set(idxKey, k.cdc.MustMarshal(&idSet))
}

// appendToUnbatchedTXIndex add at the top when tx with same fee exists
func (k *Keeper) prependToUnbatchedTXIndex(ctx sdk.Context, tokenContract common.Address, fee *types.ERC20Token, txID uint64) {
	store := ctx.KVStore(k.storeKey)
	idxKey := types.GetFeeSecondIndexKey(tokenContract, fee)
	var idSet types.IDSet
	if store.Has(idxKey) {
		bz := store.Get(idxKey)
		k.cdc.MustUnmarshal(bz, &idSet)
	}

	idSet.Ids = append([]uint64{txID}, idSet.Ids...)
	store.Set(idxKey, k.cdc.MustMarshal(&idSet))
}

// removeFromUnbatchedTXIndex removes the tx from the index and also removes it from the iterator
// GetPoolTransactions, making this tx implicitly invisible without a direct request. We remove a tx
// from the pool for good in OutgoingTxBatchExecuted, but if a batch is canceled or timed out we 'reactivate'
// an entry by adding it back to the second index.
func (k *Keeper) removeFromUnbatchedTXIndex(ctx sdk.Context, tokenContract common.Address, fee *types.ERC20Token, txID uint64) error {
	store := ctx.KVStore(k.storeKey)
	idxKey := types.GetFeeSecondIndexKey(tokenContract, fee)

	var idSet types.IDSet
	bz := store.Get(idxKey)
	if bz == nil {
		return sdkerrors.Wrap(types.ErrUnknown, "fee")
	}

	k.cdc.MustUnmarshal(bz, &idSet)
	for i := range idSet.Ids {
		if idSet.Ids[i] == txID {
			idSet.Ids = append(idSet.Ids[0:i], idSet.Ids[i+1:]...)
			if len(idSet.Ids) != 0 {
				store.Set(idxKey, k.cdc.MustMarshal(&idSet))
			} else {
				store.Delete(idxKey)
			}
			return nil
		}
	}

	return sdkerrors.Wrap(types.ErrUnknown, "tx id")
}

func (k *Keeper) setPoolEntry(ctx sdk.Context, outgoingTransferTx *types.OutgoingTransferTx) error {
	bz, err := k.cdc.Marshal(outgoingTransferTx)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetOutgoingTxPoolKey(outgoingTransferTx.Id), bz)

	return nil
}

// getPoolEntry grabs an entry from the tx pool, this *does* include transactions in batches
// so check the UnbatchedTxIndex or call GetPoolTransactions for that purpose
func (k *Keeper) getPoolEntry(ctx sdk.Context, id uint64) (*types.OutgoingTransferTx, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetOutgoingTxPoolKey(id))
	if bz == nil {
		return nil, types.ErrUnknown
	}

	var r types.OutgoingTransferTx
	k.cdc.Unmarshal(bz, &r)

	return &r, nil
}

// removePoolEntry removes an entry from the tx pool, this *does* include transactions in batches
// so you will need to run it when cleaning up after a executed batch
func (k *Keeper) removePoolEntry(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOutgoingTxPoolKey(id))
}

// GetPoolTransactions, grabs all transactions from the tx pool, useful for queries or genesis save/load
// this does not include all transactions in batches, because it iterates using the second index key
func (k *Keeper) GetPoolTransactions(ctx sdk.Context) []*types.OutgoingTransferTx {
	prefixStore := ctx.KVStore(k.storeKey)
	// we must use the second index key here because transactions are left in the store, but removed
	// from the tx sorting key, while in batches
	iter := prefixStore.ReverseIterator(PrefixRange([]byte(types.SecondIndexOutgoingTXFeeKey)))

	var ret []*types.OutgoingTransferTx
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ids types.IDSet
		k.cdc.MustUnmarshal(iter.Value(), &ids)
		for _, id := range ids.Ids {
			tx, err := k.getPoolEntry(ctx, id)
			if err != nil {
				panic("Invalid id in tx index!")
			}
			ret = append(ret, tx)
		}
	}

	return ret
}

// IterateOutgoingPoolByFee itetates over the outgoing pool which is sorted by fee
func (k *Keeper) IterateOutgoingPoolByFee(ctx sdk.Context, tokenContract common.Address, cb func(uint64, *types.OutgoingTransferTx) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.SecondIndexOutgoingTXFeeKey)
	iter := prefixStore.ReverseIterator(PrefixRange(tokenContract.Bytes()))

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var ids types.IDSet
		k.cdc.MustUnmarshal(iter.Value(), &ids)
		// cb returns true to stop early
		for _, id := range ids.Ids {
			tx, err := k.getPoolEntry(ctx, id)
			if err != nil {
				panic("Invalid id in tx index!")
			}
			if cb(id, tx) {
				return
			}
		}
	}

	return
}

// GetBatchFeesByTokenType gets the fees the next batch of a given token type would
// have if created. This info is both presented to relayers for the purpose of determining
// when to request batches and also used by the batch creation process to decide not to create
// a new batch
func (k *Keeper) GetBatchFeesByTokenType(ctx sdk.Context, tokenContractAddr common.Address) *types.BatchFees {
	batchFeesMap := k.createBatchFees(ctx)
	return batchFeesMap[tokenContractAddr]
}

// GetAllBatchFees creates a fee entry for every batch type currently in the store
// this can be used by relayers to determine what batch types are desirable to request
func (k *Keeper) GetAllBatchFees(ctx sdk.Context) (batchFees []*types.BatchFees) {
	batchFeesMap := k.createBatchFees(ctx)
	// create array of batchFees
	for _, batchFee := range batchFeesMap {
		// newBatchFee := types.BatchFees{
		// 	Token:         batchFee.Token,
		// 	TopOneHundred: batchFee.TopOneHundred,
		// }
		batchFees = append(batchFees, batchFee)
	}

	// quick sort by token to make this function safe for use
	// in consensus computations
	sort.Slice(batchFees, func(i, j int) bool {
		return batchFees[i].Token < batchFees[j].Token
	})

	return batchFees
}

// CreateBatchFees iterates over the outgoing pool and creates batch token fee map
func (k *Keeper) createBatchFees(ctx sdk.Context) map[common.Address]*types.BatchFees {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.SecondIndexOutgoingTXFeeKey)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	batchFeesMap := make(map[common.Address]*types.BatchFees)
	txCountMap := make(map[common.Address]int)

	for ; iter.Valid(); iter.Next() {
		var ids types.IDSet
		k.cdc.MustUnmarshal(iter.Value(), &ids)

		// create a map to store the token contract address and its total fee
		// Parse the iterator key to get contract address & fee
		// If len(ids.Ids) > 1, multiply fee amount with len(ids.Ids) and add it to total fee amount

		key := iter.Key()
		tokenContractBytes := key[:types.ETHContractAddressLen]
		tokenContractAddr := common.BytesToAddress(tokenContractBytes)

		feeAmountBytes := key[len(tokenContractBytes):]
		feeAmount := big.NewInt(0).SetBytes(feeAmountBytes)

		for i := 0; i < len(ids.Ids); i++ {
			if txCountMap[tokenContractAddr] >= OutgoingTxBatchSize {
				break
			} else {
				// add fee amount
				if _, ok := batchFeesMap[tokenContractAddr]; ok {
					totalFees := batchFeesMap[tokenContractAddr].TotalFees
					totalFees = totalFees.Add(sdk.NewIntFromBigInt(feeAmount))
					batchFeesMap[tokenContractAddr].TotalFees = totalFees
				} else {
					batchFeesMap[tokenContractAddr] = &types.BatchFees{
						Token:     tokenContractAddr.Hex(),
						TotalFees: sdk.NewIntFromBigInt(feeAmount)}
				}

				txCountMap[tokenContractAddr] = txCountMap[tokenContractAddr] + 1
			}
		}
	}

	return batchFeesMap
}

func (k *Keeper) AutoIncrementID(ctx sdk.Context, idKey []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(idKey)

	var id uint64
	if bz != nil {
		id = binary.BigEndian.Uint64(bz) + 1
	}

	bz = sdk.Uint64ToBigEndian(id)
	store.Set(idKey, bz)

	return id
}

func (k *Keeper) GetLastOutgoingBatchID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyLastOutgoingBatchID
	var id uint64
	bz := store.Get(key)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	return id
}

func (k *Keeper) SetLastOutgoingBatchID(ctx sdk.Context, lastOutgoingBatchID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyLastOutgoingBatchID
	bz := sdk.Uint64ToBigEndian(lastOutgoingBatchID)
	store.Set(key, bz)
}

func (k *Keeper) GetLastOutgoingPoolID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyLastTXPoolID
	var id uint64
	bz := store.Get(key)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	return id
}

func (k *Keeper) SetLastOutgoingPoolID(ctx sdk.Context, lastOutgoingPoolID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyLastTXPoolID
	bz := sdk.Uint64ToBigEndian(lastOutgoingPoolID)
	store.Set(key, bz)
}
