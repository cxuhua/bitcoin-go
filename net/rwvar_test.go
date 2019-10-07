package net

import (
	"bitcoin/script"
	"encoding/hex"
	"log"
	"testing"

	"golang.org/x/crypto/ripemd160"
)

func TestHash(t *testing.T) {
	log.Println(ripemd160.New().Sum([]byte{}))
}

func TestScript(t *testing.T) {
	s := script.NewScriptHex("76a9148fd139bb39ced713f231c58a4d07bf6954d1c20188ac")
	log.Printf(hex.EncodeToString(*s))
	stack := script.NewStack()
	s.Eval(stack, nil, 0, script.SIG_VER_BASE)
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
