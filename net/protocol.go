package net

import (
	"bitcoin/config"
	"bitcoin/util"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
)

//net message type
//https://en.bitcoin.it/wiki/Protocol_documentation#Network_address
const (
	NMT_VERSION     = "version"
	NMT_VERACK      = "verack"
	NMT_ADDR        = "addr"
	NMT_INV         = "inv"
	NMT_GETDATA     = "getdata"
	NMT_MERKLEBLOCK = "merkleblock"
	NMT_GETBLOCKS   = "getblocks"
	NMT_GETHEADERS  = "getheaders"
	NMT_TX          = "tx"
	NMT_HEADERS     = "headers"
	NMT_BLOCK       = "block"
	NMT_GETADDR     = "getaddr"
	NMT_MEMPOOL     = "mempool"
	NMT_PING        = "ping"
	NMT_PONG        = "pong"
	NMT_NOTFOUND    = "notfound"
	NMT_FILTERLOAD  = "filterload"
	NMT_FILTERADD   = "filteradd"
	NMT_FILTERCLEAR = "filterclear"
	NMT_REJECT      = "reject"
	NMT_SENDHEADERS = "sendheaders"
	NMT_FEEFILTER   = "feefilter"
	NMT_SENDCMPCT   = "sendcmpct"
	NMT_CMPCTBLOCK  = "cmpctblock"
	NMT_GETBLOCKTXN = "getblocktxn"
	NMT_BLOCKTXN    = "blocktxn"
	NMT_UNKNNOW     = "unknow"
)

const (
	NODE_NETWORK         = uint64(1)
	NODE_GETUTXO         = uint64(2)
	NODE_BLOOM           = uint64(4)
	NODE_WITNESS         = uint64(8)
	NODE_NETWORK_LIMITED = uint64(1024)
)

const (
	PROTOCOL_VERSION = uint32(70015)
	SERVICE_NETWORK  = NODE_NETWORK
	DEFAULT_PORT     = uint16(8333)
)

//message length
const (
	NMT_MESSAGE_START_SIZE = 4
	NMT_COMMAND_SIZE       = 12
	NMT_MESSAGE_SIZE_SIZE  = 4
	NMT_CHECKSUM_SIZE      = 4
)

const (
	REJECT_MALFORMED       = 0x01
	REJECT_INVALID         = 0x10
	REJECT_OBSOLETE        = 0x11
	REJECT_DUPLICATE       = 0x12
	REJECT_NONSTANDARD     = 0x40
	REJECT_DUST            = 0x41
	REJECT_INSUFFICIENTFEE = 0x42
	REJECT_CHECKPOINT      = 0x43
)

const (
	MSG_ERROR          = 0
	MSG_TX             = 1
	MSG_BLOCK          = 2
	MSG_FILTERED_BLOCK = 3
	MSG_CMPCT_BLOCK    = 4
)

var (
	SizeError = errors.New("data size error")
)

type MsgIO interface {
	Write(h *NetHeader)
	Read(h *NetHeader)
	Command() string
}

type NetHeader struct {
	*MsgBuffer
	Start    []byte //4
	Command  string //12
	CheckSum []byte //4
	Ver      uint32 //4 data source server app version
}

func NewNetHeader(cmd string, b []byte) *NetHeader {
	m := &NetHeader{}
	conf := config.GetConfig()
	m.Ver = PROTOCOL_VERSION
	m.Start = conf.MsgStart
	m.Command = cmd
	m.MsgBuffer = NewMsgBuffer(b)
	return m
}

func (m *NetHeader) HeadLen() uint32 {
	return 24
}

func (m *NetHeader) Len() uint32 {
	return uint32(len(m.Payload))
}

func (m *NetHeader) HasMagic() bool {
	conf := config.GetConfig()
	return bytes.Equal(m.Start, conf.MsgStart)
}

func (m *NetHeader) IsValid() bool {
	return bytes.Equal(util.HashP4(m.Payload), m.CheckSum)
}

func WriteMsg(w io.Writer, m MsgIO) error {
	b, err := ToMessageBytes(m)
	if err != nil {
		return err
	}
	num, err := w.Write(b)
	if err != nil {
		return err
	}
	if num != len(b) {
		return SizeError
	}
	return nil
}

func (h *NetHeader) Full(mp MsgIO) {
	mp.Read(h)
}

//read package
func ReadMsg(r io.Reader) (*NetHeader, error) {
	h := NewNetHeader(NMT_BLOCK, []byte{})
	if err := h.ReadHeader(r); err != nil {
		return nil, err
	}
	if !h.HasMagic() {
		return nil, errors.New("start bytes error")
	}
	if err := binary.Read(r, ByteOrder, h.Payload); err != nil {
		return nil, err
	}
	if !h.IsValid() {
		return nil, errors.New("checksum error")
	}
	return h, nil
}

func ToMessageBytes(w MsgIO) ([]byte, error) {
	m := NewNetHeader(w.Command(), []byte{})
	//full payload
	w.Write(m)
	//get send bytes
	hb := &bytes.Buffer{}
	if err := m.WriteHeader(hb); err != nil {
		return nil, err
	}
	return hb.Bytes(), nil
}

func (m *NetHeader) ReadHeader(r io.Reader) error {
	m.Start = []byte{0, 0, 0, 0}
	if err := binary.Read(r, ByteOrder, m.Start); err != nil {
		return err
	}
	cmd := make([]byte, NMT_COMMAND_SIZE)
	if err := binary.Read(r, ByteOrder, cmd); err != nil {
		return err
	}
	m.Command = util.String(cmd)
	pl := uint32(0)
	if err := binary.Read(r, ByteOrder, &pl); err != nil {
		return err
	}
	m.Payload = make([]byte, pl)
	m.CheckSum = []byte{0, 0, 0, 0}
	if err := binary.Read(r, ByteOrder, m.CheckSum); err != nil {
		return err
	}
	return nil
}

func (m *NetHeader) WriteHeader(w io.Writer) error {
	// 4 start
	if err := binary.Write(w, ByteOrder, m.Start); err != nil {
		return err
	}
	//12 command
	cmd := make([]byte, NMT_COMMAND_SIZE)
	copy(cmd, []byte(m.Command))
	if err := binary.Write(w, ByteOrder, cmd); err != nil {
		return err
	}
	//payload size
	pl := uint32(len(m.Payload))
	if err := binary.Write(w, ByteOrder, pl); err != nil {
		return err
	}
	m.CheckSum = util.HashP4(m.Payload)
	//checksum
	if err := binary.Write(w, ByteOrder, m.CheckSum); err != nil {
		return err
	}
	//payload
	if err := binary.Write(w, ByteOrder, m.Payload); err != nil {
		return err
	}
	return nil
}

type Address struct {
	Time    uint32 //4 Not present in version message.
	Service uint64 //8
	IpAddr  net.IP //16
	Port    uint16 //2
}

func GetAddressSize() int {
	return 4 + 8 + 16 + 2
}

func NewAddress(s uint64, addr string) *Address {
	ip, port := util.ParseAddr(addr)
	return &Address{
		Service: s,
		IpAddr:  ip,
		Port:    port,
	}
}

func (a *Address) Read(m *NetHeader, pt bool) {
	if pt {
		a.Time = m.ReadUInt32()
	}
	a.Service = m.ReadUInt64()
	a.IpAddr = make([]byte, net.IPv6len)
	m.ReadBytes(a.IpAddr)
	a.Port = m.ReadUInt16()
}

func (a *Address) Write(m *NetHeader, pt bool) {
	if pt {
		m.WriteUInt32(a.Time)
	}
	m.WriteUInt64(a.Service)
	m.WriteBytes(a.IpAddr[:])
	m.WriteUInt16(a.Port)
}

//version payload
type MsgVersion struct {
	Ver       uint32 //PROTOCOL_VERSION
	Service   uint64 //1
	Timestamp uint64
	SAddr     *Address
	DAddr     *Address
	Nonce     uint64
	SubVer    string
	Height    uint32
	Relay     uint8
}

func (m *MsgVersion) Command() string {
	return NMT_VERSION
}

func (m *MsgVersion) Read(h *NetHeader) {
	m.SAddr = NewAddress(0, "0.0.0.0:0")
	m.DAddr = NewAddress(0, "0.0.0.0:0")
	m.Ver = h.ReadUInt32()
	m.Service = h.ReadUInt64()
	m.Timestamp = h.ReadUInt64()
	m.SAddr.Read(h, false)
	if m.Ver >= 106 {
		m.DAddr.Read(h, false)
		m.Nonce = h.ReadUInt64()
		m.SubVer = h.ReadString()
		m.Height = h.ReadUInt32()
	}
	if m.Ver >= 70001 {
		m.Relay = h.ReadUint8()
	}
}

func (m *MsgVersion) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteUInt64(m.Service)
	h.WriteUInt64(m.Timestamp)
	m.SAddr.Write(h, false)
	if m.Ver >= 106 {
		m.DAddr.Write(h, false)
		h.WriteUInt64(m.Nonce)
		h.WriteString(m.SubVer)
		h.WriteUInt32(m.Height)
	}
	if m.Ver >= 70001 {
		h.WriteUint8(m.Relay)
	}
}

func NewMsgVersion(sip string, dip string) *MsgVersion {
	conf := config.GetConfig()
	m := &MsgVersion{}
	m.Ver = PROTOCOL_VERSION
	m.Service = SERVICE_NETWORK
	m.Timestamp = uint64(time.Now().Unix())
	m.SAddr = NewAddress(SERVICE_NETWORK, sip)
	m.DAddr = NewAddress(SERVICE_NETWORK, dip)
	m.Nonce = util.RandUInt64()
	m.SubVer = conf.SubVer
	m.Height = 0
	m.Relay = 1
	return m
}

//
type MsgPong struct {
	Timestamp uint64
}

func (m *MsgPong) Command() string {
	return NMT_PONG
}

func (m *MsgPong) Read(h *NetHeader) {
	m.Timestamp = h.ReadUInt64()
}

func (m *MsgPong) Write(h *NetHeader) {
	h.WriteUInt64(m.Timestamp)
}

func (m *MsgPong) Ping() int {
	v := uint64(time.Now().UnixNano()) - m.Timestamp
	return int(v / 1000000)
}

func NewMsgPong() *MsgPong {
	return &MsgPong{Timestamp: uint64(time.Now().Unix())}
}

//
type MsgPing struct {
	Timestamp uint64
}

func (m *MsgPing) Command() string {
	return NMT_PING
}

func (m *MsgPing) Read(h *NetHeader) {
	m.Timestamp = h.ReadUInt64()
}

func (m *MsgPing) Write(h *NetHeader) {
	h.WriteUInt64(m.Timestamp)
}

func NewMsgPing() *MsgPing {
	return &MsgPing{Timestamp: uint64(time.Now().UnixNano())}
}

//
type MsgVerAck struct {
}

func (m *MsgVerAck) Command() string {
	return NMT_VERACK
}

func (m *MsgVerAck) Read(h *NetHeader) {
	//no payload
}

func (m *MsgVerAck) Write(h *NetHeader) {
	//no payload
}

func NewMsgVerAck() *MsgVerAck {
	return &MsgVerAck{}
}
