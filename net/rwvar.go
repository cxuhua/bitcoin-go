package net

import (
	"bitcoin/script"
	"encoding/binary"
	"io"
)

var (
	ByteOrder = binary.LittleEndian
)

type MsgBuffer struct {
	Payload []byte //payload raw data
	io.ReadWriteSeeker
	rpos int
	wpos int
}

func NewMsgBuffer(b []byte) *MsgBuffer {
	m := &MsgBuffer{}
	m.Payload = b
	m.rpos = 0
	m.wpos = 0
	return m
}

func (m *MsgBuffer) IsEOF() bool {
	return m.rpos == len(m.Payload)
}

func (m *MsgBuffer) Read(p []byte) (n int, err error) {
	if m.rpos+len(p) > len(m.Payload) {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	copy(p, m.Payload[m.rpos:m.rpos+len(p)])
	m.rpos += len(p)
	return len(p), nil
}

func (m *MsgBuffer) Write(p []byte) (n int, err error) {
	l := len(p)
	if l == 0 {
		return 0, nil
	}
	m.Payload = append(m.Payload, p...)
	m.wpos += l
	return l, nil
}

func (m *MsgBuffer) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart {
		m.rpos = int(offset)
		m.wpos = int(offset)
	} else if whence == io.SeekCurrent {
		m.rpos += int(offset)
		m.wpos += int(offset)
	} else {
		m.rpos = len(m.Payload) + int(offset)
		m.wpos = len(m.Payload) + int(offset)
	}
	if m.rpos >= len(m.Payload) {
		return 0, io.EOF
	}
	return 0, nil
}

func (b *MsgBuffer) Skip(l int) {
	if b.rpos+l > len(b.Payload) {
		panic(io.EOF)
	}
	b.rpos += l
}

func (b *MsgBuffer) Peek(l int) []byte {
	if b.rpos+l > len(b.Payload) {
		panic(io.EOF)
	}
	return b.Payload[b.rpos : b.rpos+l]
}

func (b *MsgBuffer) Bytes() []byte {
	return b.Payload
}

//read var int
func (m *MsgBuffer) ReadVarInt() (uint64, int) {
	b := m.ReadUint8()
	if b < 0xFD {
		return uint64(b), 1
	} else if b == 0xFD {
		return uint64(m.ReadUInt16()), 3
	} else if b == 0xFE {
		return uint64(m.ReadUInt32()), 5
	} else if b == 0xFF {
		return m.ReadUInt64(), 9
	} else {
		return 0, -1
	}
}

func (m *MsgBuffer) WriteVarInt(v uint64) int {
	if v < 0xFD {
		m.WriteUint8(uint8(v & 0xFF))
		return 1
	} else if v <= 0xFFFF {
		m.WriteUint8(0xFD)
		m.WriteUInt16(uint16(v & 0xFFFFFF))
		return 3
	} else if v <= 0xFFFFFFFF {
		m.WriteUint8(0xFE)
		m.WriteUInt32(uint32(v & 0xFFFFFFFF))
		return 5
	} else {
		m.WriteUint8(0xFF)
		m.WriteUInt64(uint64(v & 0xFFFFFFFFFFFFFFFF))
		return 9
	}
}

func (m *MsgBuffer) ReadScript() *script.Script {
	l, _ := m.ReadVarInt()
	b := make([]byte, l)
	m.ReadBytes(b)
	return script.NewScript(b)
}

func (m *MsgBuffer) WriteScript(s *script.Script) {
	m.WriteVarInt(uint64((s.Len())))
	m.WriteBytes(s.Bytes())
}

func (m *MsgBuffer) ReadBytes(b []byte) {
	if len(b) == 0 {
		return
	}
	err := binary.Read(m, ByteOrder, b)
	if err != nil {
		panic(err)
	}
}

func (m *MsgBuffer) WriteBytes(b []byte) {
	if len(b) == 0 {
		return
	}
	err := binary.Write(m, ByteOrder, b)
	if err != nil {
		panic(err)
	}
}

func (m *MsgBuffer) ReadUint8() uint8 {
	v := uint8(0)
	err := binary.Read(m, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func (m *MsgBuffer) WriteUint8(v uint8) {
	err := binary.Write(m, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func (m *MsgBuffer) ReadUInt16() uint16 {
	v := uint16(0)
	err := binary.Read(m, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func (m *MsgBuffer) WriteUInt16(v uint16) {
	err := binary.Write(m, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func (m *MsgBuffer) ReadUInt32() uint32 {
	v := uint32(0)
	err := binary.Read(m, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func (m *MsgBuffer) WriteUInt32(v uint32) {
	err := binary.Write(m, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func (m *MsgBuffer) ReadUInt64() uint64 {
	v := uint64(0)
	err := binary.Read(m, ByteOrder, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func (m *MsgBuffer) WriteUInt64(v uint64) {
	err := binary.Write(m, ByteOrder, v)
	if err != nil {
		panic(err)
	}
}

func (m *MsgBuffer) ReadString() string {
	l, _ := m.ReadVarInt()
	if l == 0 {
		return ""
	}
	b := make([]byte, l)
	m.ReadBytes(b)
	return string(b)
}

func (m *MsgBuffer) WriteString(v string) {
	l := len(v)
	m.WriteVarInt(uint64(l))
	if l > 0 {
		m.WriteBytes([]byte(v))
	}
}
