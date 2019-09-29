package util

import (
	"encoding/hex"
	"log"
	"testing"
)

func TestMakeAddress(t *testing.T) {
	s, err := hex.DecodeString("0450863AD64A87AE8A2FE83C1AF1A8403CB53F53E486D8511DAD8A04887E5B23522CD470243453A299FA9E77237716103ABC11A1DF38855ED6F2EE187E9C582BA6")
	if err != nil {
		panic(err)
	}
	addr := MakeAddress(s)
	if addr != "16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM" {
		t.Error("MakeAddress error")
	}
}

func TestAddress(t *testing.T) {
	s, err := hex.DecodeString("8fd139bb39ced713f231c58a4d07bf6954d1c201")
	if err != nil {
		panic(err)
	}
	addr := MakeAddress(s)
	log.Println(addr)
}
