package store

type Marshalable interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
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
