package incentive

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestValidateProgramStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		desc        string
		status      ProgramStatus
		expectError error
	}{
		{"upcoming status", ProgramStatusUpcoming, nil},
		{"ongoing status", ProgramStatusOngoing, nil},
		{"completed status", ProgramStatusCompleted, nil},
		{"invalid status", ProgramStatusCompleted + 1, ErrInvalidProgramStatus},
	}

	for _, c := range cases {
		err := c.status.Validate()
		if c.expectError == nil {
			assert.NilError(t, err, c.desc)
		} else {
			assert.ErrorContains(t, err, c.expectError.Error(), c.desc)
		}
	}
}
