package tx

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/stretchr/testify/assert"
)

func TestClientSetters(t *testing.T) {
	f := tx.Factory{}
	f = f.WithSequence(2)

	c := Client{}
	c.txFactory = &f
	c.SetAccSeq(30)
	assert.Equal(t, 30, int(c.txFactory.Sequence()))
}
