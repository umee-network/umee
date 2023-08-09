package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/tests/accs"
	ugov "github.com/umee-network/umee/v5/x/ugov"
)

var _ ugov.WithEmergencyGroup = simpleEmergencyGroup{}

var SimpleEmergencyGroupAddr = accs.GenerateAddr("ugov emergency group")

type simpleEmergencyGroup struct {
	Addr sdk.AccAddress
}

func (s simpleEmergencyGroup) EmergencyGroup() sdk.AccAddress {
	return s.Addr
}

func NewSimpleEmergencyGroup() ugov.WithEmergencyGroup {
	return simpleEmergencyGroup{Addr: SimpleEmergencyGroupAddr}
}
