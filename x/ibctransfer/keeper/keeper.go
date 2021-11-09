package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v2/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"

	"github.com/umee-network/umee/x/ibctransfer/types"
)

// Keeper embeds the ICS-20 transfer keeper where we only override specific
// methods.
type Keeper struct {
	// embed the ICS-20 transfer keeper
	ibctransferkeeper.Keeper

	bankKeeper types.BankKeeper
}

func New(tk ibctransferkeeper.Keeper, bk types.BankKeeper) Keeper {
	return Keeper{
		Keeper:     tk,
		bankKeeper: bk,
	}
}

// SendTransfer delegates the SendTransfer call to the embedded ICS-20 transfer
// module and updates metadata if successful.
func (k Keeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {

	// first, relay the SendTransfer to the real (embedded) ICS-20 transfer keeper
	if err := k.Keeper.SendTransfer(
		ctx,
		sourcePort,
		sourceChannel,
		token,
		sender,
		receiver,
		timeoutHeight,
		timeoutTimestamp,
	); err != nil {
		return err
	}

	// track metadata
	ibcPrefix := ibctransfertypes.DenomPrefix + "/"
	if strings.HasPrefix(token.Denom, ibcPrefix) {
		// trim the denomination prefix, by default "ibc/"
		hexHash := token.Denom[len(ibcPrefix):]

		hash, err := ibctransfertypes.ParseHexHash(hexHash)
		if err != nil {
			return sdkerrors.Wrap(ibctransfertypes.ErrInvalidDenomForTransfer, err.Error())
		}

		denomTrace, ok := k.GetDenomTrace(ctx, hash)
		if !ok {
			return sdkerrors.Wrap(ibctransfertypes.ErrTraceNotFound, hexHash)
		}

		k.TrackDenomMetadata(ctx, denomTrace)
	}

	return nil
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
