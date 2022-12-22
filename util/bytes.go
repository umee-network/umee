package util

import (
	"encoding/binary"
)

// ConcatBytes creates a new slice by merging list of bytes and leaving empty amount of margin
// bytes at the end
func ConcatBytes(margin int, bzs ...[]byte) []byte {
	l := 0
	for _, bz := range bzs {
		l += len(bz)
	}
	out := make([]byte, l+margin)
	offset := 0
	for _, bz := range bzs {
		copy(out[offset:], bz)
		offset += len(bz)
	}
	return out
}

func UintWithNullPrefix(n uint64) []byte {
	bz := make([]byte, 9)
	binary.LittleEndian.PutUint64(bz[1:], n)
	return bz
}
