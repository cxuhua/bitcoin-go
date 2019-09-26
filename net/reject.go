package net

import (
	"io"
)

type MsgReject struct {
	Message string
	Code    uint8
	Reason  string
	Data    []byte
}

func (m *MsgReject) Command() string {
	return NMT_REJECT
}

func (m *MsgReject) Read(h *MessageHeader, r io.Reader) {
	m.Message = ReadString(r)
	m.Code = ReadUint8(r)
	m.Reason = ReadString(r)
	m.Data = make([]byte, 32)
	ReadBytes(r, m.Data)
}

func (m *MsgReject) Write(h *MessageHeader, w io.Writer) {
	WriteString(w, m.Message)
	WriteUint8(w, m.Code)
	WriteString(w, m.Reason)
	WriteBytes(w, m.Data)
}

func NewMsgReject() *MsgReject {
	return &MsgReject{}
}
