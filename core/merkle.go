package core

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/willf/bitset"
)

type MerkleNode []byte

type MerkleTree struct {
	trans int
	vhash []HashID
	bits  []bool
	bad   bool
}

func NewMerkleTree(num int) *MerkleTree {
	v := &MerkleTree{}
	v.trans = num
	v.vhash = []HashID{}
	v.bits = []bool{}
	v.bad = false
	return v
}

func GetMerkleTree(num int, hashs []HashID, bits *bitset.BitSet) *MerkleTree {
	v := &MerkleTree{}
	v.trans = num
	v.vhash = hashs
	v.bits = []bool{}
	for i := uint(0); i < bits.Len(); i++ {
		v.bits = append(v.bits, bits.Test(i))
	}
	v.bad = false
	return v
}

func (tree *MerkleTree) Trans() int {
	return tree.trans
}

func (tree *MerkleTree) Hashs() []HashID {
	return tree.vhash
}

func (tree *MerkleTree) Bits() *bitset.BitSet {
	ret := bitset.New(uint(len(tree.bits)))
	for i, v := range tree.bits {
		ret.SetTo(uint(i), v)
	}
	return ret
}

func (tree *MerkleTree) Hash(n1 HashID, n2 HashID) HashID {
	ret := HashID{}
	v := append([]byte{}, n1[:]...)
	v = append(v, n2[:]...)
	return HASH256To(v, &ret)
}

func (tree *MerkleTree) Height() int {
	h := 0
	for tree.TreeWidth(h) > 1 {
		h++
	}
	return h
}

func BuildMerkleTree(ids []HashID) *MerkleTree {
	num := len(ids)
	tree := NewMerkleTree(num)
	vb := bitset.New(uint(num))
	h := tree.Height()
	tree.build(h, 0, ids, vb)
	return tree
}

func (tree *MerkleTree) Build(ids []HashID, vb *bitset.BitSet) *MerkleTree {
	tree.bad = false
	tree.vhash = []HashID{}
	tree.bits = []bool{}
	h := tree.Height()
	tree.build(h, 0, ids, vb)
	return tree
}

func (tree *MerkleTree) build(h int, pos int, ids []HashID, vb *bitset.BitSet) {
	match := false
	for p := pos << h; p < (pos+1)<<h && p < tree.trans; p++ {
		if vb.Test(uint(p)) {
			match = true
		}
	}
	tree.bits = append(tree.bits, match)
	if h == 0 || !match {
		tree.vhash = append(tree.vhash, tree.CalcHash(h, pos, ids))
	} else {
		tree.build(h-1, pos*2, ids, vb)
		if pos*2+1 < tree.TreeWidth(h-1) {
			tree.build(h-1, pos*2+1, ids, vb)
		}
	}
}

func (tree *MerkleTree) Extract() (HashID, []HashID, []int) {
	ids := make([]HashID, 0)
	idx := make([]int, 0)
	tree.bad = false
	if tree.trans == 0 {
		return HashID{}, nil, nil
	}
	if uint(tree.trans) > MAX_BLOCK_WEIGHT/MIN_TRANSACTION_WEIGHT {
		return HashID{}, nil, nil
	}
	if len(tree.vhash) > tree.trans {
		return HashID{}, nil, nil
	}
	if len(tree.bits) < len(tree.vhash) {
		return HashID{}, nil, nil
	}
	h := tree.Height()
	nbits, nhash := 0, 0
	root := tree.extract(h, 0, &nbits, &nhash, &ids, &idx)
	if tree.bad {
		return HashID{}, nil, nil
	}
	if (nbits+7)/8 != (len(tree.bits)+7)/8 {
		return HashID{}, nil, nil
	}
	if nhash != len(tree.vhash) {
		return HashID{}, nil, nil
	}
	return root, ids, idx
}

func (tree *MerkleTree) extract(h int, pos int, nbits *int, nhash *int, ids *[]HashID, idx *[]int) HashID {
	if *nbits >= len(tree.bits) {
		tree.bad = true
		return HashID{}
	}
	match := tree.bits[*nbits]
	*nbits++
	if h == 0 || !match {
		if *nhash >= len(tree.vhash) {
			tree.bad = true
			return HashID{}
		}
		hash := tree.vhash[*nhash]
		*nhash++
		if h == 0 && match {
			*ids = append(*ids, hash)
			*idx = append(*idx, pos)
		}
		return hash
	} else {
		left, right := tree.extract(h-1, pos*2, nbits, nhash, ids, idx), HashID{}
		if pos*2+1 < tree.TreeWidth(h-1) {
			right = tree.extract(h-1, pos*2+1, nbits, nhash, ids, idx)
			if left.Equal(right) {
				tree.bad = true
			}
		} else {
			right = left
		}
		return tree.Hash(left, right)
	}
}

func (tree *MerkleTree) TreeWidth(h int) int {
	return (tree.trans + (1 << h) - 1) >> h
}

func (tree *MerkleTree) CalcHash(h int, pos int, ids []HashID) HashID {
	if len(ids) == 0 {
		panic(errors.New("empty merkle array"))
	}
	if h == 0 {
		return ids[pos]
	}
	left, right := tree.CalcHash(h-1, pos*2, ids), HashID{}
	if pos*2+1 < tree.TreeWidth(h-1) {
		right = tree.CalcHash(h-1, pos*2+1, ids)
	} else {
		right = left
	}
	return tree.Hash(left, right)
}

func init() {
	//set bitset endian
	bitset.LittleEndian()
}

func NewBitSet(d []byte) *bitset.BitSet {
	bl := uint(len(d) * 8)
	bits := bitset.New(bl)
	buf := &bytes.Buffer{}
	binary.Write(buf, ByteOrder, uint64(bl))
	nl := ((len(d) + 7) / 8) * 8
	nb := make([]byte, nl)
	copy(nb, d)
	binary.Write(buf, ByteOrder, nb)
	bits.ReadFrom(buf)
	return bits
}

func FromBitSet(bs *bitset.BitSet) []byte {
	buf := &bytes.Buffer{}
	bs.WriteTo(buf)
	bl := uint64(0)
	binary.Read(buf, ByteOrder, &bl)
	bl = (bl + 7) / 8
	bb := make([]byte, bl)
	binary.Read(buf, ByteOrder, bb)
	return bb
}

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

func (m *MsgMerkleBlock) Extract() (HashID, []HashID, []int) {
	tree := GetMerkleTree(int(m.Total), m.Hashs, NewBitSet(m.Flags))
	return tree.Extract()
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
