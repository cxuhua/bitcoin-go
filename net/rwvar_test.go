package net

import (
	"bytes"
	"testing"
)

func TestRWVarInt(t *testing.T) {
	buf := &bytes.Buffer{}
	vs := []uint64{0x7F, 0xFD, 0xFFFF, 0xFFFF1, 0xFFFFFFFF, 0xFFFFFFFF1, 0xFFFFFFFFFF1}
	for _, v1 := range vs {
		buf.Reset()
		l := WriteVarInt(buf, v1)
		v2 := ReadVarInt(buf)
		if v1 != v2 {
			t.Fatalf("test v1=%X v2=%X error,l=%d", v1, v2, l)
		}
	}
}
