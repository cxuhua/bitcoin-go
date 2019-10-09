package net

import (
	"bitcoin/util"
	"bytes"
	"encoding/binary"
	"log"
	"testing"
)

func TestHash(t *testing.T) {
	r := bytes.NewBuffer(util.HexDecode("39220900"))
	v := uint32(0)
	binary.Read(r, ByteOrder, &v)
	log.Println(v)
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
