package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	ibctransfer "github.com/umee-network/umee/v4/x/uibc"
)

// Keeper embeds the ICS-20 transfer keeper where we only override specific
// methods.
type Keeper struct {
	// embed the ICS-20 transfer keeper
	ibctransferkeeper.Keeper

	bankKeeper ibctransfer.BankKeeper
}

func New(tk ibctransferkeeper.Keeper, bk ibctransfer.BankKeeper) Keeper {
	return Keeper{
		Keeper:     tk,
		bankKeeper: bk,
	}
}

// Transfer defines a rpc handler method for MsgTransfer.
func (k Keeper) Transfer(goCtx context.Context, msg *ibctransfertypes.MsgTransfer,
) (*ibctransfertypes.MsgTransferResponse, error) {
	resp, err := k.Keeper.Transfer(goCtx, msg)
	if err != nil {
		return resp, err
	}

	// track metadata
	ctx := sdk.UnwrapSDKContext(goCtx)
	denom := msg.Token.Denom
	ibcPrefix := ibctransfertypes.DenomPrefix + "/"
	if strings.HasPrefix(denom, ibcPrefix) {
		// trim the denomination prefix, by default "ibc/"
		hexHash := denom[len(ibcPrefix):]
		hash, err := ibctransfertypes.ParseHexHash(hexHash)
		if err != nil {
			return resp, ibctransfertypes.ErrInvalidDenomForTransfer.Wrap(err.Error())
		}
		denomTrace, ok := k.GetDenomTrace(ctx, hash)
		if !ok {
			return resp, ibctransfertypes.ErrTraceNotFound.Wrap(hexHash)
		}

		k.TrackDenomMetadata(ctx, denomTrace)
	}

	return resp, err

}

// OnRecvPacket delegates the OnRecvPacket call to the embedded ICS-20 transfer
// module and updates metadata if successful.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data ibctransfertypes.FungibleTokenPacketData,
) error {
	if err := k.Keeper.OnRecvPacket(ctx, packet, data); err != nil {
		return err
	}

	// track metadata
	k.PostOnRecvPacket(ctx, packet, data)

	return nil
}

// PostOnRecvPacket executes arbitrary logic after a successful OnRecvPacket
// call. Currently, it checks and adds denomination metadata upon receiving an
// IBC asset.
func (k Keeper) PostOnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data ibctransfertypes.FungibleTokenPacketData,
) {
	var denomTrace ibctransfertypes.DenomTrace

	if ibctransfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		voucherPrefix := ibctransfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := data.Denom[len(voucherPrefix):]
		denomTrace = ibctransfertypes.ParseDenomTrace(unprefixedDenom)
	} else {
		sourcePrefix := ibctransfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
		prefixedDenom := sourcePrefix + data.Denom // sourcePrefix contains the trailing "/"
		denomTrace = ibctransfertypes.ParseDenomTrace(prefixedDenom)
	}

	k.TrackDenomMetadata(ctx, denomTrace)
}

// TrackDenomMetadata checks for the metadata existence of an IBC transferred
// asset and if it does not exist, it attempts to add it. Note, we cannot infer
// the exponent or any units so we default to zero.
func (k Keeper) TrackDenomMetadata(ctx sdk.Context, denomTrace ibctransfertypes.DenomTrace) {
	ibcDenom := denomTrace.IBCDenom()

	if _, ok := k.bankKeeper.GetDenomMetaData(ctx, ibcDenom); !ok {
		denomMetadata := banktypes.Metadata{
			Description: "IBC transferred asset",
			Display:     denomTrace.BaseDenom,
			Name:        denomTrace.BaseDenom,
			Symbol:      denomTrace.BaseDenom,
			Base:        ibcDenom,
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    ibcDenom,
					Exponent: 0,
				},
			},
		}

		k.bankKeeper.SetDenomMetaData(ctx, denomMetadata)
	}
}
