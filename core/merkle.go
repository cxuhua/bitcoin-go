package core

import (
	"bitcoin/util"
	"errors"
)

type MerkleNode []byte

type MerkleArray []MerkleNode

func (a MerkleArray) Hash(n1 MerkleNode, n2 MerkleNode) []byte {
	v := append([]byte{}, n1...)
	v = append(v, n2...)
	return util.HASH256(v)
}

func (a MerkleArray) TreeWidth(h int) int {
	return (len(a) + (1 << h) - 1) >> h
}

func (a MerkleArray) CalcHash(h int, pos int) []byte {
	if len(a) == 0 {
		panic(errors.New("empty merkle array"))
	}
	if h == 0 {
		return a[pos]
	}
	left, right := a.CalcHash(h-1, pos*2), []byte{}
	if pos*2+1 < a.TreeWidth(h-1) {
		right = a.CalcHash(h-1, pos*2+1)
	} else {
		right = left
	}
	return a.Hash(left, right)
}

//func (a MerkleNodeArray) Hash() []byte {
//	if len(a) == 1 {
//		return a[0]
//	}
//	v0 := a[0]
//	for i := 1; i < len(a); i++ {
//		v1 := a[i]
//		v := append([]byte{}, v0...)
//		v = append(v, v1...)
//		v0 = util.HASH256(v)
//	}
//	return v0
//}

//type MerkleTree []MerkleNodeArray
//
//func (mt MerkleTree) Root() MerkleNode {
//	l := len(mt)
//	return mt[l-1][0]
//}
//
//func (mt MerkleTree) at(i, j int) MerkleNode {
//	if i < 0 && i >= len(mt) {
//		panic(errors.New("i out bound"))
//	}
//	fs := mt[i]
//	if j < 0 && j >= len(fs) {
//		panic(errors.New("j out bound"))
//	}
//	return mt[i][j]
//}
//
//func (mt MerkleTree) findPos(node MerkleNode) (int, int) {
//	for i := 0; i < len(mt); i++ {
//		fs := mt[i]
//		for j := 0; j < len(fs); j++ {
//			if bytes.Equal(node, fs[j]) {
//				return i, j
//			}
//		}
//	}
//	return -1, -1
//}
//
//func (mt MerkleTree) FindPath(node MerkleNode) MerkleNodeArray {
//	vs := MerkleNodeArray{}
//	i, j := mt.findPos(node)
//	if i < 0 || j < 0 {
//		return vs
//	}
//	vs = append(vs, mt.at(i, j))
//	if i == len(mt)-1 && j == 0 {
//		return vs
//	}
//	x := j
//	for y := 0; y < len(mt)-1; y++ {
//		if x%2 == 0 {
//			vs = append(vs, mt.at(y, x+1))
//		} else {
//			vs = append(vs, mt.at(y, x-1))
//		}
//		x = x / 2
//	}
//	return vs
//}
//
//func ComputeMerkleTree(tree *MerkleTree, hashs MerkleNodeArray) *MerkleTree {
//	fvs := MerkleNodeArray{}
//	if len(hashs) == 1 {
//		fvs = append(fvs, hashs[0])
//		*tree = append(*tree, fvs)
//		return tree
//	}
//	if len(hashs)%2 != 0 {
//		hashs = append(hashs, hashs[len(hashs)-1])
//	}
//	hvs := MerkleNodeArray{}
//	for i := 0; i < len(hashs); i++ {
//		fvs = append(fvs, hashs[i])
//		if i%2 == 0 {
//			continue
//		}
//		v1 := hashs[i-1]
//		v2 := hashs[i-0]
//		v := append([]byte{}, v1...)
//		v = append(v, v2...)
//		hvs = append(hvs, util.HASH256(v))
//	}
//	*tree = append(*tree, fvs)
//	return ComputeMerkleTree(tree, hvs)
//}
//
//func NewMerkleTree(hashs MerkleNodeArray) *MerkleTree {
//	return ComputeMerkleTree(&MerkleTree{}, hashs)
//}

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
