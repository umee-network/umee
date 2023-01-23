package types

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q QueryMaxWithdraw) ValidateBasic() error {
	if q.Address == "" {
		return status.Error(codes.InvalidArgument, "empty address")
	}
	// if denom is empty, then we assume all possible denoms
	if q.Denom != "" {
		return ValidateBaseDenom(q.Denom)
	}
	return nil
}
