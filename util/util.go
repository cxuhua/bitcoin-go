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

func String(b []byte) string {
	for idx, c := range b {
		if c == 0 {
			return string(b[:idx])
		}
	}
	return string(b)
}
