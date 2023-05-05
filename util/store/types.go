package store

import "golang.org/x/exp/constraints"

type Marshalable interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
}

type PtrMarshalable[T any] interface {
	Marshalable
	*T
}

type gogoInteger[T any, Num constraints.Integer] interface {
	PtrMarshalable[T]
	GetValue() Num
}
