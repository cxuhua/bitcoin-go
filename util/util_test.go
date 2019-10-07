package util

import (
	"encoding/hex"
	"log"
	"testing"
)

func TestDHash(t *testing.T) {
	//0020
	//a16b5755f7f6f96dbd65f5f0d6ab9418b89af4b1f14a1bb8a09062c35f0dcb54
	d, _ := hex.DecodeString("701a8d401c84fb13e6baf169d59684e17abd9fa216c8cc5b9fc63d622ff8c58d")

	log.Println(BECH32Address(d))

}

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

func TestLongAddress(t *testing.T) {
	s, err := hex.DecodeString("04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f")
	if err != nil {
		panic(err)
	}
	addr := P2PKHAddress(s)
	if addr != "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa" {
		t.Errorf("TestLongAddress error %s", addr)
	}
}
func TestBECH32Address(t *testing.T) {
	s, err := hex.DecodeString("0279BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798")
	if err != nil {
		panic(err)
	}
	addr := BECH32Address(s)
	if addr != "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4" {
		t.Errorf("TestAddress error %s", addr)
	}
}

func TestP2SHAddress(t *testing.T) {
	data := HexDecode("52_21_0293baf0397588acc1aba056e868fd188dc0eea7554b45370aae862f9d2493a4c1_21_020ab7517cf22a46b503ee8dcae7f9f109ec4cd19f0ab9d77c89c607554f3d5aa9_52_ae")
	addr := P2SHAddress(data)
	if addr != "3Ae2TYfyHvwH11pUy6HaK7rBYn9GfGZ3Fk" {
		t.Errorf("P2SHAddress error %s", addr)
	}
}
