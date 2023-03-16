package incentive

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestValidateTier(t *testing.T) {
	cases := []struct {
		desc             string
		tier             BondTier
		canBeUnspecified bool
		expectError      error
	}{
		{"valid unspecified tier", 0, true, nil},
		{"invalid unspecified tier", 0, false, ErrInvalidTier},
		{"short tier", BondTierShort, false, nil},
		{"middle tier", BondTierMiddle, false, nil},
		{"long tier", BondTierLong, false, nil},
		{"invalid tier", BondTierLong + 1, false, ErrInvalidTier},
	}

	for _, c := range cases {
		err := c.tier.Validate(c.canBeUnspecified)
		if c.expectError == nil {
			assert.NilError(t, err, c.desc)
		} else {
			assert.ErrorContains(t, err, c.expectError.Error(), c.desc)
		}
	}
}

func TestValidateProgramStatus(t *testing.T) {
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
