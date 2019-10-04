package net

import (
	"bitcoin/script"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ByteOrder = binary.LittleEndian
)

const (
	MSG_BUFFER_READ  = 0x1
	MSG_BUFFER_WRITE = 0x2
	MSG_BUFFER_RW    = MSG_BUFFER_READ | MSG_BUFFER_WRITE
	MSG_BUFFER_MAX   = MAX_BLOCK_SERIALIZED_SIZE
)

type MsgBuffer struct {
	Payload []byte //payload raw data
	io.ReadWriteSeeker
	rwpos  int
	rwflag int
}

func NewMsgBuffer(b []byte, rw int) *MsgBuffer {
	if rw == MSG_BUFFER_RW {
		panic(errors.New("msgbuffer not support rw"))
	}
	if len(b) > int(MSG_BUFFER_MAX) {
		panic(SizeError)
	}
	m := &MsgBuffer{}
	m.Payload = b
	m.rwpos = 0
	m.rwflag = rw
	return m
}

func (m *MsgBuffer) IsRead() bool {
	return m.rwflag&MSG_BUFFER_READ != 0
}

func (m *MsgBuffer) IsWrite() bool {
	return m.rwflag&MSG_BUFFER_WRITE != 0
}

func (m *MsgBuffer) Pos() int {
	return m.rwpos
}

func (m *MsgBuffer) IsEOF() bool {
	return m.rwpos == len(m.Payload)
}

func (m *MsgBuffer) Read(p []byte) (n int, err error) {
	if m.rwpos+len(p) > len(m.Payload) {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	copy(p, m.Payload[m.rwpos:m.rwpos+len(p)])
	m.rwpos += len(p)
	return len(p), nil
}

func (m *MsgBuffer) Write(p []byte) (n int, err error) {
	l := len(p)
	if l == 0 {
		return 0, nil
	}
	m.Payload = append(m.Payload, p...)
	m.rwpos += l
	return l, nil
}

func (m *MsgBuffer) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart {
		m.rwpos = int(offset)
		m.rwpos = int(offset)
	} else if whence == io.SeekCurrent {
		m.rwpos += int(offset)
		m.rwpos += int(offset)
	} else {
		m.rwpos = len(m.Payload) + int(offset)
		m.rwpos = len(m.Payload) + int(offset)
	}
	if m.rwpos >= len(m.Payload) {
		return 0, io.EOF
	}
	return 0, nil
}

func (b *MsgBuffer) Skip(l int) {
	if b.rwpos+l > len(b.Payload) {
		panic(io.EOF)
	}
	b.rwpos += l
}

func (b *MsgBuffer) Peek(l int) []byte {
	if b.rwpos+l > len(b.Payload) {
		panic(io.EOF)
	}
	return b.Payload[b.rwpos : b.rwpos+l]
}

func (b *MsgBuffer) SubBytes(s, e int) []byte {
	return b.Payload[s:e]
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
	if s == nil {
		m.WriteVarInt(0)
	} else {
		m.WriteVarInt(uint64((s.Len())))
		m.WriteBytes(s.Bytes())
	}
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
