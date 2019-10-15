package util

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"log"
	"math/big"
	"testing"

	"github.com/spaolacci/murmur3"
)

func TestBloomFilter(t *testing.T) {
	seed := uint64(0xFBA4C795 + 5)
	m := murmur3.Sum32WithSeed([]byte{0}, uint32(seed))
	log.Println(m)
}

//扫描标签获取相关数据，填写上次记录hash，生成签名后提交到地址
//服务器获取数据，校验cmac，校验用户签名，填入none和time后生成签名，数据返回给用户，用户再次校验服务器签名，计算hash，连接到自己的记录上
//还需要校验计数器必须大于上一次记录的计数器
//数据广播到其他节点，节点验证签名后记录

// y^2 = x^3 -3x + b
// y = sqrt(x^3 -3x + b)
func TestP256PublicCompress(t *testing.T) {
	c := elliptic.P256().Params()
	pri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("make privake error %v", err)
	}
	log.Println("key = ", hex.EncodeToString(pri.D.Bytes()))
	log.Println("x=", hex.EncodeToString(pri.X.Bytes()))
	log.Println("y=", hex.EncodeToString(pri.Y.Bytes()))

	d := pri.Y.Bit(0)
	x := pri.X
	var y, x3b, x3 big.Int
	x3.SetInt64(3)
	x3.Mul(&x3, x)
	x3b.Mul(x, x)
	x3b.Mul(&x3b, x)
	x3b.Add(&x3b, c.B)
	x3b.Sub(&x3b, &x3)
	x3b.Mod(&x3b, c.P)
	y.ModSqrt(&x3b, c.P)
	if y.Bit(0) != d {
		y.Sub(c.P, &y)
	}
	if y.Cmp(pri.Y) != 0 {
		t.Errorf("failed")
	}
	log.Println("cy=", hex.EncodeToString(y.Bytes()), "ybit=", d)
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
	v, h, err := DecodeAddr(addr)
	if err != nil {
		t.Errorf("deocde addr error %v", err)
	}
	if v != 0 || len(h) != 20 {
		t.Errorf("return error")
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
	v, h, err := DecodeAddr(addr)
	if err != nil {
		t.Errorf("deocde addr error %v", err)
	}
	if v != 0 || len(h) != 20 {
		t.Errorf("return error")
	}
	if !bytes.Equal(h, s) {
		t.Errorf("not equal public hash160")
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
	v, h, err := DecodeAddr(addr)
	if err != nil {
		t.Errorf("deocde addr error %v", err)
	}
	if v != 0 || len(h) != 20 {
		t.Errorf("return error")
	}
}
func TestBECH32Address(t *testing.T) {
	s, err := hex.DecodeString("751e76e8199196d454941c45d1b3a323f1433bd6")
	if err != nil {
		panic(err)
	}
	addr := BECH32Address(s)
	if addr != "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4" {
		t.Errorf("TestAddress error %s", addr)
	}
	v, h, err := DecodeAddr(addr)
	if err != nil {
		t.Errorf("deocde addr error %v", err)
	}
	if v != 0 || len(h) != 20 {
		t.Errorf("return error")
	}
}

func TestP2SHAddress(t *testing.T) {
	data := HexDecode("52_21_0293baf0397588acc1aba056e868fd188dc0eea7554b45370aae862f9d2493a4c1_21_020ab7517cf22a46b503ee8dcae7f9f109ec4cd19f0ab9d77c89c607554f3d5aa9_52_ae")
	addr := P2SHAddress(data)
	if addr != "3Ae2TYfyHvwH11pUy6HaK7rBYn9GfGZ3Fk" {
		t.Errorf("P2SHAddress error %s", addr)
	}
	v, h, err := DecodeAddr(addr)
	if err != nil {
		t.Errorf("deocde addr error %v", err)
	}
	if v != 5 || len(h) != 20 {
		t.Errorf("return error")
	}
}
