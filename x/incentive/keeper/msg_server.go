package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
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
	rewards, err := k.updateAccount(ctx, addr)
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
	addr, denom, tier, err := addressUTokenTier(msg.Account, msg.Asset, msg.Tier)
	if err != nil {
		return nil, err
	}

	// clear completed unbondings and claim all rewards
	// this must happen before bonded amount is increased, as rewards are for the previously bonded amount only
	_, err = k.updateAccount(ctx, addr)
	if err != nil {
		return nil, err
	}

	// get current account state for the requested uToken denom only
	bonded := k.getAllBonded(ctx, addr, denom)

	// ensure account has enough collateral to bond the new amount on top of its current amount
	collateral := k.leverageKeeper.GetCollateralAmount(ctx, addr, denom)
	if collateral.IsLT(bonded.Add(msg.Asset)) {
		return nil, incentive.ErrInsufficientCollateral.Wrapf(
			"collateral: %s bonded: %s requested: %s",
			collateral, bonded, msg.Asset,
		)
	}

	err = k.increaseBond(ctx, addr, tier, msg.Asset)
	return &incentive.MsgBondResponse{}, err
}

func (s msgServer) BeginUnbonding(
	goCtx context.Context,
	msg *incentive.MsgBeginUnbonding,
) (*incentive.MsgBeginUnbondingResponse, error) {
	k, ctx := s.keeper, sdk.UnwrapSDKContext(goCtx)
	addr, denom, tier, err := addressUTokenTier(msg.Account, msg.Asset, msg.Tier)
	if err != nil {
		return nil, err
	}

	// clear completed unbondings and claim all rewards
	// this must happen before unbonding is created, as rewards are for the previously bonded amount
	_, err = k.updateAccount(ctx, addr)
	if err != nil {
		return nil, err
	}

	// get current account state for the requested uToken denom and unbonding tier only
	bonded, currentUnbonding, unbondings := k.accountBonds(ctx, addr, denom, tier)

	// prevent unbonding spam
	if len(unbondings) >= int(k.GetMaxUnbondings(ctx)) {
		return nil, incentive.ErrMaxUnbondings.Wrapf("%d", len(unbondings))
	}

	// reject unbondings greater than maximum available amount
	if currentUnbonding.Add(msg.Asset).Amount.GT(bonded.Amount) {
		return nil, incentive.ErrInsufficientBonded.Wrapf(
			"bonded: %s, unbonding: %s, requested: %s",
			bonded,
			currentUnbonding,
			msg.Asset,
		)
	}

	// start the unbonding
	err = k.addUnbonding(ctx, addr, msg.Asset, tier)
	return &incentive.MsgBeginUnbondingResponse{}, err
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

	// Error messages that follow are designed to promote third party usability, so they are more
	// verbose and situational than usual.
	program, status, err := k.GetIncentiveProgram(ctx, msg.Program)
	if err != nil {
		return nil, err
	}
	if status != incentive.ProgramStatusUpcoming {
		return nil, incentive.ErrSponsorIneligible.Wrap("program exists but is not upcoming")
	}
	if !program.FundedRewards.IsZero() {
		return nil, incentive.ErrSponsorIneligible.Wrap("program is already funded")
	}
	if program.TotalRewards.Denom != msg.Asset.Denom {
		return nil, incentive.ErrSponsorInvalid.Wrap("reward denom mismatch")
	}
	if !program.TotalRewards.Amount.Equal(msg.Asset.Amount) {
		return nil, incentive.ErrSponsorInvalid.Wrap("sponsor amount must match exact total rewards required")
	}

	// transfer rewards from sponsor to incentive module
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		sponsor,
		incentive.ModuleName,
		sdk.NewCoins(program.TotalRewards),
	)
	if err != nil {
		return nil, err
	}

	// update the program's funded amount in store
	program.FundedRewards = program.TotalRewards
	err = k.SetIncentiveProgram(ctx, program, incentive.ProgramStatusUpcoming)
	return &incentive.MsgSponsorResponse{}, err
}

func (s msgServer) GovSetParams(
	goCtx context.Context,
	msg *incentive.MsgGovSetParams,
) (*incentive.MsgGovSetParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// todo: check GetSigners security, other things

	if err := msg.Params.Validate(); err != nil {
		return &incentive.MsgGovSetParamsResponse{}, err
	}

	if err := s.keeper.SetParams(ctx, msg.Params); err != nil {
		return &incentive.MsgGovSetParamsResponse{}, err
	}

	return &incentive.MsgGovSetParamsResponse{}, nil
}

func (s msgServer) GovCreatePrograms(
	goCtx context.Context,
	msg *incentive.MsgGovCreatePrograms,
) (*incentive.MsgGovCreateProgramsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// todo: check GetSigners security, other things

	// For each program being created, create it with the next available ID
	for _, program := range msg.Programs {
		if err := s.keeper.CreateIncentiveProgram(ctx, program, msg.FromCommunityFund); err != nil {
			return &incentive.MsgGovCreateProgramsResponse{}, err
		}
	}

	return &incentive.MsgGovCreateProgramsResponse{}, nil
}

// addressUTokenTier parses common input fields from MsgBond and MsgBeginUnbonding, and ensures the asset is a uToken.
func addressUTokenTier(account string, asset sdk.Coin, tierUint uint32,
) (sdk.AccAddress, string, incentive.BondTier, error) {
	addr, err := sdk.AccAddressFromBech32(account)
	if err != nil {
		return sdk.AccAddress{}, "", incentive.BondTierUnspecified, err
	}
	tier, err := bondTier(tierUint)
	if err != nil {
		return sdk.AccAddress{}, "", incentive.BondTierUnspecified, err
	}
	if !leveragetypes.HasUTokenPrefix(asset.Denom) {
		return sdk.AccAddress{}, "", incentive.BondTierUnspecified, leveragetypes.ErrNotUToken.Wrap(asset.Denom)
	}

	return addr, asset.Denom, tier, err
}
