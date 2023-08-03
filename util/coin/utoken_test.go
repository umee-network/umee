package coin

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestToUTokenDenom(t *testing.T) {
	// Turns base token denoms into base uTokens
	assert.Equal(t, "u/uumee", ToUTokenDenom("uumee"))
	assert.Equal(t, "u/ibc/abcd", ToUTokenDenom("ibc/abcd"))

	// Empty return for uTokens
	assert.Equal(t, "", ToUTokenDenom("u/uumee"))
	assert.Equal(t, "", ToUTokenDenom("u/ibc/abcd"))

	// Edge cases
	assert.Equal(t, "u/", ToUTokenDenom(""))
}

func TestToTokenDenom(t *testing.T) {
	// Turns uToken denoms into base tokens
	assert.Equal(t, "uumee", StripUTokenDenom("u/uumee"))
	assert.Equal(t, "ibc/abcd", StripUTokenDenom("u/ibc/abcd"))

	// Empty return for base tokens
	assert.Equal(t, "", StripUTokenDenom("uumee"))
	assert.Equal(t, "", StripUTokenDenom("ibc/abcd"))

	// Empty return on repreated prefix
	assert.Equal(t, "", StripUTokenDenom("u/u/abcd"))

	// Edge cases
	assert.Equal(t, "", StripUTokenDenom("u/"))
	assert.Equal(t, "", StripUTokenDenom(""))
}
