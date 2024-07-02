package types

import (
	context "context"

	addresscodec "cosmossdk.io/core/address"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper defines the expected interface contract defined by the x/staking
// module.
type StakingKeeper interface {
	Validator(ctx context.Context, address sdk.ValAddress) (stakingtypes.ValidatorI, error)
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	TotalBondedTokens(context.Context) (sdkmath.Int, error)
	Slash(context.Context, sdk.ConsAddress, int64, int64, sdkmath.LegacyDec) (sdkmath.Int, error)
	Jail(context.Context, sdk.ConsAddress) error
	ValidatorsPowerStoreIterator(ctx context.Context) (store.Iterator, error)
	MaxValidators(context.Context) (uint32, error)
	PowerReduction(ctx context.Context) (res sdkmath.Int)
	ValidatorAddressCodec() addresscodec.Codec
}

// DistributionKeeper defines the expected interface contract defined by the
// x/distribution module.
type DistributionKeeper interface {
	AllocateTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error
	GetValidatorOutstandingRewardsCoins(ctx context.Context, val sdk.ValAddress) (sdk.DecCoins, error)
}

// AccountKeeper defines the expected interface contract defined by the x/auth
// module.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI

	// only used for simulation
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected interface contract defined by the x/bank
// module.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
}
