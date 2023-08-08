package checkers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umee-network/umee/v5/tests/tsdk"
)

func TestNumberDiff(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(IntegerMaxDiff(1, 1, 0, ""))
	assert.NoError(IntegerMaxDiff(-1, -1, 0, ""))
	assert.NoError(IntegerMaxDiff(0, 0, 0, ""))

	assert.Error(IntegerMaxDiff(1, -1, 0, ""))
	assert.Error(IntegerMaxDiff(1, -1, -1, ""))
	assert.Error(IntegerMaxDiff(-1, 1, 0, ""))
	assert.Error(IntegerMaxDiff(-1, 1, -1, ""))

	assert.NoError(IntegerMaxDiff(1, -1, 2, ""))
	assert.NoError(IntegerMaxDiff(1, -1, 3, ""))
	assert.NoError(IntegerMaxDiff(1, -1, 100, ""))
	assert.NoError(IntegerMaxDiff(1, -1, 60000000, ""))

	assert.NoError(IntegerMaxDiff(-1, 1, 2, ""))
	assert.NoError(IntegerMaxDiff(-1, 1, 3, ""))
	assert.NoError(IntegerMaxDiff(-1, 1, 100, ""))
	assert.NoError(IntegerMaxDiff(-1, 1, 60000000, ""))
}

func TestDecDiff(t *testing.T) {
	assert := assert.New(t)
	decMaxDiff := func(a, b, maxDiff float64) error {
		return DecMaxDiff(tsdk.DecF(a), tsdk.DecF(b), tsdk.DecF(maxDiff), "")
	}

	assert.NoError(decMaxDiff(1, 1, 0))
	assert.NoError(decMaxDiff(-1, -1, 0))
	assert.NoError(decMaxDiff(0, 0, 0))

	assert.NoError(decMaxDiff(0.0001, 0.0001, 0))
	assert.NoError(decMaxDiff(-0.0001, -0.0001, 0))

	assert.Error(decMaxDiff(1, -1, 0))
	assert.Error(decMaxDiff(1, -1, -1))
	assert.Error(decMaxDiff(-1, 1, 0))
	assert.Error(decMaxDiff(-1, 1, -1))

	assert.NoError(decMaxDiff(1, -1, 2))
	assert.NoError(decMaxDiff(1, -1, 3))
	assert.NoError(decMaxDiff(1, -1, 100))
	assert.NoError(decMaxDiff(1, -1, 60000000))

	assert.NoError(decMaxDiff(-1, 1, 2))
	assert.NoError(decMaxDiff(-1, 1, 3))
	assert.NoError(decMaxDiff(-1, 1, 100))
	assert.NoError(decMaxDiff(-1, 1, 60000000))
}
