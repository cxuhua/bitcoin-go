package net

import (
	"encoding/hex"
	"io"
	"log"
	"strings"
)

type BHash [32]byte

type Inventory struct {
	Type uint32
	Hash BHash
}

func (m *Inventory) Read(r io.Reader) {
	m.Type = ReadUint32(r)
	ReadBytes(r, m.Hash[:])
}

func (m *Inventory) Write(w io.Writer) {
	WriteUint32(w, m.Type)
	WriteBytes(w, m.Hash[:])
}

func (h BHash) String() string {
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

type BHeader struct {
	Ver       uint32
	Prev      BHash
	Root      BHash
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
	Count     uint64
}

func (m *BHeader) Read(r io.Reader) {
	m.Ver = ReadUint32(r)
	ReadBytes(r, m.Prev[:])
	ReadBytes(r, m.Root[:])
	m.Timestamp = ReadUint32(r)
	m.Bits = ReadUint32(r)
	m.Nonce = ReadUint32(r)
	m.Count = ReadVarInt(r)
}

func (m *BHeader) Write(w io.Writer) {
	WriteUint32(w, m.Ver)
	WriteBytes(w, m.Prev[:])
	WriteBytes(w, m.Root[:])
	WriteUint32(w, m.Timestamp)
	WriteUint32(w, m.Bits)
	WriteUint32(w, m.Nonce)
	WriteVarInt(w, m.Count)
}

type TxOutPoint struct {
	Hash  BHash
	Index uint32
}

func (m *TxOutPoint) Read(r io.Reader) {
	ReadBytes(r, m.Hash[:])
	m.Index = ReadUint32(r)
}

func (m *TxOutPoint) Write(w io.Writer) {
	WriteBytes(w, m.Hash[:])
	WriteUint32(w, m.Index)
}

type Script []byte

type TxOut struct {
	Value  uint64
	Script Script
}

func (m *TxOut) Read(r io.Reader) {
	m.Value = ReadUint64(r)
	l := ReadVarInt(r)
	m.Script = make(Script, l)
	ReadBytes(r, m.Script)
}

func (m *TxOut) Write(w io.Writer) {
	WriteUint64(w, m.Value)
	WriteVarInt(w, uint64(len(m.Script)))
	WriteBytes(w, m.Script)
}

type TxIn struct {
	Output   TxOutPoint
	Script   Script
	Sequence uint32
}

func (m *TxIn) Read(r io.Reader) {
	m.Output.Read(r)
	l := ReadVarInt(r)
	m.Script = make(Script, l)
	log.Println("Script len", l)
	ReadBytes(r, m.Script)
	m.Sequence = ReadUint32(r)
}

func (m *TxIn) Write(w io.Writer) {
	m.Output.Write(w)
	WriteVarInt(w, uint64(len(m.Script)))
	WriteBytes(w, m.Script)
	WriteUint32(w, m.Sequence)
}

type TX struct {
	Ver      uint32
	Ins      []*TxIn
	Outs     []*TxOut
	LockTime uint32
}

func (m *TX) Read(r io.Reader) {
	m.Ver = ReadUint32(r)
	log.Println("Tx VER = ", m.Ver)
	il := ReadVarInt(r)
	m.Ins = make([]*TxIn, il)
	for i, _ := range m.Ins {
		v := &TxIn{}
		v.Read(r)
		m.Ins[i] = v
	}
	ol := ReadVarInt(r)
	m.Outs = make([]*TxOut, ol)
	for i, _ := range m.Outs {
		v := &TxOut{}
		v.Read(r)
		m.Outs[i] = v
	}
	m.LockTime = ReadUint32(r)
}

func (m *TX) Write(w io.Writer) {
	WriteUint32(w, m.Ver)
	WriteVarInt(w, uint64(len(m.Ins)))
	for _, v := range m.Ins {
		v.Write(w)
	}
	WriteVarInt(w, uint64(len(m.Outs)))
	for _, v := range m.Outs {
		v.Write(w)
	}
	WriteUint32(w, m.LockTime)
}

//

type MsgBlock struct {
	Ver       uint32
	Prev      BHash
	Root      BHash
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
	Txs       []*TX
}

func (m *MsgBlock) Command() string {
	return NMT_BLOCK
}

func (m *MsgBlock) Read(h *MessageHeader, r io.Reader) {
	m.Ver = ReadUint32(r)
	ReadBytes(r, m.Prev[:])
	ReadBytes(r, m.Root[:])
	log.Println(m.Prev, m.Root)
	m.Timestamp = ReadUint32(r)
	m.Bits = ReadUint32(r)
	m.Nonce = ReadUint32(r)
	l := ReadVarInt(r)
	m.Txs = make([]*TX, l)
	for i, _ := range m.Txs {
		v := &TX{}
		v.Read(r)
		m.Txs[i] = v
	}
}

func (m *MsgBlock) Write(h *MessageHeader, w io.Writer) {

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
	num := ReadVarInt(r)
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

//
type MsgINV struct {
	Invs []*Inventory
}

func (m *MsgINV) Command() string {
	return NMT_INV
}

func (m *MsgINV) Read(h *MessageHeader, r io.Reader) {
	num := ReadVarInt(r)
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
	m.Ver = ReadUint32(r)
	num := ReadVarInt(r)
	m.Blocks = make([]*BHash, num)
	for i, _ := range m.Blocks {
		m.Blocks[i] = &BHash{}
		ReadBytes(r, m.Blocks[i][:])
	}
	m.Stop = &BHash{}
	ReadBytes(r, m.Stop[:])
}

func (m *MsgGetHeaders) Write(h *MessageHeader, w io.Writer) {
	WriteUint32(w, m.Ver)
	WriteVarInt(w, uint64(len(m.Blocks)))
	for _, v := range m.Blocks {
		WriteBytes(w, v[:])
	}
	WriteBytes(w, m.Stop[:])
}

func NewMsgGetHeaders() *MsgGetHeaders {
	return &MsgGetHeaders{}
}
