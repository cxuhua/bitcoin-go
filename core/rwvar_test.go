package core

import (
	"bitcoin/util"
	"bytes"
	"container/list"
	"encoding/binary"
	"log"
	"testing"
)

func TestList(t *testing.T) {
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	log.Println(l)
}

func TestHash(t *testing.T) {
	r := bytes.NewBuffer(util.HexDecode("39220900"))
	v := uint32(0)
	binary.Read(r, ByteOrder, &v)
	log.Println(v)
}

func TestCoinbaseReward(t *testing.T) {
	ch := 210000 * 3
	if GetCoinbaseReward(ch) != Amount(6.25*float64(COIN)) {
		t.Errorf("error")
	}
	ch = 210000*3 - 1
	if GetCoinbaseReward(ch) != Amount(12.5*float64(COIN)) {
		t.Errorf("error")
	}
	ch = 210000 * 2
	if GetCoinbaseReward(ch) != Amount(12.5*float64(COIN)) {
		t.Errorf("error")
	}
	ch = 210000*2 - 1
	if GetCoinbaseReward(ch) != 25*COIN {
		t.Errorf("error")
	}
	ch = 210000
	if GetCoinbaseReward(ch) != Amount(25*float64(COIN)) {
		t.Errorf("error")
	}
	ch = 210000 - 1
	if GetCoinbaseReward(ch) != Amount(50*float64(COIN)) {
		t.Errorf("error")
	}
}

func TestMsgBuffer(t *testing.T) {
	w := NewMsgWriter()
	w.Write([]byte{0})
	if w.Len() != 1 {
		t.Errorf("len error")
	}
	w.Write([]byte{1, 2, 3, 4, 5})
	if w.Len() != 6 {
		t.Errorf("len error")
	}
}
