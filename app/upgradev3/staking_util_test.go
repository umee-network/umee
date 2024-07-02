package upgradev3

import (
	"crypto/rand"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ StakingKeeper = &MockStakingKeeper{}

// MockStakingKeeper implements the StakingKeeper interface.
type MockStakingKeeper struct {
	validators []types.Validator
	params     types.Params
}

// GetValidator implements StakingKeeper
func (m *MockStakingKeeper) GetValidator(_ sdk.Context, addr sdk.ValAddress) (types.Validator, bool) {
	var (
		validator types.Validator
		found     bool
	)

	for _, v := range m.validators {
		if v.GetOperator() == addr.String() {
			found = true
			validator = v
			break
		}
	}

	return validator, found
}

// BeforeValidatorModified implements StakingKeeper
func (*MockStakingKeeper) BeforeValidatorModified(sdk.Context, sdk.ValAddress) error {
	return nil
}

// GetAllValidators implements StakingKeeper
func (m *MockStakingKeeper) GetAllValidators(sdk.Context) (validators []types.Validator) {
	return m.validators
}

// GetParams implements StakingKeeper
func (m *MockStakingKeeper) GetParams(sdk.Context) types.Params {
	return m.params
}

// SetParams implements StakingKeeper
func (m *MockStakingKeeper) SetParams(_ sdk.Context, params types.Params) {
	m.params = params
}

// SetValidator implements StakingKeeper
func (m *MockStakingKeeper) SetValidator(_ sdk.Context, validator types.Validator) {
	for index, v := range m.validators {
		if v.GetOperator() == validator.GetOperator() {
			m.validators[index] = validator
			break
		}
	}
}

// GenerateRandomTestCase
func GenerateRandomTestCase() ([]sdk.ValAddress, MockStakingKeeper) {
	mockValidators := []types.Validator{}

	var valAddrs []sdk.ValAddress
	randNum, _ := rand.Int(rand.Reader, big.NewInt(10000))
	numInputs := 10 + int((randNum.Int64() % 100))
	for i := 0; i < numInputs; i++ {
		pubKey := secp256k1.GenPrivKey().PubKey()
		valValAddr := sdk.ValAddress(pubKey.Address())
		mockValidator, _ := types.NewValidator(valValAddr.String(), pubKey, types.Description{})
		mockValidators = append(mockValidators, mockValidator)
	}

	// adding 0.01 to first validator
	val := mockValidators[0]
	val.Commission.Rate = sdkmath.LegacyMustNewDecFromStr("0.01")
	mockValidators[0] = val

	// adding more then minimumCommissionRate to validator 2
	val = mockValidators[1]
	val.Commission.Rate = types.DefaultMinCommissionRate.Add(sdkmath.LegacyMustNewDecFromStr("1"))
	mockValidators[1] = val

	for i := 0; i < 2; i++ {
		valAddr, _ := sdk.ValAddressFromBech32(mockValidators[i].GetOperator())
		valAddr = append(valAddr, valAddr...)
	}
	stakingKeeper := NewMockStakingKeeper(mockValidators)

	return valAddrs, stakingKeeper
}

func NewMockStakingKeeper(validators []types.Validator) MockStakingKeeper {
	return MockStakingKeeper{
		validators: validators,
		params:     types.DefaultParams(),
	}
}
