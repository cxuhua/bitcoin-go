package util

import (
	"encoding/hex"
	"log"
	"strings"
	"testing"
)

func TestMakePublicToAddress(t *testing.T) {
	s, err := hex.DecodeString("0450863AD64A87AE8A2FE83C1AF1A8403CB53F53E486D8511DAD8A04887E5B23522CD470243453A299FA9E77237716103ABC11A1DF38855ED6F2EE187E9C582BA6")
	if err != nil {
		panic(err)
	}
	addr := P2PKHAddress(s)
	if addr != "16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM" {
		t.Error("MakeAddress error")
	}
}

func TestMakePKHAddress(t *testing.T) {
	s, err := hex.DecodeString("a896db19ae4746d8862fcdd7cb886ca5765296e8")
	if err != nil {
		panic(err)
	}
	addr := P2PKHAddress(s)
	if addr != "1GNREsqR6D3Sfo2CVScS1SDFBuzLJGs8WQ" {
		t.Error("MakeAddress error")
	}
}

func TestAddress(t *testing.T) {
	s, err := hex.DecodeString("8fd139bb39ced713f231c58a4d07bf6954d1c201")
	if err != nil {
		panic(err)
	}
	addr := P2PKHAddress(s)
	log.Println(addr)
}

func TestP2SHAddress(t *testing.T) {
	s, err := hex.DecodeString("0048304502210080075aa29c42f8062f75cf6ab32004944417af974775581719008052c78719710220409fee54c6ddf2ca83e090077e443f95b427a63cc1ad87fca2625951b789d1c201493046022100b61d8f206d17efd6db32dad106f754f231ee8a16882929b1eb39a58bfd36b39e022100c62cff92dd6fb22b373025fc9b87044cf1b33502acc9de707e5f54d1c8a042a7014752210293baf0397588acc1aba056e868fd188dc0eea7554b45370aae862f9d2493a4c121020ab7517cf22a46b503ee8dcae7f9f109ec4cd19f0ab9d77c89c607554f3d5aa952ae")
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(s); i++ {
		addr := P2SHAddress(s[i:])
		if strings.HasPrefix(addr, "3Ae2") {
			log.Println(addr, hex.EncodeToString(s[i:]))
		}
	}

	//if addr != "3Q99PXASzaXweaLWD4x3bFn49KK9kGoTgK" {
	//	t.Errorf("P2SHAddress error %s", addr)
	//}
}
