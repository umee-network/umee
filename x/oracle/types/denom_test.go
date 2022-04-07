package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDenomString(t *testing.T) {
	denom := Denom{
		BaseDenom:   "uumee",
		SymbolDenom: "umee",
		Exponent:    16,
	}
	require.Equal(
		t,
		"base_denom: uumee\nsymbol_denom: umee\nexponent: 16\n",
		denom.String(),
	)
}

func TestDenomEqual(t *testing.T) {
	umee := Denom{
		BaseDenom:   "uumee",
		SymbolDenom: "umee",
		Exponent:    16,
	}
	clone := Denom{
		BaseDenom:   "uumee",
		SymbolDenom: "umee",
		Exponent:    16,
	}
	similar := Denom{
		BaseDenom:   "uumee",
		SymbolDenom: "atom",
		Exponent:    16,
	}
	atom := Denom{
		BaseDenom:   "uatom",
		SymbolDenom: "atom",
		Exponent:    8,
	}
	require.True(t, umee.Equal(&clone))
	require.False(t, umee.Equal(&similar))
	require.False(t, umee.Equal(&atom))
}

func TestDenomListString(t *testing.T) {
	denoms := DenomList{
		Denom{
			BaseDenom:   "uumee",
			SymbolDenom: "umee",
			Exponent:    16,
		},
		Denom{
			BaseDenom:   "uatom",
			SymbolDenom: "atom",
			Exponent:    8,
		},
	}
	require.Equal(
		t,
		"base_denom: uumee\nsymbol_denom: umee\nexponent: 16\n\nbase_denom: uatom\nsymbol_denom: atom\nexponent: 8",
		denoms.String(),
	)
}

func TestDenomListContains(t *testing.T) {
	denoms := DenomList{
		Denom{
			BaseDenom:   "uumee",
			SymbolDenom: "umee",
			Exponent:    16,
		},
		Denom{
			BaseDenom:   "uatom",
			SymbolDenom: "atom",
			Exponent:    8,
		},
	}
	require.True(t, denoms.Contains("atom"))
	require.True(t, denoms.Contains("umee"))
	require.False(t, denoms.Contains("foo"))
}
