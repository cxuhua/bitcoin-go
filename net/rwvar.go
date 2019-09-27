package net

import (
	"bitcoin/script"
	"encoding/binary"
	"io"
)

var (
	ByteOrder = binary.LittleEndian
)

func a() (int, int) {
	return 1, 2
}

//read var int
func ReadVarInt(r io.Reader) (uint64, int) {
	b := ReadUint8(r)
	if b < 0xFD {
		return uint64(b), 1
	} else if b == 0xFD {
		return uint64(ReadUInt16(r)), 3
	} else if b == 0xFE {
		return uint64(ReadUInt32(r)), 5
	} else if b == 0xFF {
		return ReadUInt64(r), 9
	} else {
		return 0, -1
	}
}

func WriteVarInt(w io.Writer, v uint64) int {
	if v < 0xFD {
		WriteUint8(w, uint8(v&0xFF))
		return 1
	} else if v <= 0xFFFF {
		WriteUint8(w, 0xFD)
		WriteUInt16(w, uint16(v&0xFFFFFF))
		return 3
	} else if v <= 0xFFFFFFFF {
		WriteUint8(w, 0xFE)
		WriteUInt32(w, uint32(v&0xFFFFFFFF))
		return 5
	} else {
		WriteUint8(w, 0xFF)
		WriteUInt64(w, uint64(v&0xFFFFFFFFFFFFFFFF))
		return 9
	}
}

func ReadScript(r io.Reader) *script.Script {
	l, _ := ReadVarInt(r)
	b := make([]byte, l)
	ReadBytes(r, b)
	return script.NewScript(b)
}

func WriteScript(w io.Writer, s *script.Script) {
	WriteVarInt(w, uint64((s.Len())))
	WriteBytes(w, s.Bytes())
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

func ReadUInt16(r io.Reader) uint16 {
	v := uint16(0)
	err := binary.Read(r, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUInt16(w io.Writer, v uint16) {
	err := binary.Write(w, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func ReadUInt32(r io.Reader) uint32 {
	v := uint32(0)
	err := binary.Read(r, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUInt32(w io.Writer, v uint32) {
	err := binary.Write(w, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func ReadUInt64(r io.Reader) uint64 {
	v := uint64(0)
	err := binary.Read(r, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUInt64(w io.Writer, v uint64) {
	err := binary.Write(w, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func ReadString(r io.Reader) string {
	l, _ := ReadVarInt(r)
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
