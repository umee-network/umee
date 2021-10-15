package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper is expected keeper for staking module
type StakingKeeper interface {
	// get validator by operator address; nil when validator not found
	Validator(ctx sdk.Context, address sdk.ValAddress) stakingtypes.ValidatorI
	// total bonded tokens within the validator set
	TotalBondedTokens(sdk.Context) sdk.Int
	// slash the validator and delegators of the validator,
	// specifying offense height, offense
	// power, and slash fraction
	Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec)
	// jail a validator
	Jail(sdk.Context, sdk.ConsAddress)
	// an iterator for the current validator power store
	ValidatorsPowerStoreIterator(ctx sdk.Context) sdk.Iterator
	// MaxValidators returns the
	// maximum amount of bonded validators
	MaxValidators(sdk.Context) uint32
	PowerReduction(ctx sdk.Context) (res sdk.Int)
}

// DistributionKeeper is expected keeper for distribution module
type DistributionKeeper interface {
	AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins)

	// only used for simulation
	GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins
}

// AccountKeeper is expected keeper for auth module
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI // only used for simulation
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)

	// only used for simulation
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
