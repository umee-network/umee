package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/x/leverage/types"
)

func TestUTokenFromTokenDenom(t *testing.T) {
	require.Equal(t, "u/uumee", types.UTokenFromTokenDenom("uumee"))
}
