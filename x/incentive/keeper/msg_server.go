package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/incentive"
	leveragetypes "github.com/umee-network/umee/v5/x/leverage/types"
)

var _ incentive.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/incentive
// module.
func NewMsgServerImpl(keeper Keeper) incentive.MsgServer {
	return &msgServer{keeper: keeper}
}

func (s msgServer) Claim(
	goCtx context.Context,
	msg *incentive.MsgClaim,
) (*incentive.MsgClaimResponse, error) {
	k, ctx := s.keeper, sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(msg.Account)
	if err != nil {
		return nil, err
	}

	// clear completed unbondings and claim all rewards
	rewards, err := k.UpdateAccount(ctx, addr)
	if err != nil {
		return nil, err
	}

	return &incentive.MsgClaimResponse{Amount: rewards}, nil
}

func (s msgServer) Bond(
	goCtx context.Context,
	msg *incentive.MsgBond,
) (*incentive.MsgBondResponse, error) {
	k, ctx := s.keeper, sdk.UnwrapSDKContext(goCtx)
	addr, denom, err := addressUToken(msg.Account, msg.UToken)
	if err != nil {
		return nil, err
	}

	// clear completed unbondings and claim all rewards
	// this must happen before bonded amount is increased, as rewards are for the previously bonded amount only
	_, err = k.UpdateAccount(ctx, addr)
	if err != nil {
		return nil, err
	}

	// get current account state for the requested uToken denom only
	bonded := k.GetBonded(ctx, addr, denom)

	// ensure account has enough collateral to bond the new amount on top of its current amount
	collateral := k.leverageKeeper.GetCollateral(ctx, addr, denom)
	if collateral.IsLT(bonded.Add(msg.UToken)) {
		return nil, incentive.ErrInsufficientCollateral.Wrapf(
			"collateral: %s bonded: %s requested: %s",
			collateral, bonded, msg.UToken,
		)
	}

	err = k.increaseBond(ctx, addr, msg.UToken)
	return &incentive.MsgBondResponse{}, err
}

func (s msgServer) BeginUnbonding(
	goCtx context.Context,
	msg *incentive.MsgBeginUnbonding,
) (*incentive.MsgBeginUnbondingResponse, error) {
	k, ctx := s.keeper, sdk.UnwrapSDKContext(goCtx)
	addr, denom, err := addressUToken(msg.Account, msg.UToken)
	if err != nil {
		return nil, err
	}

	// clear completed unbondings and claim all rewards
	// this must happen before unbonding is created, as rewards are for the previously bonded amount
	_, err = k.UpdateAccount(ctx, addr)
	if err != nil {
		return nil, err
	}

	// get current account state for the requested uToken denom only
	bonded, currentUnbonding, unbondings := k.BondSummary(ctx, addr, denom)

	maxUnbondings := int(k.GetParams(ctx).MaxUnbondings)
	if maxUnbondings > 0 && len(unbondings) >= maxUnbondings {
		// reject concurrent unbondings that would exceed max unbondings - zero is unlimited
		return nil, incentive.ErrMaxUnbondings.Wrapf("%d", len(unbondings))
	}

	// reject unbondings greater than maximum available amount
	if currentUnbonding.Add(msg.UToken).Amount.GT(bonded.Amount) {
		return nil, incentive.ErrInsufficientBonded.Wrapf(
			"bonded: %s, unbonding: %s, requested: %s",
			bonded,
			currentUnbonding,
			msg.UToken,
		)
	}

	// start the unbonding
	err = k.addUnbonding(ctx, addr, msg.UToken)
	return &incentive.MsgBeginUnbondingResponse{}, err
}

func (s msgServer) EmergencyUnbond(
	goCtx context.Context,
	msg *incentive.MsgEmergencyUnbond,
) (*incentive.MsgEmergencyUnbondResponse, error) {
	k, ctx := s.keeper, sdk.UnwrapSDKContext(goCtx)
	addr, denom, err := addressUToken(msg.Account, msg.UToken)
	if err != nil {
		return nil, err
	}

	// clear completed unbondings and claim all rewards
	// this must happen before emergency unbonding, as rewards are for the previously bonded amount
	_, err = k.UpdateAccount(ctx, addr)
	if err != nil {
		return nil, err
	}

	maxEmergencyUnbond := k.restrictedCollateral(ctx, addr, denom)

	// reject emergency unbondings greater than maximum available amount
	if msg.UToken.Amount.GT(maxEmergencyUnbond.Amount) {
		return nil, incentive.ErrInsufficientBonded.Wrapf(
			"requested: %s, maximum: %s",
			msg.UToken,
			maxEmergencyUnbond,
		)
	}

	// instant unbonding penalty is donated to the leverage module as uTokens which are immediately
	// burned. leverage reserved amount increases by token equivalent.
	penaltyAmount := k.GetParams(ctx).EmergencyUnbondFee.MulInt(msg.UToken.Amount).TruncateInt()
	if err := k.leverageKeeper.DonateCollateral(ctx, addr, sdk.NewCoin(denom, penaltyAmount)); err != nil {
		return nil, err
	}

	// reduce account's bonded and unbonding amounts, thus releasing the appropriate collateral.
	newBondPlusUnbond := maxEmergencyUnbond.Sub(msg.UToken)
	// besides the penalty fee, this is the same mechanism used to free collateral before liquidation.
	err = k.reduceBondTo(ctx, addr, newBondPlusUnbond)
	return &incentive.MsgEmergencyUnbondResponse{}, err
}

func (s msgServer) Sponsor(
	goCtx context.Context,
	msg *incentive.MsgSponsor,
) (*incentive.MsgSponsorResponse, error) {
	k, ctx := s.keeper, sdk.UnwrapSDKContext(goCtx)

	sponsor, err := sdk.AccAddressFromBech32(msg.Sponsor)
	if err != nil {
		return nil, err
	}

	err = k.sponsorIncentiveProgram(ctx, sponsor, msg.Program)
	return &incentive.MsgSponsorResponse{}, err
}

func (s msgServer) GovSetParams(
	goCtx context.Context,
	msg *incentive.MsgGovSetParams,
) (*incentive.MsgGovSetParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.Params.Validate(); err != nil {
		return &incentive.MsgGovSetParamsResponse{}, err
	}

	if err := s.keeper.setParams(ctx, msg.Params); err != nil {
		return &incentive.MsgGovSetParamsResponse{}, err
	}

	return &incentive.MsgGovSetParamsResponse{}, nil
}

func (s msgServer) GovCreatePrograms(
	goCtx context.Context,
	msg *incentive.MsgGovCreatePrograms,
) (*incentive.MsgGovCreateProgramsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// For each program being created, create it with the next available ID
	for _, program := range msg.Programs {
		if err := s.keeper.createIncentiveProgram(ctx, program, msg.FromCommunityFund); err != nil {
			return &incentive.MsgGovCreateProgramsResponse{}, err
		}
	}

	return &incentive.MsgGovCreateProgramsResponse{}, nil
}

// addressUToken parses common input fields from MsgBond and MsgBeginUnbonding, and ensures the asset is a uToken.
// Returns account as AccAddress, uToken denom and error.
func addressUToken(account string, asset sdk.Coin) (sdk.AccAddress, string, error) {
	addr, err := sdk.AccAddressFromBech32(account)
	if err != nil {
		return sdk.AccAddress{}, "", err
	}
	if !leveragetypes.HasUTokenPrefix(asset.Denom) {
		return sdk.AccAddress{}, "", leveragetypes.ErrNotUToken.Wrap(asset.Denom)
	}

	return addr, asset.Denom, err
}
