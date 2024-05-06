package store

type Unmarshalable interface {
	Unmarshal(data []byte) error
}

type PtrUnmarshalable[T any] interface {
	Unmarshalable
	*T
}

type Marshalable interface {
	Marshal() ([]byte, error)
}

type PtrMarshalable[T any] interface {
	Marshalable
	*T
}

type BinMarshalable interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
}

type PtrBinMarshalable[T any] interface {
	BinMarshalable
	*T
}

type Integer interface {
	~int32 | ~int64 | ~uint32 | ~uint64 | byte
}

type KV[K any, V any] struct {
	Key K
	Val V
}
