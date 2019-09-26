package net

import (
	"bitcoin/config"
	"bitcoin/util"
	"bytes"
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
	Write(h *MessageHeader, w io.Writer)
	Read(h *MessageHeader, r io.Reader)
	Command() string
}

type MessageHeader struct {
	Start      []byte
	Command    string
	PayloadLen uint32
	CheckSum   []byte
	Ver        uint32 //data source server app version
}

func (m *MessageHeader) HasMagic() bool {
	conf := config.GetConfig()
	return bytes.Equal(m.Start, conf.MsgStart)
}

func (m *MessageHeader) IsValid(b []byte) bool {
	return bytes.Equal(util.HashP4(b), m.CheckSum)
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

type NetMessage struct {
	Header  *MessageHeader
	Payload io.Reader
}

func (m *NetMessage) Full(mp MsgIO) {
	mp.Read(m.Header, m.Payload)
}

//read package
func ReadMsg(r io.Reader) (h *NetMessage, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			err = rerr.(error)
		}
	}()
	h = &NetMessage{
		Header: &MessageHeader{},
	}
	h.Header.Read(r)
	if !h.Header.HasMagic() {
		panic(errors.New("start bytes error"))
	}
	pv := make([]byte, h.Header.PayloadLen)
	ReadBytes(r, pv)
	if !h.Header.IsValid(pv) {
		panic(errors.New("checksum error"))
	}
	h.Payload = bytes.NewReader(pv)
	return
}

func ToMessageBytes(w MsgIO) (ret []byte, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			ret = nil
			err = rerr.(error)
		}
	}()
	conf := config.GetConfig()

	m := &MessageHeader{}

	m.Start = conf.MsgStart
	m.Command = w.Command()

	pbuf := &bytes.Buffer{}
	w.Write(m, pbuf)

	m.PayloadLen = uint32(pbuf.Len())
	m.CheckSum = util.HashP4(pbuf.Bytes())

	hbuf := &bytes.Buffer{}
	m.Write(hbuf)

	hbuf.Write(pbuf.Bytes())

	ret = hbuf.Bytes()
	err = nil
	return
}

func (m *MessageHeader) Read(r io.Reader) {
	m.Start = []byte{0, 0, 0, 0}
	ReadBytes(r, m.Start)
	cmd := make([]byte, NMT_COMMAND_SIZE)
	ReadBytes(r, cmd)
	m.Command = util.String(cmd)
	m.PayloadLen = ReadUint32(r)
	m.CheckSum = []byte{0, 0, 0, 0}
	ReadBytes(r, m.CheckSum)
}

func (m *MessageHeader) Write(w io.Writer) {
	// 4 start
	WriteBytes(w, m.Start)
	//12 command
	cmd := make([]byte, NMT_COMMAND_SIZE)
	copy(cmd, []byte(m.Command))
	WriteBytes(w, cmd)
	//payload size
	WriteUint32(w, m.PayloadLen)
	//checksum
	WriteBytes(w, m.CheckSum)
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

func (a *Address) Read(pt bool, r io.Reader) {
	if pt {
		a.Time = ReadUint32(r)
	}
	a.Service = ReadUint64(r)
	a.IpAddr = make([]byte, net.IPv6len)
	ReadBytes(r, a.IpAddr)
	a.Port = ReadUint16(r)
}

func (a *Address) Write(pt bool, w io.Writer) {
	if pt {
		WriteUint32(w, a.Time)
	}
	WriteUint64(w, a.Service)
	WriteBytes(w, a.IpAddr[:])
	WriteUint16(w, a.Port)
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

func (m *MsgVersion) Read(h *MessageHeader, r io.Reader) {
	m.SAddr = NewAddress(0, "0.0.0.0:0")
	m.DAddr = NewAddress(0, "0.0.0.0:0")
	m.Ver = ReadUint32(r)
	m.Service = ReadUint64(r)
	m.Timestamp = ReadUint64(r)
	m.SAddr.Read(false, r)
	if m.Ver >= 106 {
		m.DAddr.Read(false, r)
		m.Nonce = ReadUint64(r)
		m.SubVer = ReadString(r)
		m.Height = ReadUint32(r)
	}
	if m.Ver >= 70001 {
		m.Relay = ReadUint8(r)
	}
}

func (m *MsgVersion) Write(h *MessageHeader, w io.Writer) {
	WriteUint32(w, m.Ver)
	WriteUint64(w, m.Service)
	WriteUint64(w, m.Timestamp)
	m.SAddr.Write(false, w)
	if m.Ver >= 106 {
		m.DAddr.Write(false, w)
		WriteUint64(w, m.Nonce)
		WriteString(w, m.SubVer)
		WriteUint32(w, m.Height)
	}
	if m.Ver >= 70001 {
		WriteUint8(w, m.Relay)
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

func (m *MsgPong) Read(h *MessageHeader, r io.Reader) {
	m.Timestamp = ReadUint64(r)
}

func (m *MsgPong) Write(h *MessageHeader, w io.Writer) {
	WriteUint64(w, m.Timestamp)
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

func (m *MsgPing) Read(h *MessageHeader, r io.Reader) {
	m.Timestamp = ReadUint64(r)
}

func (m *MsgPing) Write(h *MessageHeader, w io.Writer) {
	WriteUint64(w, m.Timestamp)
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

func (m *MsgVerAck) Read(h *MessageHeader, r io.Reader) {
	//no payload
}

func (m *MsgVerAck) Write(h *MessageHeader, w io.Writer) {
	//no payload
}

func NewMsgVerAck() *MsgVerAck {
	return &MsgVerAck{}
}
