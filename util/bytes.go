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

// UintWithNullPrefix efficiently serializes uint using LittleEndian and
// prepends zero byte (null prefix).
func UintWithNullPrefix(n uint64) []byte {
	bz := make([]byte, 9)
	binary.LittleEndian.PutUint64(bz[1:], n)
	return bz
}

// KeyWithUint32 concatenates prefix big endian serialized n value.
// No zero byte is appended at the end.
func KeyWithUint32(prefix []byte, n uint32) []byte {
	out := make([]byte, len(prefix)+4)
	copy(out, prefix)
	binary.BigEndian.PutUint32(out[len(prefix):], n)
	return out
}

// KeyWithUint64 concatenates prefix big endian serialized n value.
// No zero byte is appended at the end.
func KeyWithUint64(prefix []byte, n uint64) []byte {
	out := make([]byte, len(prefix)+8)
	copy(out, prefix)
	binary.BigEndian.PutUint64(out[len(prefix):], n)
	return out
}
