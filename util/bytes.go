package util

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
