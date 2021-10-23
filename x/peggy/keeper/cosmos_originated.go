package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

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

// DenomToERC20 returns if an asset is native to Cosmos or Ethereum, and get its corresponding ERC20 address
// This will return an error if it cant parse the denom as a peggy denom, and then also can't find the denom
// in an index of ERC20 contracts deployed on Ethereum to serve as synthetic Cosmos assets.
func (k *Keeper) DenomToERC20Lookup(ctx sdk.Context, denomStr string) (isCosmosOriginated bool, tokenContract common.Address, err error) {
	// First try parsing the ERC20 out of the denom
	peggyDenom, denomErr := types.NewPeggyDenomFromString(denomStr)
	if denomErr == nil {
		// This is an Ethereum-originated asset
		tokenContractFromDenom, _ := peggyDenom.TokenContract()
		return false, tokenContractFromDenom, nil
	}

	// If denom is native cosmos coin denom, return Cosmos coin ERC20 contract address from Params
	if denomStr == k.GetCosmosCoinDenom(ctx) {
		// isCosmosOriginated assumed to be false, since the native cosmos coin
		// expected to be mapped from Ethereum mainnet in first place, i.e. its origin
		// is still from Ethereum.
		return false, k.GetCosmosCoinERC20Contract(ctx), nil
	}

	// Look up ERC20 contract in index and error if it's not in there
	tokenContract, exists := k.GetCosmosOriginatedERC20(ctx, denomStr)
	if !exists {
		err = errors.Errorf(
			"denom (%s) not a peggy voucher coin (parse error: %s), and also not in cosmos-originated ERC20 index",
			denomStr, denomErr.Error(),
		)

		return false, common.Address{}, err
	}

	isCosmosOriginated = true
	return isCosmosOriginated, tokenContract, nil
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

// ERC20ToDenom returns if an ERC20 address represents an asset is native to Cosmos or Ethereum,
// and get its corresponding peggy denom.
func (k *Keeper) ERC20ToDenomLookup(ctx sdk.Context, tokenContract common.Address) (isCosmosOriginated bool, denom string) {
	// First try looking up tokenContract in index
	denomStr, exists := k.GetCosmosOriginatedDenom(ctx, tokenContract)
	if exists {
		isCosmosOriginated = true
		return isCosmosOriginated, denomStr
	} else if tokenContract == k.GetCosmosCoinERC20Contract(ctx) {
		return false, k.GetCosmosCoinDenom(ctx)
	}

	// If it is not in there, it is not a cosmos originated token, turn the ERC20 into a peggy denom
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
