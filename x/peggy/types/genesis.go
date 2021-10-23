package types

import (
	"bytes"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultParamspace defines the default auth module parameter subspace
const (
	// todo: implement oracle constants as params
	DefaultParamspace = ModuleName
)

var (
	// AttestationVotesPowerThreshold threshold of votes power to succeed
	AttestationVotesPowerThreshold = sdk.NewInt(66)

	// ParamsStoreKeyPeggyID stores the peggy id
	ParamsStoreKeyPeggyID = []byte("PeggyID")

	// ParamsStoreKeyContractHash stores the contract hash
	ParamsStoreKeyContractHash = []byte("ContractHash")

	// ParamsStoreKeyBridgeContractAddress stores the contract address
	ParamsStoreKeyBridgeContractAddress = []byte("BridgeContractAddress")

	// ParamsStoreKeyBridgeContractStartHeight stores the bridge contract deployed height
	ParamsStoreKeyBridgeContractStartHeight = []byte("BridgeContractChainHeight")

	// ParamsStoreKeyBridgeContractChainID stores the bridge chain id
	ParamsStoreKeyBridgeContractChainID = []byte("BridgeChainID")

	// ParamsStoreKeyCosmosCoinDenom stores native cosmos coin denom
	ParamsStoreKeyCosmosCoinDenom = []byte("CosmosCoinDenom")

	// ParamsStoreKeyCosmosCoinErc20Contract store L1 erc20 contract address of cosmos native coin
	ParamsStoreKeyCosmosCoinErc20Contract = []byte("CosmosCoinErc20Contract")

	// ParamsStoreKeySignedValsetsWindow stores the signed blocks window
	ParamsStoreKeySignedValsetsWindow = []byte("SignedValsetsWindow")

	// ParamsStoreKeySignedBatchesWindow stores the signed blocks window
	ParamsStoreKeySignedBatchesWindow = []byte("SignedBatchesWindow")

	// ParamsStoreKeySignedClaimsWindow stores the signed blocks window
	ParamsStoreKeySignedClaimsWindow = []byte("SignedClaimsWindow")

	// ParamsStoreKeyTargetBatchTimeout stores
	ParamsStoreKeyTargetBatchTimeout = []byte("TargetBatchTimeout")

	// ParamsStoreKeyAverageBlockTime stores the average block time of the Injective Chain in milliseconds
	ParamsStoreKeyAverageBlockTime = []byte("AverageBlockTime")

	// ParamsStoreKeyAverageEthereumBlockTime stores the average block time of Ethereum in milliseconds
	ParamsStoreKeyAverageEthereumBlockTime = []byte("AverageEthereumBlockTime")

	// ParamsStoreSlashFractionValset stores the slash fraction valset
	ParamsStoreSlashFractionValset = []byte("SlashFractionValset")

	// ParamsStoreSlashFractionBatch stores the slash fraction Batch
	ParamsStoreSlashFractionBatch = []byte("SlashFractionBatch")

	// ParamsStoreSlashFractionClaim stores the slash fraction Claim
	ParamsStoreSlashFractionClaim = []byte("SlashFractionClaim")

	// ParamsStoreSlashFractionConflictingClaim stores the slash fraction ConflictingClaim
	ParamsStoreSlashFractionConflictingClaim = []byte("SlashFractionConflictingClaim")

	// ParamStoreUnbondSlashingValsetsWindow stores unbond slashing valset window
	ParamStoreUnbondSlashingValsetsWindow = []byte("UnbondSlashingValsetsWindow")

	// ParamStoreClaimSlashingEnabled stores ClaimSlashing is enabled or not
	ParamStoreClaimSlashingEnabled = []byte("ClaimSlashingEnabled")

	// ParamStoreSlashFractionBadEthSignature stores the amount by which a validator making a fraudulent eth signature will be slashed
	ParamStoreSlashFractionBadEthSignature = []byte("SlashFractionBadEthSignature")

	// ParamStoreValsetRewardAmount is the amount of the coin, both denom and amount to issue
	// to a relayer when they relay a valset
	ParamStoreValsetRewardAmount = []byte("ValsetReward")

	// Ensure that params implements the proper interface
	_ paramtypes.ParamSet = &Params{}
)

// ValidateBasic validates genesis state by looping through the params and
// calling their validation functions
func (s GenesisState) ValidateBasic() error {
	if err := s.Params.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "params")
	}
	return nil
}

// DefaultGenesisState returns empty genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// DefaultParams returns a copy of the default params
func DefaultParams() *Params {
	return &Params{
		PeggyId:                       "injective-peggyid",
		SignedValsetsWindow:           10000,
		SignedBatchesWindow:           10000,
		SignedClaimsWindow:            10000,
		TargetBatchTimeout:            43200000,
		AverageBlockTime:              5000,
		AverageEthereumBlockTime:      15000,
		SlashFractionValset:           sdk.NewDec(1).Quo(sdk.NewDec(1000)),
		SlashFractionBatch:            sdk.NewDec(1).Quo(sdk.NewDec(1000)),
		SlashFractionClaim:            sdk.NewDec(1).Quo(sdk.NewDec(1000)),
		SlashFractionConflictingClaim: sdk.NewDec(1).Quo(sdk.NewDec(1000)),
		SlashFractionBadEthSignature:  sdk.NewDec(1).Quo(sdk.NewDec(1000)),
		CosmosCoinDenom:               "inj",
		UnbondSlashingValsetsWindow:   10000,
		ClaimSlashingEnabled:          false,
	}
}

// ValidateBasic checks that the parameters have valid values.
func (p Params) ValidateBasic() error {
	if err := validatePeggyID(p.PeggyId); err != nil {
		return sdkerrors.Wrap(err, "peggy id")
	}
	if err := validateContractHash(p.ContractSourceHash); err != nil {
		return sdkerrors.Wrap(err, "contract hash")
	}
	if err := validateBridgeContractAddress(p.BridgeEthereumAddress); err != nil {
		return sdkerrors.Wrap(err, "bridge contract address")
	}
	if err := validateBridgeContractStartHeight(p.BridgeContractStartHeight); err != nil {
		return sdkerrors.Wrap(err, "bridge contract start height")
	}
	if err := validateBridgeChainID(p.BridgeChainId); err != nil {
		return sdkerrors.Wrap(err, "bridge chain id")
	}
	if err := validateCosmosCoinDenom(p.CosmosCoinDenom); err != nil {
		return sdkerrors.Wrap(err, "cosmos coin denom")
	}
	if err := validateCosmosCoinErc20Contract(p.CosmosCoinErc20Contract); err != nil {
		return sdkerrors.Wrap(err, "cosmos coin erc20 contract address")
	}
	if err := validateTargetBatchTimeout(p.TargetBatchTimeout); err != nil {
		return sdkerrors.Wrap(err, "Batch timeout")
	}
	if err := validateAverageBlockTime(p.AverageBlockTime); err != nil {
		return sdkerrors.Wrap(err, "Block time")
	}
	if err := validateAverageEthereumBlockTime(p.AverageEthereumBlockTime); err != nil {
		return sdkerrors.Wrap(err, "Ethereum block time")
	}
	if err := validateSignedValsetsWindow(p.SignedValsetsWindow); err != nil {
		return sdkerrors.Wrap(err, "signed blocks window")
	}
	if err := validateSignedBatchesWindow(p.SignedBatchesWindow); err != nil {
		return sdkerrors.Wrap(err, "signed blocks window")
	}
	if err := validateSignedClaimsWindow(p.SignedClaimsWindow); err != nil {
		return sdkerrors.Wrap(err, "signed blocks window")
	}
	if err := validateSlashFractionValset(p.SlashFractionValset); err != nil {
		return sdkerrors.Wrap(err, "slash fraction valset")
	}
	if err := validateSlashFractionBatch(p.SlashFractionBatch); err != nil {
		return sdkerrors.Wrap(err, "slash fraction valset")
	}
	if err := validateSlashFractionClaim(p.SlashFractionClaim); err != nil {
		return sdkerrors.Wrap(err, "slash fraction valset")
	}
	if err := validateSlashFractionConflictingClaim(p.SlashFractionConflictingClaim); err != nil {
		return sdkerrors.Wrap(err, "slash fraction valset")
	}
	if err := validateSlashFractionBadEthSignature(p.SlashFractionBadEthSignature); err != nil {
		return sdkerrors.Wrap(err, "slash fraction BadEthSignature")
	}
	if err := validateUnbondSlashingValsetsWindow(p.UnbondSlashingValsetsWindow); err != nil {
		return sdkerrors.Wrap(err, "unbond Slashing valset window")
	}
	if err := validateClaimSlashingEnabled(p.ClaimSlashingEnabled); err != nil {
		return sdkerrors.Wrap(err, "claim slashing enabled")
	}

	return nil
}

// ParamKeyTable for auth module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamsStoreKeyPeggyID, &p.PeggyId, validatePeggyID),
		paramtypes.NewParamSetPair(ParamsStoreKeyContractHash, &p.ContractSourceHash, validateContractHash),
		paramtypes.NewParamSetPair(ParamsStoreKeyBridgeContractAddress, &p.BridgeEthereumAddress, validateBridgeContractAddress),
		paramtypes.NewParamSetPair(ParamsStoreKeyBridgeContractChainID, &p.BridgeChainId, validateBridgeChainID),
		paramtypes.NewParamSetPair(ParamsStoreKeyCosmosCoinDenom, &p.CosmosCoinDenom, validateCosmosCoinDenom),
		paramtypes.NewParamSetPair(ParamsStoreKeyCosmosCoinErc20Contract, &p.CosmosCoinErc20Contract, validateCosmosCoinErc20Contract),
		paramtypes.NewParamSetPair(ParamsStoreKeySignedValsetsWindow, &p.SignedValsetsWindow, validateSignedValsetsWindow),
		paramtypes.NewParamSetPair(ParamsStoreKeySignedBatchesWindow, &p.SignedBatchesWindow, validateSignedBatchesWindow),
		paramtypes.NewParamSetPair(ParamsStoreKeySignedClaimsWindow, &p.SignedClaimsWindow, validateSignedClaimsWindow),
		paramtypes.NewParamSetPair(ParamsStoreKeyAverageBlockTime, &p.AverageBlockTime, validateAverageBlockTime),
		paramtypes.NewParamSetPair(ParamsStoreKeyTargetBatchTimeout, &p.TargetBatchTimeout, validateTargetBatchTimeout),
		paramtypes.NewParamSetPair(ParamsStoreKeyAverageEthereumBlockTime, &p.AverageEthereumBlockTime, validateAverageEthereumBlockTime),
		paramtypes.NewParamSetPair(ParamsStoreSlashFractionValset, &p.SlashFractionValset, validateSlashFractionValset),
		paramtypes.NewParamSetPair(ParamsStoreSlashFractionBatch, &p.SlashFractionBatch, validateSlashFractionBatch),
		paramtypes.NewParamSetPair(ParamsStoreSlashFractionClaim, &p.SlashFractionClaim, validateSlashFractionClaim),
		paramtypes.NewParamSetPair(ParamsStoreSlashFractionConflictingClaim, &p.SlashFractionConflictingClaim, validateSlashFractionConflictingClaim),
		paramtypes.NewParamSetPair(ParamStoreSlashFractionBadEthSignature, &p.SlashFractionBadEthSignature, validateSlashFractionBadEthSignature),
		paramtypes.NewParamSetPair(ParamStoreUnbondSlashingValsetsWindow, &p.UnbondSlashingValsetsWindow, validateUnbondSlashingValsetsWindow),
		paramtypes.NewParamSetPair(ParamStoreClaimSlashingEnabled, &p.ClaimSlashingEnabled, validateClaimSlashingEnabled),
		paramtypes.NewParamSetPair(ParamsStoreKeyBridgeContractStartHeight, &p.BridgeContractStartHeight, validateBridgeContractStartHeight),
		paramtypes.NewParamSetPair(ParamStoreValsetRewardAmount, &p.ValsetReward, validateValsetReward),
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

func validatePeggyID(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if _, err := strToFixByteArray(v); err != nil {
		return err
	}
	return nil
}

func validateContractHash(i interface{}) error {
	if _, ok := i.(string); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBridgeChainID(i interface{}) error {
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBridgeContractStartHeight(i interface{}) error {
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateTargetBatchTimeout(i interface{}) error {
	val, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	} else if val < 60000 {
		return fmt.Errorf("invalid target batch timeout, less than 60 seconds is too short")
	}
	return nil
}

func validateAverageBlockTime(i interface{}) error {
	val, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	} else if val < 100 {
		return fmt.Errorf("invalid average Cosmos block time, too short for latency limitations")
	}
	return nil
}

func validateAverageEthereumBlockTime(i interface{}) error {
	val, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	} else if val < 100 {
		return fmt.Errorf("invalid average Ethereum block time, too short for latency limitations")
	}
	return nil
}

func validateBridgeContractAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := ValidateEthAddress(v); err != nil {
		if !strings.Contains(err.Error(), "empty") {
			return err
		}
	}
	return nil
}

func validateSignedValsetsWindow(i interface{}) error {
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateUnbondSlashingValsetsWindow(i interface{}) error {
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionValset(i interface{}) error {
	if _, ok := i.(sdk.Dec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSignedBatchesWindow(i interface{}) error {
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSignedClaimsWindow(i interface{}) error {
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionBatch(i interface{}) error {
	if _, ok := i.(sdk.Dec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionClaim(i interface{}) error {
	if _, ok := i.(sdk.Dec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionConflictingClaim(i interface{}) error {
	if _, ok := i.(sdk.Dec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func strToFixByteArray(s string) ([32]byte, error) {
	var out [32]byte
	if len([]byte(s)) > 32 {
		return out, fmt.Errorf("string too long")
	}
	copy(out[:], s)
	return out, nil
}

func validateCosmosCoinDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if _, err := strToFixByteArray(v); err != nil {
		return err
	}
	return nil
}

func validateCosmosCoinErc20Contract(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// empty address is valid
	if v == "" {
		return nil
	}

	return ValidateEthAddress(v)
}

func validateClaimSlashingEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionBadEthSignature(i interface{}) error {
	if _, ok := i.(sdk.Dec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateValsetReward(i interface{}) error {
	return nil
}
