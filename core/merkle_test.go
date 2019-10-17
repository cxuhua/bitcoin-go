package core

import (
	"bytes"
	"encoding/binary"
	"log"
	"testing"

	"github.com/dchest/siphash"

	"github.com/willf/bitset"
)

var (
	amap = map[HashID]int{}
)

func TestByteMap(t *testing.T) {
	id1 := HashID{0}
	log.Println(id1)
	amap[id1] = 1

	id2 := HashID{1}
	amap[id2] = 2

	log.Println(amap[HashID{0}], amap[HashID{1}])
}

//data from bitcoin
func TestSipHash(t *testing.T) {
	k0 := uint64(0x0706050403020100)
	k1 := uint64(0x0F0E0D0C0B0A0908)
	hv := NewHashID("1f1e1d1c1b1a191817161514131211100f0e0d0c0b0a09080706050403020100")
	v := siphash.Hash(k0, k1, hv.Bytes())
	if v != 0x7127512f72f27cce {
		t.Errorf("test siphash error")
	}
	key := make([]byte, 16)
	k0 = uint64(0x0706050403020100)
	k1 = uint64(0x0F0E0D0C0B0A0908)
	binary.LittleEndian.PutUint64(key[:8], k0)
	binary.LittleEndian.PutUint64(key[8:], k1)
	h := siphash.New(key)
	if h.Sum64() != 0x726fdb47dd0e0e31 {
		t.Errorf("test hash error")
	}
	h.Write([]byte{0})
	if h.Sum64() != 0x74f839c593dc67fd {
		t.Errorf("write 1 error")
	}
	h.Write([]byte{1, 2, 3, 4, 5, 6, 7})
	if h.Sum64() != 0x93f5f5799a932462 {
		t.Errorf("write 2 error")
	}
	b8 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b8, 0x0F0E0D0C0B0A0908)
	h.Write(b8)
	if h.Sum64() != 0x3f2acc7f57c29bdb {
		t.Errorf("write 3 error")
	}
}

func TestNewBitSet(t *testing.T) {
	d := []byte{1, 2, 3, 4, 5}
	bs := NewBitSet(d)
	v := FromBitSet(bs)
	if !bytes.Equal(d, v) {
		t.Errorf("test newbitset failed")
	}
}

func TestMerkleArray(t *testing.T) {
	a := []HashID{}
	for i := 0; i < 21; i++ {
		tmp := HashID{byte(i)}
		a = append(a, tmp)
	}
	bs := bitset.New(uint(len(a)))
	bs.Set(20)

	tree := NewMerkleTree(len(a))
	tree.Build(a, bs)

	nt := GetMerkleTree(tree.Trans(), tree.Hashs(), tree.Bits())
	_, _, c1 := nt.Extract()
	if len(c1) != 1 || c1[0] != 20 {
		t.Errorf("test extrace error")
	}
}
