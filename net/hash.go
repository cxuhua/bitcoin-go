package net

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	HashIDWidth = 256 / 32
)

type HashID [32]byte

type UIHash [HashIDWidth]uint32

func NewUHash(bs []byte) UIHash {
	return NewBHash(bs).ToUHash()
}

func NewHexUHash(s string) UIHash {
	return NewHexBHash(s).ToUHash()
}

func (h HashID) ToUHash() UIHash {
	x := UIHash{}
	for i := 0; i < HashIDWidth; i++ {
		x[i] = ByteOrder.Uint32(h[i*4 : i*4+4])
	}
	return x
}

func (h UIHash) String() string {
	s := ""
	for i := HashIDWidth - 1; i >= 0; i-- {
		b4 := []byte{0, 0, 0, 0}
		ByteOrder.PutUint32(b4, h[i])
		s += fmt.Sprintf("%.2x%.2x%.2x%.2x", b4[3], b4[2], b4[1], b4[0])
	}
	return s
}

func (b UIHash) Low64() uint64 {
	return uint64(b[0]) | (uint64(b[1]) << 32)
}

func (b UIHash) Bits() uint {
	for pos := HashIDWidth - 1; pos >= 0; pos-- {
		if b[pos] != 0 {
			for bits := uint(31); bits > 0; bits-- {
				if b[pos]&uint32(1<<bits) != 0 {
					return uint(32*pos) + bits + 1
				}
			}
			return uint(32*pos) + 1
		}
	}
	return 0
}

// c = a * b
func (h UIHash) Mul(v UIHash) UIHash {
	a := UIHash{}
	for j := 0; j < HashIDWidth; j++ {
		carry := uint64(0)
		for i := 0; i+j < HashIDWidth; i++ {
			n := carry + uint64(a[i+j]) + uint64(h[j])*uint64(v[i])
			a[i+j] = uint32(n & 0xffffffff)
			carry = n >> 32
		}
	}
	return a
}

// /

//>>
func (b UIHash) Rshift(shift uint) UIHash {
	x := b
	for i := 0; i < HashIDWidth; i++ {
		b[i] = 0
	}
	k := int(shift / 32)
	shift = shift % 32
	for i := 0; i < HashIDWidth; i++ {
		if i-k-1 >= 0 && shift != 0 {
			b[i-k-1] |= (x[i] << (32 - shift))
		}
		if i-k >= 0 {
			b[i-k] |= (x[i] >> shift)
		}
	}
	return b
}

func (b UIHash) Lshift(shift uint) UIHash {
	x := b
	for i := 0; i < HashIDWidth; i++ {
		b[i] = 0
	}
	k := int(shift / 32)
	shift = shift % 32
	for i := 0; i < HashIDWidth; i++ {
		if i+k+1 < HashIDWidth && shift != 0 {
			b[i+k+1] |= (x[i] >> (32 - shift))
		}
		if i+k < HashIDWidth {
			b[i+k] |= (x[i] << shift)
		}
	}
	return b
}

func NewU64Hash(v uint64) UIHash {
	r := UIHash{}
	r[0] = uint32(v)
	r[1] = uint32(v >> 32)
	return r
}

//return Negative,Overflow
func (b *UIHash) SetCompact(c uint32) (bool, bool) {
	size := c >> 24
	word := c & 0x007fffff
	if size <= 3 {
		word >>= 8 * (3 - size)
		*b = NewU64Hash(uint64(word))
	} else {
		*b = NewU64Hash(uint64(word))
		*b = b.Lshift(8 * uint(size-3))
	}
	negative := word != 0 && (c&0x00800000) != 0
	overflow := word != 0 && ((size > 34) || (word > 0xff && size > 33) || (word > 0xffff && size > 32))
	return negative, overflow
}

func (b UIHash) Compact(negative bool) uint32 {
	size := (b.Bits() + 7) / 8
	compact := uint64(0)
	if size <= 3 {
		compact = b.Low64() << (8 * (3 - uint64(size)))
	} else {
		nb := b.Rshift(8 * (size - 3))
		compact = nb.Low64()
	}
	if compact&0x00800000 != 0 {
		compact >>= 8
		size++
	}
	compact |= uint64(size) << 24
	if negative && (compact&0x007fffff) != 0 {
		compact |= 0x00800000
	} else {
		compact |= 0
	}
	return uint32(compact)
}

func (h UIHash) ToHashID() HashID {
	x := HashID{}
	for i := 0; i < HashIDWidth; i++ {
		b4 := []byte{0, 0, 0, 0}
		ByteOrder.PutUint32(b4, h[i])
		copy(x[i*4+0:i*4+4], b4)
	}
	return x
}

func (h HashID) String() string {
	sv := h.Swap()
	return hex.EncodeToString(sv[:])
}

func (b HashID) IsZero() bool {
	bz := make([]byte, len(b))
	return bytes.Equal(b[:], bz)
}

func (b HashID) Equal(v HashID) bool {
	return bytes.Equal(b[:], v[:])
}

func (b HashID) Bytes() []byte {
	return b[:]
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

func NewBHash(bs []byte) HashID {
	b := HashID{}
	l := len(bs)
	if l > 32 {
		panic(SizeError)
	}
	copy(b[32-l:], bs)
	return b
}

func NewHexBHash(s string) HashID {
	if len(s) > 64 {
		panic(SizeError)
	}
	if len(s) < 64 {
		s = strings.Repeat("0", 64-len(s)) + s
	}
	v, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewBHash(v).Swap()
}
