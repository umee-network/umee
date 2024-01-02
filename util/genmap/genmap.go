package genmap

// Pick creates a new map based on `m` by selecting entries from `keys`.
// If a key is in keys, but not map, then it's not included.
// Map keys and values are copied.
func Pick[K comparable, V any](m map[K]V, keys []K) map[K]V {
	picked := make(map[K]V)
	for i := range keys {
		if v, ok := m[keys[i]]; ok {
			picked[keys[i]] = v
		}
	}
	return picked
}

// MapValues constructs a list of all values of the map in a non-deterministic order.
// NOTE: when used in the protocol functions, the values must be sorted.
func MapValues[K comparable, V any](m map[K]V) []V {
	ls := make([]V, len(m))
	i := 0
	for _, o := range m {
		ls[i] = o
		i++
	}
	return ls
}
