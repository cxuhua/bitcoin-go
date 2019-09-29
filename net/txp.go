package net

import (
	"bitcoin/script"
	"bitcoin/util"
	"encoding/hex"
)

type HashID [32]byte

func (h HashID) String() string {
	return hex.EncodeToString(h[:])
}

func (b HashID) Swap() HashID {
	v := HashID{}
	j := 0
	for i := len(b) - 1; i >= 0; i-- {
		v[j] = b[i]
		j++
	}
	return v
}

func NewHexBHash(s string) HashID {
	b := HashID{}
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
	ID   HashID
}

func (m *Inventory) Read(h *NetHeader) {
	m.Type = h.ReadUInt32()
	h.ReadBytes(m.ID[:])
}

func (m *Inventory) Write(h *NetHeader) {
	h.WriteUInt32(m.Type)
	h.WriteBytes(m.ID[:])
}

type BHeader struct {
	Ver       uint32
	Prev      HashID
	Root      HashID //Merkle tree root
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
	Count     uint64
}

func (m *BHeader) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	h.ReadBytes(m.Prev[:])
	h.ReadBytes(m.Root[:])
	m.Timestamp = h.ReadUInt32()
	m.Bits = h.ReadUInt32()
	m.Nonce = h.ReadUInt32()
	m.Count, _ = h.ReadVarInt()
}

func (m *BHeader) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteBytes(m.Prev[:])
	h.WriteBytes(m.Root[:])
	h.WriteUInt32(m.Timestamp)
	h.WriteUInt32(m.Bits)
	h.WriteUInt32(m.Nonce)
	h.WriteVarInt(m.Count)
}

type TxOutPoint struct {
	Hash  HashID
	Index uint32
}

func (m *TxOutPoint) Read(h *NetHeader) {
	h.ReadBytes(m.Hash[:])
	m.Index = h.ReadUInt32()
}

func (m *TxOutPoint) Write(h *NetHeader) {
	h.WriteBytes(m.Hash[:])
	h.WriteUInt32(m.Index)
}

type TxOut struct {
	Value  uint64
	Script *script.Script
}

func (m *TxOut) Read(h *NetHeader) {
	m.Value = h.ReadUInt64()
	m.Script = h.ReadScript()
}

func (m *TxOut) Write(h *NetHeader) {
	h.WriteUInt64(m.Value)
	h.WriteScript(m.Script)
}

type TxIn struct {
	Output   TxOutPoint
	Script   *script.Script
	Sequence uint32
	Witness  *TxWitnesses
}

func (m *TxIn) Read(h *NetHeader) {
	m.Output.Read(h)
	m.Script = h.ReadScript()
	m.Sequence = h.ReadUInt32()
}

func (m *TxIn) Write(h *NetHeader) {
	m.Output.Write(h)
	h.WriteScript(m.Script)
	h.WriteUInt32(m.Sequence)
}

type TxWitnesses struct {
	Script []*script.Script
}

func (m *TxWitnesses) Read(h *NetHeader) {
	wl, _ := h.ReadVarInt()
	m.Script = make([]*script.Script, wl)
	for i, _ := range m.Script {
		v := h.ReadScript()
		m.Script[i] = v
	}
}

func (m *TxWitnesses) Write(h *NetHeader) {
	h.WriteVarInt(uint64(len(m.Script)))
	for _, v := range m.Script {
		h.WriteScript(v)
	}
}

type TX struct {
	Ver      int32
	Flag     []byte //If present, always 0001
	Ins      []*TxIn
	Outs     []*TxOut
	LockTime uint32
}

func (m *TX) IsCoinBase() bool {
	panic("Not Imp")
}

func (m *TX) HasFlag() bool {
	return len(m.Flag) == 2 && m.Flag[0] == 0 && m.Flag[1] == 1
}

func (m *TX) ReadWitnesses(h *NetHeader) {
	for i, _ := range m.Ins {
		v := &TxWitnesses{}
		v.Read(h)
		m.Ins[i].Witness = v
	}
}

func (m *TX) Read(h *NetHeader) {
	m.Ver = int32(h.ReadUInt32())
	//check flag for witnesses
	m.Flag = h.Peek(2)
	if m.HasFlag() {
		h.Skip(2)
	}
	il, _ := h.ReadVarInt()
	m.Ins = make([]*TxIn, il)
	for i, _ := range m.Ins {
		v := &TxIn{}
		v.Read(h)
		m.Ins[i] = v
	}
	ol, _ := h.ReadVarInt()
	m.Outs = make([]*TxOut, ol)
	for i, _ := range m.Outs {
		v := &TxOut{}
		v.Read(h)
		m.Outs[i] = v
	}
	//if has witnesses
	if m.HasFlag() {
		m.ReadWitnesses(h)
	}
	m.LockTime = h.ReadUInt32()
}

func (m *TX) HasWitness() bool {
	for _, v := range m.Ins {
		if v.Witness == nil {
			return false
		}
	}
	return true
}

func (m *TX) WriteWitnesses(h *NetHeader) {
	for _, v := range m.Ins {
		if v.Witness == nil {
			continue
		}
		v.Witness.Write(h)
	}
}

func (m *TX) Write(h *NetHeader) {
	h.WriteUInt32(uint32(m.Ver))
	if m.HasWitness() {
		h.WriteBytes([]byte{0, 1})
	}
	h.WriteVarInt(uint64(len(m.Ins)))
	for _, v := range m.Ins {
		v.Write(h)
	}
	h.WriteVarInt(uint64(len(m.Outs)))
	for _, v := range m.Outs {
		v.Write(h)
	}
	if m.HasWitness() {
		m.WriteWitnesses(h)
	}
	h.WriteUInt32(m.LockTime)
}

//
type MsgHeaders struct {
	Headers []*BHeader
}

func (m *MsgHeaders) Command() string {
	return NMT_HEADERS
}

func (m *MsgHeaders) Read(h *NetHeader) {
	num, _ := h.ReadVarInt()
	m.Headers = make([]*BHeader, num)
	for i, _ := range m.Headers {
		v := &BHeader{}
		v.Read(h)
		m.Headers[i] = v
	}
}

func (m *MsgHeaders) Write(h *NetHeader) {
	h.WriteVarInt(uint64(len(m.Headers)))
	for _, v := range m.Headers {
		v.Write(h)
	}
}

func NewMsgHeaders() *MsgHeaders {
	return &MsgHeaders{}
}

//
type MsgGetBlocks struct {
	Ver    uint32
	Blocks []*HashID
	Stop   *HashID
}

func (m *MsgGetBlocks) Command() string {
	return NMT_GETBLOCKS
}

func (m *MsgGetBlocks) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	num, _ := h.ReadVarInt()
	m.Blocks = make([]*HashID, num)
	for i, _ := range m.Blocks {
		v := &HashID{}
		h.ReadBytes(v[:])
		m.Blocks[i] = v
	}
	m.Stop = &HashID{}
	h.ReadBytes(m.Stop[:])
}

func (m *MsgGetBlocks) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteVarInt(uint64(len(m.Blocks)))
	for _, v := range m.Blocks {
		h.WriteBytes(v[:])
	}
	h.WriteBytes(m.Stop[:])
}

func NewMsgGetBlocks() *MsgGetBlocks {
	return &MsgGetBlocks{
		Ver:  PROTOCOL_VERSION,
		Stop: &HashID{},
	}
}

//

type MsgNotFound struct {
	Invs []*Inventory
}

func (m *MsgNotFound) Command() string {
	return NMT_NOTFOUND
}

func (m *MsgNotFound) Read(h *NetHeader) {
	size, _ := h.ReadVarInt()
	m.Invs = make([]*Inventory, size)
	for i, _ := range m.Invs {
		v := &Inventory{}
		v.Read(h)
		m.Invs[i] = v
	}
}

func (m *MsgNotFound) Write(h *NetHeader) {
	h.WriteVarInt(uint64(len(m.Invs)))
	for _, v := range m.Invs {
		v.Write(h)
	}
}

func NewMsgNotFound() *MsgNotFound {
	return &MsgNotFound{}
}

//

type MsgBlock struct {
	Ver       uint32
	Prev      HashID
	Root      HashID //Merkle tree root
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
	Txs       []*TX
}

func (m *MsgBlock) Command() string {
	return NMT_BLOCK
}

func (m *MsgBlock) HashID() HashID {
	buf := NewMsgBuffer([]byte{})
	buf.WriteUInt32(m.Ver)
	buf.WriteBytes(m.Prev[:])
	buf.WriteBytes(m.Root[:])
	buf.WriteUInt32(m.Timestamp)
	buf.WriteUInt32(m.Bits)
	buf.WriteUInt32(m.Nonce)
	hid := util.HASH256(buf.Bytes())
	id := HashID{}
	copy(id[:], hid)
	return id
}

func (m *MsgBlock) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	h.ReadBytes(m.Prev[:])
	h.ReadBytes(m.Root[:])
	m.Timestamp = h.ReadUInt32()
	m.Bits = h.ReadUInt32()
	m.Nonce = h.ReadUInt32()
	l, _ := h.ReadVarInt()
	m.Txs = make([]*TX, l)
	for i, _ := range m.Txs {
		v := &TX{}
		v.Read(h)
		m.Txs[i] = v
	}
}

func (m *MsgBlock) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteBytes(m.Prev[:])
	h.WriteBytes(m.Root[:])
	h.WriteUInt32(m.Timestamp)
	h.WriteUInt32(m.Bits)
	h.WriteUInt32(m.Nonce)
	h.WriteVarInt(uint64(len(m.Txs)))
	for _, v := range m.Txs {
		v.Write(h)
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

func (m *MsgGetData) Read(h *NetHeader) {
	num, _ := h.ReadVarInt()
	m.Invs = make([]*Inventory, num)
	for i, _ := range m.Invs {
		v := &Inventory{}
		v.Read(h)
		m.Invs[i] = v
	}
}
func (m *MsgGetData) Add(inv *Inventory) {
	m.Invs = append(m.Invs, inv)
}

func (m *MsgGetData) Write(h *NetHeader) {
	h.WriteVarInt(uint64(len(m.Invs)))
	for _, v := range m.Invs {
		v.Write(h)
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

func (m *MsgTX) Read(h *NetHeader) {
	m.Tx.Read(h)
}

func (m *MsgTX) Write(h *NetHeader) {
	m.Tx.Write(h)
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

func (m *MsgINV) Read(h *NetHeader) {
	num, _ := h.ReadVarInt()
	m.Invs = make([]*Inventory, num)
	for i, _ := range m.Invs {
		v := &Inventory{}
		v.Read(h)
		m.Invs[i] = v
	}
}

func (m *MsgINV) Write(h *NetHeader) {
	h.WriteVarInt(uint64(len(m.Invs)))
	for _, v := range m.Invs {
		v.Write(h)
	}
}

func NewMsgINV() *MsgINV {
	return &MsgINV{}
}

//

type MsgGetHeaders struct {
	Ver    uint32
	Blocks []*HashID
	Stop   *HashID
}

func (m *MsgGetHeaders) Command() string {
	return NMT_GETHEADERS
}

func (m *MsgGetHeaders) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	num, _ := h.ReadVarInt()
	m.Blocks = make([]*HashID, num)
	for i, _ := range m.Blocks {
		m.Blocks[i] = &HashID{}
		h.ReadBytes(m.Blocks[i][:])
	}
	m.Stop = &HashID{}
	h.ReadBytes(m.Stop[:])
}

func (m *MsgGetHeaders) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteVarInt(uint64(len(m.Blocks)))
	for _, v := range m.Blocks {
		h.WriteBytes(v[:])
	}
	h.WriteBytes(m.Stop[:])
}

func NewMsgGetHeaders() *MsgGetHeaders {
	return &MsgGetHeaders{}
}
