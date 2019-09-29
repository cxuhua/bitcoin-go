package net

import (
	"bitcoin/script"
	"encoding/hex"
	"io"
	"log"
	"strings"
)

type BHash [32]byte

func NewBHashWithString(s string) BHash {
	b := BHash{}
	if len(s) != len(b)*2 {
		panic(SizeError)
	}
	v, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	copy(b[:], v)
	return b
}

type Inventory struct {
	Type uint32
	Hash BHash
}

func (m *Inventory) Read(r io.Reader) {
	m.Type = ReadUInt32(r)
	ReadBytes(r, m.Hash[:])
}

func (m *Inventory) Write(w io.Writer) {
	WriteUInt32(w, m.Type)
	WriteBytes(w, m.Hash[:])
}

func (h BHash) String() string {
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

type BHeader struct {
	Ver       uint32
	Prev      BHash
	Root      BHash //Merkle tree root
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
	Count     uint64
}

func (m *BHeader) Read(r io.Reader) {
	m.Ver = ReadUInt32(r)
	ReadBytes(r, m.Prev[:])
	ReadBytes(r, m.Root[:])
	m.Timestamp = ReadUInt32(r)
	m.Bits = ReadUInt32(r)
	m.Nonce = ReadUInt32(r)
	m.Count, _ = ReadVarInt(r)
}

func (m *BHeader) Write(w io.Writer) {
	WriteUInt32(w, m.Ver)
	WriteBytes(w, m.Prev[:])
	WriteBytes(w, m.Root[:])
	WriteUInt32(w, m.Timestamp)
	WriteUInt32(w, m.Bits)
	WriteUInt32(w, m.Nonce)
	WriteVarInt(w, m.Count)
}

type TxOutPoint struct {
	Hash  BHash
	Index uint32
}

func (m *TxOutPoint) Read(r io.Reader) {
	ReadBytes(r, m.Hash[:])
	m.Index = ReadUInt32(r)
}

func (m *TxOutPoint) Write(w io.Writer) {
	WriteBytes(w, m.Hash[:])
	WriteUInt32(w, m.Index)
}

type TxOut struct {
	Value  uint64
	Script *script.Script
}

func (m *TxOut) Read(r io.Reader) {
	m.Value = ReadUInt64(r)
	m.Script = ReadScript(r)
}

func (m *TxOut) Write(w io.Writer) {
	WriteUInt64(w, m.Value)
	WriteScript(w, m.Script)
}

type TxIn struct {
	Output   TxOutPoint
	Script   *script.Script
	Sequence uint32
}

func (m *TxIn) Read(r io.Reader) {
	m.Output.Read(r)
	m.Script = ReadScript(r)
	m.Sequence = ReadUInt32(r)
}

func (m *TxIn) Write(w io.Writer) {
	m.Output.Write(w)
	WriteScript(w, m.Script)
	WriteUInt32(w, m.Sequence)
}

type TX struct {
	Ver      int32
	Ins      []*TxIn
	Outs     []*TxOut
	LockTime uint32
}

func (m *TX) Read(r io.Reader) {
	m.Ver = int32(ReadUInt32(r))
	w := ReadUInt16(r)
	log.Println(w)
	il, _ := ReadVarInt(r)
	m.Ins = make([]*TxIn, il)
	for i, _ := range m.Ins {
		v := &TxIn{}
		v.Read(r)
		m.Ins[i] = v
	}
	ol, _ := ReadVarInt(r)
	m.Outs = make([]*TxOut, ol)
	for i, _ := range m.Outs {
		v := &TxOut{}
		v.Read(r)
		m.Outs[i] = v
	}
	m.LockTime = ReadUInt32(r)
}

func (m *TX) Write(w io.Writer) {
	WriteUInt32(w, uint32(m.Ver))
	WriteVarInt(w, uint64(len(m.Ins)))
	for _, v := range m.Ins {
		v.Write(w)
	}
	WriteVarInt(w, uint64(len(m.Outs)))
	for _, v := range m.Outs {
		v.Write(w)
	}
	WriteUInt32(w, m.LockTime)
}

//
type MsgHeaders struct {
	Headers []*BHeader
}

func (m *MsgHeaders) Command() string {
	return NMT_HEADERS
}

func (m *MsgHeaders) Read(h *MessageHeader, r io.Reader) {
	num, _ := ReadVarInt(r)
	m.Headers = make([]*BHeader, num)
	for i, _ := range m.Headers {
		v := &BHeader{}
		v.Read(r)
		m.Headers[i] = v
	}
}

func (m *MsgHeaders) Write(h *MessageHeader, w io.Writer) {
	WriteVarInt(w, uint64(len(m.Headers)))
	for _, v := range m.Headers {
		v.Write(w)
	}
}

func NewMsgHeaders() *MsgHeaders {
	return &MsgHeaders{}
}

//
type MsgGetBlocks struct {
	Ver    uint32
	Blocks []*BHash
	Stop   *BHash
}

func (m *MsgGetBlocks) Command() string {
	return NMT_GETBLOCKS
}

func (m *MsgGetBlocks) Read(h *MessageHeader, r io.Reader) {
	m.Ver = ReadUInt32(r)
	num, _ := ReadVarInt(r)
	m.Blocks = make([]*BHash, num)
	for i, _ := range m.Blocks {
		m.Blocks[i] = &BHash{}
		ReadBytes(r, m.Blocks[i][:])
	}
	m.Stop = &BHash{}
	ReadBytes(r, m.Stop[:])
}

func (m *MsgGetBlocks) Write(h *MessageHeader, w io.Writer) {
	WriteUInt32(w, m.Ver)
	WriteVarInt(w, uint64(len(m.Blocks)))
	for _, v := range m.Blocks {
		WriteBytes(w, v[:])
	}
	WriteBytes(w, m.Stop[:])
}

func NewMsgGetBlocks() *MsgGetBlocks {
	return &MsgGetBlocks{}
}

//

type MsgNotFound struct {
	Invs []*Inventory
}

func (m *MsgNotFound) Command() string {
	return NMT_NOTFOUND
}

func (m *MsgNotFound) Read(h *MessageHeader, r io.Reader) {
	size, _ := ReadVarInt(r)
	m.Invs = make([]*Inventory, size)
	for i, _ := range m.Invs {
		v := &Inventory{}
		v.Read(r)
		m.Invs[i] = v
	}
}

func (m *MsgNotFound) Write(h *MessageHeader, w io.Writer) {
	WriteVarInt(w, uint64(len(m.Invs)))
	for _, v := range m.Invs {
		v.Write(w)
	}
}

func NewMsgNotFound() *MsgNotFound {
	return &MsgNotFound{}
}

//

type MsgBlock struct {
	Ver       uint32
	Prev      BHash
	Root      BHash //Merkle tree root
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
	Txs       []*TX
}

func (m *MsgBlock) Command() string {
	return NMT_BLOCK
}

func (m *MsgBlock) Read(h *MessageHeader, r io.Reader) {
	m.Ver = ReadUInt32(r)
	ReadBytes(r, m.Prev[:])
	ReadBytes(r, m.Root[:])
	m.Timestamp = ReadUInt32(r)
	m.Bits = ReadUInt32(r)
	m.Nonce = ReadUInt32(r)
	l, _ := ReadVarInt(r)
	m.Txs = make([]*TX, l)
	for i, _ := range m.Txs {
		v := &TX{}
		v.Read(r)
		m.Txs[i] = v
	}
}

func (m *MsgBlock) Write(h *MessageHeader, w io.Writer) {
	WriteUInt32(w, m.Ver)
	WriteBytes(w, m.Prev[:])
	WriteBytes(w, m.Root[:])
	WriteUInt32(w, m.Timestamp)
	WriteUInt32(w, m.Bits)
	WriteUInt32(w, m.Nonce)
	WriteVarInt(w, uint64(len(m.Txs)))
	for _, v := range m.Txs {
		v.Write(w)
	}
}

func NewMsgBlock() *MsgBlock {
	return &MsgBlock{}
}

//
type MsgGetData struct {
	Invs []*Inventory
}

func (m *MsgGetData) Command() string {
	return NMT_GETDATA
}

func (m *MsgGetData) Read(h *MessageHeader, r io.Reader) {
	num, _ := ReadVarInt(r)
	m.Invs = make([]*Inventory, num)
	for i, _ := range m.Invs {
		m.Invs[i] = &Inventory{}
		m.Invs[i].Read(r)
	}
}
func (m *MsgGetData) Add(inv *Inventory) {
	m.Invs = append(m.Invs, inv)
}

func (m *MsgGetData) Write(h *MessageHeader, w io.Writer) {
	WriteVarInt(w, uint64(len(m.Invs)))
	for _, v := range m.Invs {
		v.Write(w)
	}
}

func NewMsgGetData() *MsgGetData {
	return &MsgGetData{
		Invs: []*Inventory{},
	}
}

//
type MsgTX struct {
	Tx TX
}

func (m *MsgTX) Command() string {
	return NMT_TX
}

func (m *MsgTX) Read(h *MessageHeader, r io.Reader) {
	m.Tx.Read(r)
}

func (m *MsgTX) Write(h *MessageHeader, w io.Writer) {
	m.Tx.Write(w)
}

func NewMsgTX() *MsgTX {
	return &MsgTX{}
}

type MsgINV struct {
	Invs []*Inventory
}

func (m *MsgINV) Command() string {
	return NMT_INV
}

func (m *MsgINV) Read(h *MessageHeader, r io.Reader) {
	num, _ := ReadVarInt(r)
	m.Invs = make([]*Inventory, num)
	for i, _ := range m.Invs {
		m.Invs[i] = &Inventory{}
		m.Invs[i].Read(r)
	}
}

func (m *MsgINV) Write(h *MessageHeader, w io.Writer) {
	WriteVarInt(w, uint64(len(m.Invs)))
	for _, v := range m.Invs {
		v.Write(w)
	}
}

func NewMsgINV() *MsgINV {
	return &MsgINV{}
}

//

type MsgGetHeaders struct {
	Ver    uint32
	Blocks []*BHash
	Stop   *BHash
}

func (m *MsgGetHeaders) Command() string {
	return NMT_GETHEADERS
}

func (m *MsgGetHeaders) Read(h *MessageHeader, r io.Reader) {
	m.Ver = ReadUInt32(r)
	num, _ := ReadVarInt(r)
	m.Blocks = make([]*BHash, num)
	for i, _ := range m.Blocks {
		m.Blocks[i] = &BHash{}
		ReadBytes(r, m.Blocks[i][:])
	}
	m.Stop = &BHash{}
	ReadBytes(r, m.Stop[:])
}

func (m *MsgGetHeaders) Write(h *MessageHeader, w io.Writer) {
	WriteUInt32(w, m.Ver)
	WriteVarInt(w, uint64(len(m.Blocks)))
	for _, v := range m.Blocks {
		WriteBytes(w, v[:])
	}
	WriteBytes(w, m.Stop[:])
}

func NewMsgGetHeaders() *MsgGetHeaders {
	return &MsgGetHeaders{}
}
