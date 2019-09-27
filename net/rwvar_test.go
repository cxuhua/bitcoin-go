package net

import (
	"bytes"
	"io/ioutil"
	"log"
	"testing"

	"golang.org/x/crypto/ripemd160"
)

func TestHash(t *testing.T) {
	log.Println(ripemd160.New().Sum([]byte{}))
}

func TestRWVarInt(t *testing.T) {
	buf := &bytes.Buffer{}
	vs := []uint64{254, 0xFD, 0xFFFF, 0xFFFF1, 0xFFFFFFFF, 0xFFFFFFFF1, 0xFFFFFFFFFF1}
	for _, v1 := range vs {
		buf.Reset()
		l1 := WriteVarInt(buf, v1)
		v2, l2 := ReadVarInt(buf)
		if l1 != l2 || v1 != v2 {
			t.Fatalf("test v1=%X v2=%X error,l1=%d,l2=%d", v1, v2, l1, l2)
		}
	}
}

//tx.dat
func TestMsgTX(t *testing.T) {
	data, err := ioutil.ReadFile("tx.dat")
	if err != nil {
		panic(err)
	}
	h := &MessageHeader{}
	pr := bytes.NewReader(data)
	m := NewMsgTX()
	m.Read(h, pr)
	for _, v := range m.Tx.Ins {
		log.Println(v.Script.HasValidOps())
	}

}
