package net

import (
	"bitcoin/config"
	"log"
	"testing"
	"time"
)

func TestU32HashMul(t *testing.T) {
	x1 := NewHexUHash("7D1DE5EAF9B156D53208F033B5AA8122D2d2355d5e12292b121156cfdb4a529c")
	x2 := NewHexUHash("7D1DE5EAF9B156D53208F033B5AA8122D2d2355d5e12292b121156cfdb4a529c")
	x := x1.Mul(x2)
	if x.String() != "62a38c0486f01e45879d7910a7761bf30d5237e9873f9bff3642a732c4d84f10" {
		t.Errorf("mul error")
	}
	//
	x1 = NewHexUHash("7D1DE5EAF9B156D53208F033B5AA8122D2d2355d5e12292b121156cfdb4a529c")
	x2 = NewUHash([]byte("\x70\x32\x1d\x7c\x47\xa5\x6b\x40\x26\x7e\x0a\xc3\xa6\x9c\xb6\xbf\x13\x30\x47\xa3\x19\x2d\xda\x71\x49\x13\x72\xf0\xb4\xca\x81\xd7"))
	x = x1.Mul(x2)
	if x.String() != "de37805e9986996cfba76ff6ba51c008df851987d9dd323f0e5de07760529c40" {
		t.Errorf("mul error")
	}
}

func TestSetCompact(t *testing.T) {

	v := NewU64Hash(0)
	n, o := v.SetCompact(0x01123456)
	if v.String() != "0000000000000000000000000000000000000000000000000000000000000012" {
		t.Errorf("test set compact 1 failed")
	}
	if n != false || o != false {
		t.Errorf("test set compact 2 failed")
	}
	if v.Compact(n) != 0x01120000 {
		t.Errorf("test compact 1 failed")
	}

	v = NewU64Hash(0x80)
	if v.Compact(false) != 0x02008000 {
		t.Errorf("test compact 2 failed")
	}

	n, o = v.SetCompact(0x01fedcba)
	if v.String() != "000000000000000000000000000000000000000000000000000000000000007e" {
		t.Errorf("test set compact 3 failed")
	}
	if n != true || o != false {
		t.Errorf("test set compact 4 failed")
	}

	if v.Compact(n) != 0x01fe0000 {
		t.Errorf("test compact 3 failed")
	}

	n, o = v.SetCompact(0xff123456)
	if n != false || o != true {
		t.Errorf("test set compact 5 failed")
	}

	n, o = v.SetCompact(0x20123456)
	if n != false || o != false {
		t.Errorf("test set compact 6 failed")
	}
	if v.String() != "1234560000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("test set compact 7 failed")
	}

	if v.Compact(n) != 0x20123456 {
		t.Errorf("test compact 4 failed")
	}
}

//(203616-2016) -> 2012-10-03 09:11:00
//(203616-1) -> 2012-10-17 07:14:55
//(203616-1) Bits = 0x1a057e08
//result:1a0575ef
func TestU32HashCompact2(t *testing.T) {
	conf := config.GetConfig()
	limit := NewHexUHash(conf.PowLimit)
	span := int64(14 * 24 * 60 * 60)
	t1, _ := time.Parse("2006-01-02 15:04:05", "2012-10-03 09:11:00")
	t2, _ := time.Parse("2006-01-02 15:04:05", "2012-10-17 07:14:55")
	sub := t2.Sub(t1)
	subv := int64(sub.Seconds())
	if subv < span/4 {
		subv = span / 4
	}
	if subv > span*4 {
		subv = span * 4
	}
	n := UIHash{}
	log.Println(n.SetCompact(0x1a057e08))
	log.Println(n)
	n = n.MulUInt32(uint32(subv))
	n = n.Div(NewU64Hash(uint64(span)))
	if n.Cmp(limit) > 0 {
		n = limit
	}
	if n.Compact(false) != 0x1a0575ef {
		t.Errorf("error")
	}
}

func TestHashEqual(t *testing.T) {
	v1 := NewHexBHash("0101000000000000000000000000000000000000000000000000000000001234")
	v2 := v1.ToUHash().ToHashID()
	if !v1.Equal(v2) {
		t.Errorf("test Equal failed")
	}
}

func TestU32HashShift(t *testing.T) {
	s := "0100000000000000000000000000000000000000000000000000000000000000"
	one := NewHexUHash(s)

	for i := uint(0); i < 254; i++ {
		one = one.Lshift(1)
	}
	for i := uint(0); i < 254; i++ {
		one = one.Rshift(1)
	}
	if one.String() != s {
		t.Errorf("test shift error")
	}
}

func TestU32HashBits(t *testing.T) {
	s := "0100000000000000000000000000000000000000000000000000000000000000"
	v1 := NewHexUHash(s)
	if v1.String() != s {
		t.Errorf("string error")
	}
	if v1.Bits() != 1 {
		t.Errorf("bits error")
	}
	s = "0000000000000000000000000000000000000000000000000000000000000080"
	v1 = NewHexUHash(s)
	log.Println(v1.Bits())
	if v1.String() != s {
		t.Errorf("string error")
	}
	if v1.Bits() != 256 {
		t.Errorf("bits error")
	}
}
