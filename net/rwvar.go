package net

import (
	"encoding/binary"
	"io"
)

var (
	ByteOrder = binary.LittleEndian
)

//read var int
func ReadVarInt(r io.Reader, v ...*int) uint64 {
	n := uint64(0)
	for {
		b := ReadUint8(r)
		if len(v) > 0 {
			*v[0]++
		}
		n = (n << 7) | uint64(b&0x7F)
		if b&0x80 != 0 {
			n++
		} else {
			return n
		}
	}
}

func WriteVarInt(w io.Writer, v uint64) int {
	b := make([]byte, 10)
	l := 0
	for {
		if l > 0 {
			b[l] = byte(v&0x7F) | 0x80
		} else {
			b[l] = byte(v&0x7F) | 0x00
		}
		if v <= 0x7F {
			break
		}
		v = (v >> 7) - 1
		l++
	}
	for i := l; i >= 0; i-- {
		WriteUint8(w, b[i])
	}
	return l + 1
}

func ReadBytes(r io.Reader, b []byte) {
	if len(b) == 0 {
		return
	}
	err := binary.Read(r, ByteOrder, b)
	if err != nil {
		panic(err)
	}
}

func WriteBytes(w io.Writer, b []byte) {
	if len(b) == 0 {
		return
	}
	err := binary.Write(w, ByteOrder, b)
	if err != nil {
		panic(err)
	}
}

func ReadUint8(r io.Reader) uint8 {
	v := uint8(0)
	err := binary.Read(r, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUint8(w io.Writer, v uint8) {
	err := binary.Write(w, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func ReadUint16(r io.Reader) uint16 {
	v := uint16(0)
	err := binary.Read(r, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUint16(w io.Writer, v uint16) {
	err := binary.Write(w, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func ReadUint32(r io.Reader) uint32 {
	v := uint32(0)
	err := binary.Read(r, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUint32(w io.Writer, v uint32) {
	err := binary.Write(w, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func ReadUint64(r io.Reader) uint64 {
	v := uint64(0)
	err := binary.Read(r, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUint64(w io.Writer, v uint64) {
	err := binary.Write(w, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func ReadString(r io.Reader) string {
	l := ReadVarInt(r)
	if l == 0 {
		return ""
	}
	b := make([]byte, l)
	ReadBytes(r, b)
	return string(b)
}

func WriteString(w io.Writer, v string) {
	l := len(v)
	WriteVarInt(w, uint64(l))
	if l > 0 {
		WriteBytes(w, []byte(v))
	}
}
