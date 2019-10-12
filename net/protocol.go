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
	NMT_ALERT       = "alert"
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
	REJECT_MALFORMED       = byte(0x01)
	REJECT_INVALID         = byte(0x10)
	REJECT_OBSOLETE        = byte(0x11)
	REJECT_DUPLICATE       = byte(0x12)
	REJECT_NONSTANDARD     = byte(0x40)
	REJECT_DUST            = byte(0x41)
	REJECT_INSUFFICIENTFEE = byte(0x42)
	REJECT_CHECKPOINT      = byte(0x43)
)

const (
	MSG_ERROR          = 0
	MSG_TX             = 1
	MSG_BLOCK          = 2
	MSG_FILTERED_BLOCK = 3
	MSG_CMPCT_BLOCK    = 4
)

const (
	MAX_BLOCK_SERIALIZED_SIZE           = uint(4000000)
	MAX_BLOCK_WEIGHT                    = uint(4000000)
	MAX_BLOCK_SIGOPS_COST               = int64(80000)
	COINBASE_MATURITY                   = 100
	WITNESS_SCALE_FACTOR                = 4
	MIN_TRANSACTION_WEIGHT              = WITNESS_SCALE_FACTOR * 60
	MIN_SERIALIZABLE_TRANSACTION_WEIGHT = WITNESS_SCALE_FACTOR * 10
	LOCKTIME_VERIFY_SEQUENCE            = uint(1 << 0)
	LOCKTIME_MEDIAN_TIME_PAST           = uint(1 << 1)
)

var (
	SizeError = errors.New("data size error")
)

func HASH256To(b []byte, h *HashID) {
	copy((*h)[:], util.HASH256(b))
}

func HASH160To(b []byte, h []byte) {
	copy(h, util.HASH160(b))
}

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

//read  rwd = string
//write rwd = []byte
//default
func NewNetHeader(rwds ...interface{}) *NetHeader {
	m := &NetHeader{}
	if len(rwds) == 0 {
		m.Command = NMT_UNKNNOW
		m.MsgBuffer = NewMsgBuffer([]byte{}, MSG_BUFFER_WRITE)
	} else if av, ok := rwds[0].([]byte); ok {
		m.Command = NMT_UNKNNOW
		m.MsgBuffer = NewMsgBuffer(av, MSG_BUFFER_READ)
	} else if av, ok := rwds[0].(string); ok {
		m.Command = av
		m.MsgBuffer = NewMsgBuffer([]byte{}, MSG_BUFFER_WRITE)
	} else {
		m.Command = NMT_UNKNNOW
		m.MsgBuffer = NewMsgBuffer([]byte{}, MSG_BUFFER_WRITE)
	}
	conf := config.GetConfig()
	m.Ver = PROTOCOL_VERSION
	m.Start = conf.MsgStart
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

func (h *NetHeader) Full(mp MsgIO) MsgIO {
	mp.Read(h)
	return mp
}

//read package
func ReadMsg(r io.Reader) (*NetHeader, error) {
	h := NewNetHeader(MSG_BUFFER_READ)
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
	m := NewNetHeader(w.Command())
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
type MsgAlert struct {
	Ver        int32
	Relay      int64
	Expiration int64
	ID         int32
	Cancels    []int32
	MinVer     int32
	MaxVer     int32
	SubVers    []string
	Priority   int32
	Comment    string
	StatusBar  string
	Reserved   string
}

func (m *MsgAlert) Command() string {
	return NMT_ALERT
}

func (m *MsgAlert) Read(h *NetHeader) {
	m.Ver = h.ReadInt32()
	m.Relay = h.ReadInt64()
	m.Expiration = h.ReadInt64()
	m.ID = h.ReadInt32()
	cl, _ := h.ReadVarInt()
	m.Cancels = make([]int32, cl)
	for i, _ := range m.Cancels {
		m.Cancels[i] = h.ReadInt32()
	}
	m.MinVer = h.ReadInt32()
	m.MaxVer = h.ReadInt32()
	sl, _ := h.ReadVarInt()
	m.SubVers = make([]string, sl)
	for i, _ := range m.SubVers {
		m.SubVers[i] = h.ReadString()
	}
	m.Priority = h.ReadInt32()
	m.Comment = h.ReadString()
	m.StatusBar = h.ReadString()
	m.Reserved = h.ReadString()
}

func (m *MsgAlert) Write(h *NetHeader) {
	//no payload
}

func NewMsgAlert() *MsgAlert {
	return &MsgAlert{}
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
