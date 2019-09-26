package net

import (
	"bytes"
	"io/ioutil"
	"log"
	"testing"
)

func TestRWVarInt(t *testing.T) {
	buf := &bytes.Buffer{}
	vs := []uint64{254, 0xFD, 0xFFFF, 0xFFFF1, 0xFFFFFFFF, 0xFFFFFFFF1, 0xFFFFFFFFFF1}
	for _, v1 := range vs {
		buf.Reset()
		l := WriteVarInt(buf, v1)
		v2 := ReadVarInt(buf)
		if v1 != v2 {
			t.Fatalf("test v1=%X v2=%X error,l=%d", v1, v2, l)
		}
	}
}

//tx.dat
func TestMsgTX(t *testing.T) {
	data, err := ioutil.ReadFile("../dat/tx.dat")
	if err != nil {
		panic(err)
	}
	h := &MessageHeader{}
	pr := bytes.NewReader(data)
	m := NewMsgTX()
	m.Read(h, pr)
	log.Println(m)
}
