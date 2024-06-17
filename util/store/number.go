package store

import (
	"encoding/binary"
)

type Uint32 uint32
type Uint64 uint64

func (o *Uint32) Unmarshal(bz []byte) error {
	*o = Uint32(binary.BigEndian.Uint32(bz))
	return nil
}

func (o *Uint64) Unmarshal(bz []byte) error {
	*o = Uint64(binary.BigEndian.Uint64(bz))
	return nil
}
