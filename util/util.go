package util

import (
	"crypto/rand"
	"encoding/binary"
)

func RandUInt64() uint64 {
	v := uint64(0)
	if err := binary.Read(rand.Reader, binary.LittleEndian, &v); err != nil {
		panic(err)
	}
	return v
}
