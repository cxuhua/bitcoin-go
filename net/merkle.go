package net

type MsgMerkleBlock struct {
	Version    int32
	PrevBlock  HashID
	MerkleRoot HashID
	Timestamp  uint32
	Bits       uint32
	Nonce      uint32
	Total      uint32
	Hashs      []HashID
	Flags      []byte
}

func (m *MsgMerkleBlock) Command() string {
	return NMT_MERKLEBLOCK
}

func (m *MsgMerkleBlock) Read(h *NetHeader) {
	m.Version = h.ReadInt32()
	m.PrevBlock = h.ReadHash()
	m.MerkleRoot = h.ReadHash()
	m.Timestamp = h.ReadUInt32()
	m.Bits = h.ReadUInt32()
	m.Nonce = h.ReadUInt32()
	m.Total = h.ReadUInt32()
	hc, _ := h.ReadVarInt()
	m.Hashs = make([]HashID, hc)
	for i, _ := range m.Hashs {
		m.Hashs[i] = h.ReadHash()
	}
	fc, _ := h.ReadVarInt()
	m.Flags = make([]byte, fc)
	h.ReadBytes(m.Flags)
}

func (m *MsgMerkleBlock) Write(h *NetHeader) {
	h.WriteInt32(m.Version)
	h.WriteHash(m.PrevBlock)
	h.WriteHash(m.MerkleRoot)
	h.WriteUInt32(m.Timestamp)
	h.WriteUInt32(m.Bits)
	h.WriteUInt32(m.Nonce)
	h.WriteUInt32(m.Total)
	h.WriteVarInt(len(m.Hashs))
	for _, v := range m.Hashs {
		h.WriteHash(v)
	}
	h.WriteVarInt(len(m.Flags))
	h.WriteBytes(m.Flags)
}

func NewMsgMerkleBlock() *MsgMerkleBlock {
	return &MsgMerkleBlock{}
}
