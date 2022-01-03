package keeper

import (
	"fmt"
	"math/big"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/umee-network/umee/x/peggy/types"
)

// AttestationHandler processes 'observed' attestations.
type AttestationHandler struct {
	keeper     Keeper
	bankKeeper types.BankKeeper
}

func NewAttestationHandler(bankKeeper types.BankKeeper, keeper Keeper) AttestationHandler {
	return AttestationHandler{
		keeper:     keeper,
		bankKeeper: bankKeeper,
	}
}

// Handle is the entry point for Attestation processing.
func (a AttestationHandler) Handle(ctx sdk.Context, claim types.EthereumClaim) error {
	switch claim := claim.(type) {
	// deposit in this context means a deposit into the Ethereum side of the bridge
	case *types.MsgDepositClaim:
		// Check if coin is Cosmos-originated asset and get denom
		isCosmosOriginated, denom := a.keeper.ERC20ToDenomLookup(ctx, common.HexToAddress(claim.TokenContract))

		if isCosmosOriginated {
			// If it is cosmos originated, unlock the coins
			coins := sdk.Coins{
				sdk.NewCoin(denom, claim.Amount),
			}

			addr, err := sdk.AccAddressFromBech32(claim.CosmosReceiver)
			if err != nil {
				return sdkerrors.Wrap(err, "invalid receiver address")
			}

			if err = a.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins); err != nil {
				return sdkerrors.Wrap(err, "transfer vouchers")
			}
		} else {
			// Check if supply overflows with claim amount
			currentSupply := a.bankKeeper.GetSupply(ctx, denom)
			newSupply := new(big.Int).Add(currentSupply.Amount.BigInt(), claim.Amount.BigInt())
			if newSupply.BitLen() > 256 {
				return sdkerrors.Wrap(types.ErrSupplyOverflow, "invalid supply")
			}

			// If it is not cosmos originated, mint the coins (aka vouchers)
			coins := sdk.Coins{
				sdk.NewCoin(denom, claim.Amount),
			}

			if err := a.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
				return sdkerrors.Wrapf(err, "mint vouchers coins: %s", coins)
			}

			addr, err := sdk.AccAddressFromBech32(claim.CosmosReceiver)
			if err != nil {
				return sdkerrors.Wrap(err, "invalid receiver address")
			}

			if err = a.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins); err != nil {
				return sdkerrors.Wrap(err, "transfer vouchers")
			}
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("MsgSendToCosmosAmount", claim.Amount.String()),
				sdk.NewAttribute("MsgSendToCosmosNonce", strconv.Itoa(int(claim.GetEventNonce()))),
				sdk.NewAttribute("MsgSendToCosmosToken", claim.TokenContract),
			),
		)

		// withdraw in this context means a withdraw from the Ethereum side of the bridge
	case *types.MsgWithdrawClaim:
		tokenContract := common.HexToAddress(claim.TokenContract)
		a.keeper.OutgoingTxBatchExecuted(ctx, tokenContract, claim.BatchNonce)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("MsgBatchSendToEthClaim", strconv.Itoa(int(claim.BatchNonce))),
			),
		)
		return nil

	case *types.MsgERC20DeployedClaim:
		if err := a.verifyERC20DeployedEvent(ctx, claim); err != nil {
			return err
		}

		// add to ERC20 mapping
		a.keeper.SetCosmosOriginatedDenomToERC20(ctx, claim.CosmosDenom, common.HexToAddress(claim.TokenContract))

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("MsgERC20DeployedClaimToken", claim.TokenContract),
				sdk.NewAttribute("MsgERC20DeployedClaim", strconv.Itoa(int(claim.GetEventNonce()))),
			),
		)

		return nil

	case *types.MsgValsetUpdatedClaim:
		// TODO here we should check the contents of the validator set against
		// the store, if they differ we should take some action to indicate to the
		// user that bridge highjacking has occurred
		a.keeper.SetLastObservedValset(ctx, types.Valset{
			Nonce:        claim.ValsetNonce,
			Members:      claim.Members,
			RewardAmount: claim.RewardAmount,
			RewardToken:  claim.RewardToken,
		})

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("MsgValsetUpdatedClaim", strconv.Itoa(int(claim.GetEventNonce()))),
			),
		)

	default:
		panic(fmt.Sprintf("invalid event type for attestations %s", claim.GetType()))
	}

	return nil
}

func (a AttestationHandler) verifyERC20DeployedEvent(ctx sdk.Context, claim *types.MsgERC20DeployedClaim) error {
	// check if it already exists
	existingERC20, exists := a.keeper.GetCosmosOriginatedERC20(ctx, claim.CosmosDenom)
	if exists {
		return sdkerrors.Wrap(
			types.ErrInvalid,
			fmt.Sprintf("ERC20 %s already exists for denom %s", existingERC20, claim.CosmosDenom))
	}

	// We expect that all Cosmos-based tokens have metadata defined. In the case
	// a token does not have metadata defined, e.g. an IBC token, we successfully
	// handle the token under the following conditions:
	//
	// 1. The ERC20 name is equal to the token's denomination. Otherwise, this
	// 		means that ERC20 tokens would have an untenable UX.
	// 2. The ERC20 token has zero decimals as this is what we default to since
	// 		we cannot know or infer the real decimal value for the Cosmos token.
	// 3. The ERC20 symbol is empty.
	//
	// NOTE: This path is not encouraged and all supported assets should have
	// metadata defined. If metadata cannot be defined, consider adding the token's
	// metadata on the fly.
	if md, ok := a.keeper.bankKeeper.GetDenomMetaData(ctx, claim.CosmosDenom); ok && md.Base != "" {
		return verifyERC20Token(md, claim)
	}

	if supply := a.keeper.bankKeeper.GetSupply(ctx, claim.CosmosDenom); supply.IsZero() {
		return sdkerrors.Wrapf(
			types.ErrInvalidERC20Event,
			"no supply exists for token %s without metadata", claim.CosmosDenom,
		)
	}

	if claim.Name != claim.CosmosDenom {
		return sdkerrors.Wrapf(
			types.ErrInvalidERC20Event,
			"invalid ERC20 name for token without metadata; got: %s, expected: %s", claim.Name, claim.CosmosDenom,
		)
	}

	if claim.Symbol != "" {
		return sdkerrors.Wrapf(
			types.ErrInvalidERC20Event,
			"expected empty ERC20 symbol for token without metadata; got: %s", claim.Symbol,
		)
	}

	if claim.Decimals != 0 {
		return sdkerrors.Wrapf(
			types.ErrInvalidERC20Event,
			"expected zero ERC20 decimals for token without metadata; got: %d", claim.Decimals,
		)
	}

	return nil
}

func verifyERC20Token(metadata banktypes.Metadata, claim *types.MsgERC20DeployedClaim) error {
	if claim.Name != metadata.Display {
		return sdkerrors.Wrapf(
			types.ErrInvalidERC20Event,
			"ERC20 name %s does not match the denom display %s", claim.Name, metadata.Display,
		)
	}

	if claim.Symbol != metadata.Display {
		return sdkerrors.Wrapf(
			types.ErrInvalidERC20Event,
			"ERC20 symbol %s does not match denom display %s", claim.Symbol, metadata.Display,
		)
	}

	// ERC20 tokens use a very simple mechanism to tell you where to display the
	// decimal point. The "decimals" field simply tells you how many decimal places
	// there will be.
	//
	// Cosmos denoms have a system that is much more full featured, with
	// enterprise-ready token denominations. There is a DenomUnits array that
	// tells you what the name of each denomination of the token is.
	//
	// To correlate this with an ERC20 "decimals" field, we have to search through
	// the DenomUnits array to find the DenomUnit which matches up to the main
	// token "display" value. Then we take the "exponent" from this DenomUnit.
	//
	// If the correct DenomUnit is not found, it will default to 0. This will
	// result in there being no decimal places in the token's ERC20 on Ethereum.
	// For example, if this happened with ATOM, 1 ATOM would appear on Ethereum
	// as 1 million ATOM, having 6 extra places before the decimal point.
	var decimals uint32
	for _, denomUnit := range metadata.DenomUnits {
		if denomUnit.Denom == metadata.Display {
			decimals = denomUnit.Exponent
			break
		}
	}

	if uint64(decimals) != claim.Decimals {
		return sdkerrors.Wrapf(
			types.ErrInvalidERC20Event,
			"ERC20 decimals %d does not match denom decimals %d", claim.Decimals, decimals,
		)
	}

	return nil
}
