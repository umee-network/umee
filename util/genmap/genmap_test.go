package genmap

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestPick(t *testing.T) {
	m := map[string]int{
		"one": 1, "two": 2, "three": 3,
	}

	m2 := Pick(m, []string{"one"})
	assert.DeepEqual(t, map[string]int{"one": 1}, m2)

	m2 = Pick(m, []string{"two"})
	assert.DeepEqual(t, map[string]int{"two": 2}, m2)

	m2 = Pick(m, []string{"one", "three"})
	assert.DeepEqual(t, map[string]int{"one": 1, "three": 3}, m2)

	m2 = Pick(m, []string{})
	assert.DeepEqual(t, map[string]int{}, m2)

	m2 = Pick(m, []string{"other"})
	assert.DeepEqual(t, map[string]int{}, m2)
}
