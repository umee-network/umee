package tsdk_test

import (
	"testing"

	"github.com/umee-network/umee/v6/tests/tsdk"
	"gotest.tools/v3/assert"
)

// TestGenerateString checks the randomness and length properties of the generated string.
func TestGenerateString(t *testing.T) {
	length := 10
	str := tsdk.GenerateString(uint(length))
	assert.Equal(t, len(str), length, "Generated string length should match the input length")
}
