package genmap

import "testing"
import "github.com/stretchr/testify/require"

func TestPick(t *testing.T) {
	require := require.New(t)
	m := map[string]int{
		"one": 1, "two": 2, "three": 3}

	m2 := Pick(m, []string{"one"})
	require.Equal(map[string]int{"one": 1}, m2)

	m2 = Pick(m, []string{"two"})
	require.Equal(map[string]int{"two": 2}, m2)

	m2 = Pick(m, []string{"one", "three"})
	require.Equal(map[string]int{"one": 1, "three": 3}, m2)

	m2 = Pick(m, []string{})
	require.Equal(map[string]int{}, m2)

	m2 = Pick(m, []string{"other"})
	require.Equal(map[string]int{}, m2)
}
