package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/umee-network/umee/x/peggy/types"
)

func (k *Keeper) GetCosmosOriginatedDenom(ctx sdk.Context, tokenContract common.Address) (string, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetERC20ToCosmosDenomKey(tokenContract))
	if bz == nil {
		return "", false
	}

	return string(bz), true
}

func (k *Keeper) GetCosmosOriginatedERC20(ctx sdk.Context, denom string) (common.Address, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetCosmosDenomToERC20Key(denom))
	if bz == nil {
		return common.Address{}, false
	}

	return common.BytesToAddress(bz), true
}

func (k *Keeper) SetCosmosOriginatedDenomToERC20(ctx sdk.Context, denom string, tokenContract common.Address) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetCosmosDenomToERC20Key(denom), tokenContract.Bytes())
	store.Set(types.GetERC20ToCosmosDenomKey(tokenContract), []byte(denom))
}

// DenomToERC20Lookup returns an ERC20 token contract address and a boolean
// signaling if the denomination is a Cosmos native asset or not. It will return
// an error if it cannot parse the denom as a Peggy denom or if it cannot find
// the denom in an index of ERC20 contracts deployed on Ethereum to serve as
// synthetic Cosmos assets.
func (k *Keeper) DenomToERC20Lookup(ctx sdk.Context, denom string) (bool, common.Address, error) {
	// first try parsing the ERC20 out of the denom and if no error is returned,
	// we treat the asset as Ethereum-originated.
	peggyDenom, denomErr := types.NewPeggyDenomFromString(denom)
	if denomErr == nil {
		tokenContractFromDenom, err := peggyDenom.TokenContract()
		if err != nil {
			return false, common.Address{}, err
		}

		return false, tokenContractFromDenom, nil
	}

	// look up ERC20 contract in index and error if it's not in there
	tokenContract, exists := k.GetCosmosOriginatedERC20(ctx, denom)
	if !exists {
		return false, common.Address{}, fmt.Errorf(
			"denom (%s) not a peggy voucher coin (parse error: %s), and also not in cosmos-originated ERC20 index",
			denom, denomErr.Error(),
		)
	}

	return true, tokenContract, nil
}

// RewardToERC20Lookup is a specialized function wrapping DenomToERC20Lookup designed to validate
// the validator set reward any time we generate a validator set
func (k *Keeper) RewardToERC20Lookup(ctx sdk.Context, coin sdk.Coin) (common.Address, sdk.Int) {
	if len(coin.Denom) == 0 || coin.Amount.BigInt() == nil || coin.Amount == sdk.NewInt(0) {
		panic("Bad validator set relaying reward!")
	} else {
		// reward case, pass to DenomToERC20Lookup
		_, addressStr, err := k.DenomToERC20Lookup(ctx, coin.Denom)
		if err != nil {
			// This can only ever happen if governance sets a value for the reward
			// which is not a valid ERC20 that as been bridged before (either from or to Cosmos)
			// We'll classify that as operator error and just panic
			panic("Invalid Valset reward! Correct or remove the paramater value")
		}
		err = types.ValidateEthAddress(addressStr.Hex())
		if err != nil {
			panic("Invalid Valset reward! Correct or remove the paramater value")
		}
		return addressStr, coin.Amount
	}
}

// ERC20ToDenomLookup attempts to do a reverse lookup for denomination by an ERC20
// token contract address and returns the corresponding denomination and a boolean
// signaling if the asset is a Cosmos native asset or not.
func (k *Keeper) ERC20ToDenomLookup(ctx sdk.Context, tokenContract common.Address) (bool, string) {
	// first, try looking up tokenContract by index
	denom, exists := k.GetCosmosOriginatedDenom(ctx, tokenContract)
	if exists {
		return true, denom
	}

	// Since the index does not exist, it is not a cosmos originated token. So we
	// return the ERC20 as a Peggy denom.
	return false, types.NewPeggyDenom(tokenContract).String()
}

// IterateERC20ToDenom iterates over erc20 to denom relations
func (k *Keeper) IterateERC20ToDenom(ctx sdk.Context, cb func(k []byte, v *types.ERC20ToDenom) (stop bool)) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ERC20ToDenomKey)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		erc20ToDenom := types.ERC20ToDenom{
			Erc20: common.BytesToAddress(iter.Key()).Hex(),
			Denom: string(iter.Value()),
		}

		if cb(iter.Key(), &erc20ToDenom) {
			break
		}
	}
}
