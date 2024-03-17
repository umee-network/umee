package tcheckers

import (
	"testing"

	"gotest.tools/v3/assert"
)

func ErrorContains(t *testing.T, err error, expectedErr, testName string) {
	if expectedErr == "" {
		assert.NilError(t, err, testName)
	} else {
		assert.ErrorContains(t, err, expectedErr, testName)
	}
}
