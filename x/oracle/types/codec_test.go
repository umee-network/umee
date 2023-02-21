package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"gotest.tools/v3/assert"
)

func TestRegisterInterfaces(t *testing.T) {
	registry := types.NewInterfaceRegistry()
	RegisterInterfaces(registry)
	assert.DeepEqual(t, registry.ListAllInterfaces(), []string([]string{}))
}
