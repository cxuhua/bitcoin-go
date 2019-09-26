package net

import (
	"io"
)

type MsgSendCmpct struct {
	Inter uint8
	Ver   uint64
}

func (m *MsgSendCmpct) Command() string {
	return NMT_SENDCMPCT
}

func (m *MsgSendCmpct) Read(h *MessageHeader, r io.Reader) {
	m.Inter = ReadUint8(r)
	m.Ver = ReadUInt64(r)
}

func (m *MsgSendCmpct) Write(h *MessageHeader, w io.Writer) {
	WriteUint8(w, m.Inter)
	WriteUInt64(w, m.Ver)
}

func NewMsgSendCmpct() *MsgSendCmpct {
	return &MsgSendCmpct{}
}

//

type MsgFeeFilter struct {
	FeeRate uint64
}

func (m *MsgFeeFilter) Command() string {
	return NMT_FEEFILTER
}

func (m *MsgFeeFilter) Read(h *MessageHeader, r io.Reader) {
	m.FeeRate = ReadUInt64(r)
}

func (m *MsgFeeFilter) Write(h *MessageHeader, w io.Writer) {
	WriteUInt64(w, m.FeeRate)
}

func NewMsgFeeFilter() *MsgFeeFilter {
	return &MsgFeeFilter{}
}

//

type MsgSendHeaders struct {
}

func (m *MsgSendHeaders) Command() string {
	return NMT_SENDHEADERS
}

func (m *MsgSendHeaders) Read(h *MessageHeader, r io.Reader) {
	//no payload
}

func (m *MsgSendHeaders) Write(h *MessageHeader, w io.Writer) {
	//no payload
}

func NewMsgSendHeaders() *MsgSendHeaders {
	return &MsgSendHeaders{}
}

//

type MsgAddr struct {
	Num   uint64
	Addrs []*Address
}

func (m *MsgAddr) Command() string {
	return NMT_ADDR
}

func (m *MsgAddr) Read(h *MessageHeader, r io.Reader) {
	siz := GetAddressSize()
	l := 0
	m.Num = ReadVarInt(r, &l)
	num := (h.PayloadLen - uint32(l)) / uint32(siz)
	m.Addrs = make([]*Address, num)
	for i, _ := range m.Addrs {
		v := NewAddress(0, "0.0.0.0:0")
		v.Read(h.Ver >= 31402, r)
		m.Addrs[i] = v
	}
}

func (m *MsgAddr) Write(h *MessageHeader, w io.Writer) {
	//no payload
}

func NewMsgAddr() *MsgAddr {
	return &MsgAddr{}
}

//
type MsgGetAddr struct {
}

func (m *MsgGetAddr) Command() string {
	return NMT_GETADDR
}

func (m *MsgGetAddr) Read(h *MessageHeader, r io.Reader) {
	//no payload
}

func (m *MsgGetAddr) Write(h *MessageHeader, w io.Writer) {
	//no payload
}

func NewMsgGetAddr() *MsgGetAddr {
	return &MsgGetAddr{}
}
