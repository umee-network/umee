package uics20

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrIs(t *testing.T) {
	err := errMemoValidation{errWrongSigner}

	assert.True(t, errors.Is(err, errMemoValidation{}))
}
