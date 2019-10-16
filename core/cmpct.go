package core

import "bitcoin/util"

type MsgBlockTxn struct {
	Hash HashID
	Txs  []TX
}

func (m *MsgBlockTxn) Command() string {
	return NMT_BLOCKTXN
}

func (m *MsgBlockTxn) Read(h *NetHeader) {
	m.Hash = h.ReadHash()
	ic, _ := h.ReadVarInt()
	m.Txs = make([]TX, ic)
	for i, _ := range m.Txs {
		m.Txs[i].Read(h)
	}
}

func (m *MsgBlockTxn) Write(h *NetHeader) {
	h.WriteHash(m.Hash)
	h.WriteVarInt(len(m.Txs))
	for _, v := range m.Txs {
		v.Write(h)
	}
}

func NewMsgBlockTxn() *MsgBlockTxn {
	return &MsgBlockTxn{}
}

//

type MsgGetBlockTxn struct {
	Hash   HashID
	Indexs []uint32
}

func (m *MsgGetBlockTxn) Command() string {
	return NMT_GETBLOCKTXN
}

func (m *MsgGetBlockTxn) Read(h *NetHeader) {
	m.Hash = h.ReadHash()
	ic, _ := h.ReadVarInt()
	m.Indexs = make([]uint32, ic)
	for i, _ := range m.Indexs {
		v, _ := h.ReadVarInt()
		m.Indexs[i] = uint32(v)
	}
}

func (m *MsgGetBlockTxn) Write(h *NetHeader) {
	h.WriteHash(m.Hash)
	h.WriteVarInt(len(m.Indexs))
	for _, v := range m.Indexs {
		h.WriteVarInt(v)
	}
}

func NewMsgGetBlockTxn() *MsgGetBlockTxn {
	return &MsgGetBlockTxn{}
}

//

type PreFilledTx struct {
	Index uint32
	Tx    TX
}

func (m *PreFilledTx) Read(h *NetHeader) {
	idx, _ := h.ReadVarInt()
	m.Index = uint32(idx)
	m.Tx.Read(h)
}

func (m *PreFilledTx) Write(h *NetHeader) {
	h.WriteVarInt(m.Index)
	m.Tx.Write(h)
}

type CmpctHeader struct {
	Ver       uint32
	Prev      HashID
	Merkle    HashID
	Timestamp uint32
	Bits      uint32
	Nonce     uint32
}

func (m *CmpctHeader) Read(h *NetHeader) {
	m.Ver = h.ReadUInt32()
	h.ReadBytes(m.Prev[:])
	h.ReadBytes(m.Merkle[:])
	m.Timestamp = h.ReadUInt32()
	m.Bits = h.ReadUInt32()
	m.Nonce = h.ReadUInt32()
}

func (m *CmpctHeader) Write(h *NetHeader) {
	h.WriteUInt32(m.Ver)
	h.WriteBytes(m.Prev[:])
	h.WriteBytes(m.Merkle[:])
	h.WriteUInt32(m.Timestamp)
	h.WriteUInt32(m.Bits)
	h.WriteUInt32(m.Nonce)
}

type MsgCmpctBlock struct {
	Header   CmpctHeader
	Nonce    uint64
	ShortIds []uint64 //6 bytes int
	PreTxs   []PreFilledTx
	k1       uint64 //siphash k1,use FillSelector set
	k2       uint64 //k2
}

func (m *MsgCmpctBlock) Command() string {
	return NMT_CMPCTBLOCK
}

func (m *MsgCmpctBlock) FillSelector(header CmpctHeader, nonce uint64) {
	h := NewNetHeader()
	header.Write(h)
	h.WriteUInt64(nonce)
	sv := util.SHA256(h.Payload)
	hv := NewHashID(sv)
	m.k1 = hv.GetUint64(0)
	m.k2 = hv.GetUint64(1)
}

func (m *MsgCmpctBlock) GetShortId(hv HashID) uint64 {
	v := SipHash(m.k1, m.k2, hv)
	return v & 0xffffffffffff
}

func (m *MsgCmpctBlock) Read(h *NetHeader) {
	m.Header.Read(h)
	m.Nonce = h.ReadUInt64()
	sc, _ := h.ReadVarInt()
	m.ShortIds = make([]uint64, sc)
	for i, _ := range m.ShortIds {
		m.ShortIds[i] = h.ReadShortId()
	}
	tc, _ := h.ReadVarInt()
	m.PreTxs = make([]PreFilledTx, tc)
	for i, _ := range m.PreTxs {
		m.PreTxs[i].Read(h)
	}
	m.FillSelector(m.Header, m.Nonce)
}

func (m *MsgCmpctBlock) Write(h *NetHeader) {
	m.Header.Write(h)
	h.WriteUInt64(m.Nonce)
	h.WriteVarInt(len(m.ShortIds))
	for _, v := range m.ShortIds {
		h.WriteShortId(v)
	}
	h.WriteVarInt(len(m.PreTxs))
	for _, v := range m.PreTxs {
		v.Write(h)
	}
}

func NewMsgCmpctBlock() *MsgCmpctBlock {
	return &MsgCmpctBlock{}
}

///

type MsgSendCmpct struct {
	Inter uint8
	Ver   uint64
}

func (m *MsgSendCmpct) Command() string {
	return NMT_SENDCMPCT
}

func (m *MsgSendCmpct) Read(h *NetHeader) {
	m.Inter = h.ReadUint8()
	m.Ver = h.ReadUInt64()
}

func (m *MsgSendCmpct) Write(h *NetHeader) {
	h.WriteUint8(m.Inter)
	h.WriteUInt64(m.Ver)
}

func NewMsgSendCmpct() *MsgSendCmpct {
	return &MsgSendCmpct{}
}
