package keeper_test

import (
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// StakingKeeperMock is a mock staking keeper for use in the tests.
type StakingKeeperMock struct {
	BondedValidators []stakingtypes.Validator
	ValidatorPower   map[string]int64
}

func NewStakingKeeperMock(operators ...sdk.ValAddress) *StakingKeeperMock {
	r := &StakingKeeperMock{
		BondedValidators: make([]stakingtypes.Validator, 0),
		ValidatorPower:   make(map[string]int64),
	}

	for _, a := range operators {
		r.BondedValidators = append(r.BondedValidators, stakingtypes.Validator{
			ConsensusPubkey: codectypes.UnsafePackAny(ed25519.GenPrivKey().PubKey()),
			OperatorAddress: a.String(),
			Status:          stakingtypes.Bonded,
		})

		r.ValidatorPower[a.String()] = 100
	}

	return r
}

func (s *StakingKeeperMock) GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator {
	return s.BondedValidators
}

func (s *StakingKeeperMock) GetLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress) int64 {
	v, ok := s.ValidatorPower[operator.String()]
	if !ok {
		panic("unknown address")
	}
	return v
}

func (s *StakingKeeperMock) GetLastTotalPower(ctx sdk.Context) (power sdk.Int) {
	var total int64
	for _, v := range s.ValidatorPower {
		total += v
	}
	return sdk.NewInt(total)
}

func (s *StakingKeeperMock) IterateValidators(ctx sdk.Context, cb func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
	for i, val := range s.BondedValidators {
		stop := cb(int64(i), val)
		if stop {
			break
		}
	}
}

func (s *StakingKeeperMock) IterateBondedValidatorsByPower(ctx sdk.Context, cb func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
	for i, val := range s.BondedValidators {
		stop := cb(int64(i), val)
		if stop {
			break
		}
	}
}

func (s *StakingKeeperMock) IterateLastValidators(ctx sdk.Context, cb func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
	for i, val := range s.BondedValidators {
		stop := cb(int64(i), val)
		if stop {
			break
		}
	}
}

func (s *StakingKeeperMock) Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
	for _, val := range s.BondedValidators {
		if val.GetOperator().Equals(addr) {
			return val
		}
	}

	return nil
}

func (s *StakingKeeperMock) ValidatorByConsAddr(ctx sdk.Context, addr sdk.ConsAddress) stakingtypes.ValidatorI {
	for _, val := range s.BondedValidators {
		cons, err := val.GetConsAddr()
		if err != nil {
			panic(err)
		}

		if cons.Equals(addr) {
			return val
		}
	}

	return nil
}

func (s *StakingKeeperMock) GetParams(ctx sdk.Context) stakingtypes.Params {
	return stakingtypes.DefaultParams()
}

func (s *StakingKeeperMock) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool) {
	panic("unexpected call")
}

func (s *StakingKeeperMock) PowerReduction(ctx sdk.Context) (res sdk.Int) {
	return sdk.DefaultPowerReduction
}

func (s *StakingKeeperMock) ValidatorQueueIterator(ctx sdk.Context, endTime time.Time, endHeight int64) sdk.Iterator {
	store := ctx.KVStore(sdk.NewKVStoreKey("staking"))
	return store.Iterator(stakingtypes.ValidatorQueueKey, sdk.InclusiveEndBytes(stakingtypes.GetValidatorQueueKey(endTime, endHeight)))
}

func (s *StakingKeeperMock) Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec) {}
func (s *StakingKeeperMock) Jail(sdk.Context, sdk.ConsAddress)                         {}
