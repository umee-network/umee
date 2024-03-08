package uics20

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrMemoValidation(t *testing.T) {
	err := errMemoValidation{errWrongSigner}

	assert.True(t, errors.Is(err, errMemoValidation{}))
	assert.True(t, errors.Is(errMemoValidation{}, err))
	assert.False(t, errors.Is(errWrongSigner, err))

	assert.Equal(t, errWrongSigner, errors.Unwrap(err))
	assert.Contains(t, err.Error(), errWrongSigner.Error())
}
