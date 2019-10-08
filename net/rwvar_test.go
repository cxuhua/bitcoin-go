package net

import (
	"log"
	"testing"

	"golang.org/x/crypto/ripemd160"
)

func TestHash(t *testing.T) {
	log.Println(ripemd160.New().Sum([]byte{}))
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
