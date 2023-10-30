package itest

import (
	"cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"
)

type TestTransaction struct {
	Name        string
	Command     *cobra.Command
	Args        []string
	ExpectedErr *errors.Error
}

type TestQuery struct {
	Name    string
	Command *cobra.Command
	Args    []string
	// object to decode response into
	Response         proto.Message
	ExpectedResponse proto.Message
	ErrMsg           string
}
