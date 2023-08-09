package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ugov "github.com/umee-network/umee/v5/x/ugov"
)

// NewParamsBuilder creates a ParamsKeeper builder function
func NewParamsBuilder(pk ugov.ParamsKeeper) ugov.ParamsKeeperBuilder {
	return func(_ *sdk.Context) ugov.ParamsKeeper {
		return pk
	}
}

// NewEmergencyGroupBuilder creates a  WithEmergencyGroup builder function
func NewEmergencyGroupBuilder(pk ugov.WithEmergencyGroup) ugov.EmergencyGroupBuilder {
	return func(_ *sdk.Context) ugov.WithEmergencyGroup {
		return pk
	}
}

// NewSimpleEmergencyGroupBuilder creates a EmergencyGroupBuilder builder function
func NewSimpleEmergencyGroupBuilder() ugov.EmergencyGroupBuilder {
	return func(_ *sdk.Context) ugov.WithEmergencyGroup {
		return NewSimpleEmergencyGroup()
	}
}
