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
	PROTOCOL_VERSION = uint32(70015)
	SERVICE_NETWORK  = uint64(1)
	DEFAULT_PORT     = uint16(8333)
)

//message length
const (
	NMT_MESSAGE_START_SIZE  = 4
	NMT_COMMAND_SIZE        = 12
	NMT_MESSAGE_SIZE_SIZE   = 4
	NMT_CHECKSUM_SIZE       = 4
	NMT_MESSAGE_SIZE_OFFSET = NMT_MESSAGE_START_SIZE + NMT_COMMAND_SIZE
	NMT_CHECKSUM_OFFSET     = NMT_MESSAGE_SIZE_OFFSET + NMT_MESSAGE_SIZE_SIZE
	NMT_HEADER_SIZE         = NMT_MESSAGE_START_SIZE + NMT_COMMAND_SIZE + NMT_MESSAGE_SIZE_SIZE + NMT_CHECKSUM_SIZE
)

var (
	SizeError = errors.New("data size error")
)

type MsgIO interface {
	Write(w io.Writer)
	Read(r io.Reader)
	Command() string
}

type MessageHeader struct {
	Start      []byte
	Command    string
	PayloadLen uint32
	CheckSum   []byte
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

//read package
func ReadMsg(r io.Reader) (h *MessageHeader, pr io.Reader, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			pr = nil
			h = nil
			err = rerr.(error)
		}
	}()
	conf := config.GetConfig()
	ret := &MessageHeader{}
	ret.Read(r)
	if !bytes.Equal(ret.Start, conf.MsgStart) {
		panic(errors.New("start bytes error"))
	}
	h = ret
	if h.PayloadLen == 0 {
		return
	}
	pv := make([]byte, h.PayloadLen)
	ReadBytes(r, pv)
	sum := util.HashP4(pv)
	if !bytes.Equal(sum, h.CheckSum) {
		panic(errors.New("checksum error"))
	}
	pr = bytes.NewReader(pv)
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

	pbuf := &bytes.Buffer{}
	w.Write(pbuf)

	m := MessageHeader{}

	m.Start = make([]byte, 4)
	copy(m.Start, conf.MsgStart)

	m.Command = w.Command()

	m.PayloadLen = uint32(pbuf.Len())

	if m.PayloadLen > 0 {
		m.CheckSum = util.HashP4(pbuf.Bytes())
	}
	hbuf := &bytes.Buffer{}
	m.Write(hbuf)
	hbuf.Write(pbuf.Bytes())
	ret = hbuf.Bytes()
	err = nil
	return
}

func ReadBytes(r io.Reader, b []byte) {
	num, err := r.Read(b)
	if err != nil {
		panic(err)
	}
	if num != len(b) {
		panic(SizeError)
	}
}

func WriteBytes(w io.Writer, b []byte) {
	num, err := w.Write(b)
	if err != nil {
		panic(err)
	}
	if num != len(b) {
		panic(SizeError)
	}
}

func ReadUint8(r io.Reader) uint8 {
	v := uint8(0)
	err := binary.Read(r, binary.LittleEndian, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUint8(w io.Writer, v uint8) {
	err := binary.Write(w, binary.LittleEndian, v)
	if err != nil {
		panic(err)
	}
}

func ReadUint16(r io.Reader) uint16 {
	v := uint16(0)
	err := binary.Read(r, binary.LittleEndian, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUint16(w io.Writer, v uint16) {
	err := binary.Write(w, binary.LittleEndian, v)
	if err != nil {
		panic(err)
	}
}

func ReadUint32(r io.Reader) uint32 {
	v := uint32(0)
	err := binary.Read(r, binary.LittleEndian, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUint32(w io.Writer, v uint32) {
	err := binary.Write(w, binary.LittleEndian, v)
	if err != nil {
		panic(err)
	}
}

func ReadUint64(r io.Reader) uint64 {
	v := uint64(0)
	err := binary.Read(r, binary.LittleEndian, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func WriteUint64(w io.Writer, v uint64) {
	err := binary.Write(w, binary.LittleEndian, v)
	if err != nil {
		panic(err)
	}
}

func ReadString(r io.Reader) string {
	l := ReadUint8(r)
	if l == 0 {
		return ""
	}
	b := make([]byte, l)
	ReadBytes(r, b)
	return string(b)
}

func WriteString(w io.Writer, v string) {
	l := len(v)
	WriteUint8(w, uint8(l))
	if l > 0 {
		WriteBytes(w, []byte(v))
	}
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
	if m.PayloadLen > 0 {
		WriteBytes(w, m.CheckSum)
	}
}

type Address struct {
	Service uint64 //1
	IpAddr  net.IP
	Port    uint16
}

func NewAddress(s uint64, ip string, port uint16) *Address {
	return &Address{
		Service: s,
		IpAddr:  net.ParseIP(ip),
		Port:    port,
	}
}

func (a *Address) Read(r io.Reader) {
	a.Service = ReadUint64(r)
	a.IpAddr = make([]byte, net.IPv6len)
	ReadBytes(r, a.IpAddr)
	a.Port = ReadUint16(r)
}

func (a *Address) Write(w io.Writer) {
	WriteUint64(w, a.Service)
	WriteBytes(w, a.IpAddr[:])
	WriteUint16(w, a.Port)
}

type MsgVersion struct {
	//version payload
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

func (m *MsgVersion) Read(r io.Reader) {
	m.SAddr = NewAddress(0, "0.0.0.0", 0)
	m.DAddr = NewAddress(0, "0.0.0.0", 0)
	m.Ver = ReadUint32(r)
	m.Service = ReadUint64(r)
	m.Timestamp = ReadUint64(r)
	m.SAddr.Read(r)
	m.DAddr.Read(r)
	m.Nonce = ReadUint64(r)
	m.SubVer = ReadString(r)
	m.Height = ReadUint32(r)
	m.Relay = ReadUint8(r)
}

func (m *MsgVersion) Write(w io.Writer) {
	WriteUint32(w, m.Ver)
	WriteUint64(w, m.Service)
	WriteUint64(w, m.Timestamp)
	m.SAddr.Write(w)
	m.DAddr.Write(w)
	WriteUint64(w, m.Nonce)
	WriteString(w, m.SubVer)
	WriteUint32(w, m.Height)
	WriteUint8(w, m.Relay)
}

func NewMsgVersion(sip string, dip string) *MsgVersion {
	conf := config.GetConfig()
	m := &MsgVersion{}
	m.Ver = PROTOCOL_VERSION
	m.Service = SERVICE_NETWORK
	m.Timestamp = uint64(time.Now().Unix())
	m.SAddr = NewAddress(SERVICE_NETWORK, sip, DEFAULT_PORT)
	m.DAddr = NewAddress(SERVICE_NETWORK, dip, DEFAULT_PORT)
	m.Nonce = util.RandUInt64()
	m.SubVer = conf.SubVer
	m.Height = 0
	m.Relay = 1
	return m
}
